package logic

import (
	"context"
	"errors"
	"fansX/common/util"
	"fansX/services/content/public/proto/publicContentRpc"
	interlua "fansX/services/feed/internal/lua"
	"fansX/services/relation/proto/relationRpc"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"fansX/services/feed/internal/svc"
	"fansX/services/feed/proto/feedRpc"

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

	set := make(map[[3]int64]bool)

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
		err = searchBack(timeout, logger, in.Limit, resp.UserId, in.TimeStamp, set, contentClient)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Info("find inbox,to get out box")
		searchBig(timeout, logger, in.Limit, resp.UserId, set, cache, contentClient)
		interSlice := inter.([]interface{})

		for i := 0; i < len(interSlice); i += 2 {
			str := strings.Split(interSlice[i].(string), ";")

			userId, _ := strconv.ParseInt(str[0], 10, 64)
			contentId, _ := strconv.ParseInt(str[1], 10, 64)
			score, _ := strconv.ParseInt(interSlice[i+1].(string), 10, 64)
			set[[3]int64{userId, contentId, score}] = true
		}

	}

	top := GetTopK(set, int(in.Limit))

	res := &feedRpc.PullResp{
		UserId:    make([]int64, min(int(in.Limit), len(top))),
		ContentId: make([]int64, min(int(in.Limit), len(top))),
		TimeStamp: make([]int64, min(int(in.Limit), len(top))),
	}

	for i := 0; i < len(res.UserId); i++ {
		item := top[i]
		res.UserId[i] = item[0]
		res.ContentId[i] = item[1]
		res.TimeStamp[i] = item[2]
	}

	return res, nil
}

func searchBack(arguments ...interface{}) error {
	ctx := arguments[0].(context.Context)
	logger := arguments[1].(*slog.Logger)
	limit := arguments[2].(int64)
	followingId := arguments[3].([]int64)
	timeStamp := arguments[4].(int64)
	set := arguments[5].(map[[3]int64]bool)
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
			set[[3]int64{id, v, resp.TimeStamp[i]}] = true
		}
	}
	return nil
}
