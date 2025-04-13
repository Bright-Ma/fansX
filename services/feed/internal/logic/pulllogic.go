package logic

import (
	"bilibili/common/util"
	"bilibili/services/content/public/proto/publicContentRpc"
	interlua "bilibili/services/feed/internal/lua"
	"bilibili/services/relation/proto/relationRpc"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"bilibili/services/feed/internal/svc"
	"bilibili/services/feed/proto/feedRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PullLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPullLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PullLogic {
	return &PullLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PullLogic) Pull(in *feedRpc.PullReq) (*feedRpc.PullResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	cache := l.svcCtx.Cache
	relationClient := l.svcCtx.RelationClient
	contentClient := l.svcCtx.ContentClient
	executor := l.svcCtx.Executor

	logger.Info("user pull feed stream", "userId", in.UserId, "size", in.Limit, "timeStamp", in.TimeStamp)

	resp, err := relationClient.ListFollowing(timeout, &relationRpc.ListFollowingReq{
		UserId: in.UserId,
		All:    true,
	})
	if err != nil {
		logger.Error("list following:" + err.Error())
		return nil, err
	}

	list := make([][]int64, 0)

	inbox := "inbox:" + strconv.FormatInt(in.UserId, 10)
	inter, err := executor.Execute(timeout, interlua.GetRevByScoreScript(), []string{inbox}, 0, in.TimeStamp).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Error("get inbox:" + err.Error())
		return nil, err
	} else if errors.Is(err, redis.Nil) {
		logger.Warn("not find inbox:" + err.Error())
	}

	if errors.Is(err, redis.Nil) || len(inter.([]interface{})) == 0 {
		logger.Info("search back")
		err = searchBack(timeout, logger, in.Limit, resp.UserId, in.TimeStamp, list, contentClient)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Info("find inbox,to get out box")
		searchBig(timeout, logger, in.Limit, resp.UserId, list, cache, contentClient)
		interSlice := inter.([]interface{})

		for i := 0; i < len(interSlice); i += 2 {
			str := strings.Split(interSlice[i].(string), ";")

			userId, _ := strconv.ParseInt(str[0], 10, 64)
			contentId, _ := strconv.ParseInt(str[1], 10, 64)
			score, _ := strconv.ParseInt(interSlice[i+1].(string), 10, 64)

			list = append(list, []int64{userId, contentId, score})
		}

	}

	sort.Slice(list, func(i, j int) bool {
		return list[i][2] > list[j][2]
	})

	res := &feedRpc.PullResp{
		UserId:    make([]int64, in.Limit),
		ContentId: make([]int64, in.Limit),
		TimeStamp: make([]int64, in.Limit),
	}

	for i := 0; i < int(in.Limit) && i < len(list); i++ {
		res.UserId[i] = list[i][0]
		res.ContentId[i] = list[i][1]
		res.TimeStamp[i] = list[i][2]
	}

	return res, nil

}

func searchBack(arguments ...interface{}) error {
	ctx := arguments[0].(context.Context)
	logger := arguments[1].(*slog.Logger)
	limit := arguments[2].(int64)
	followingId := arguments[3].([]int64)
	timeStamp := arguments[4].(int64)
	list := arguments[5].([][]int64)
	client := arguments[6].(publicContentRpc.PublicContentServiceClient)

	for _, id := range followingId {
		resp, err := client.GetUserContentList(ctx, &publicContentRpc.GetUserContentListReq{
			Id:        id,
			TimeStamp: timeStamp,
			Limit:     limit,
		})
		if err != nil {
			logger.Error("get user content list:" + err.Error())
			return err
		}
		for i, v := range resp.Id {
			list = append(list, []int64{id, v, resp.TimeStamp[i]})
		}
	}
	return nil
}
