package logic

import (
	"bilibili/common/lua"
	luaZset "bilibili/common/lua/script/zset"
	"bilibili/common/util"
	bigcache "bilibili/internal/cache"
	"bilibili/services/content/public/proto/publicContentRpc"
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

type PullLatestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPullLatestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PullLatestLogic {
	return &PullLatestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PullLatestLogic) PullLatest(in *feedRpc.PullLatestReq) (*feedRpc.PullResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	client := l.svcCtx.RClient
	cache := l.svcCtx.Cache
	relationClient := l.svcCtx.RelationClient
	contentClient := l.svcCtx.ContentClient
	executor := l.svcCtx.Executor

	logger.Info("user pull latest feed stream", "userId", in.UserId, "size", in.Limit)

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
	inter, err := client.Eval(timeout, `
local key=KEYS[1]
local limit=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==0
    then return nil
end

local res=redis.call("ZREVRANGE",0,tonumber(limit),"WITHSCORES")
return res
`, []string{inbox}, in.Limit).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Error("get inbox:" + err.Error())
		return nil, err
	} else if errors.Is(err, redis.Nil) {
		err = searchAll(timeout, logger, in.Limit, in.UserId, resp.UserId, list, cache, executor, contentClient)
		if err != nil {
			return nil, err
		}
	} else {
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

func searchBig(ctx context.Context, logger *slog.Logger, limit int64, followingId []int64, list [][]int64, cache *bigcache.Cache, contentClient publicContentRpc.PublicContentServiceClient) {
	for _, id := range followingId {
		if cache.IsBig(id) {
			resp, err := contentClient.GetUserContentList(ctx, &publicContentRpc.GetUserContentListReq{
				Id:        id,
				TimeStamp: time.Now().Unix(),
				Limit:     limit,
			})
			if err != nil {
				logger.Error("get user content list:"+err.Error(), "userId", id)
			} else {
				for i, v := range resp.Id {
					list = append(list, []int64{id, v, resp.TimeStamp[i]})
				}
			}
		}
	}
	return
}

func searchAll(ctx context.Context, logger *slog.Logger, limit int64, userId int64, followingId []int64, list [][]int64, cache *bigcache.Cache, executor *lua.Executor, contentClient publicContentRpc.PublicContentServiceClient) error {
	build := make([]string, 0)
	for _, id := range followingId {
		resp, err := contentClient.GetUserContentList(ctx, &publicContentRpc.GetUserContentListReq{
			Id:        id,
			TimeStamp: time.Now().Unix(),
			Limit:     limit,
		})
		if err != nil {
			logger.Error("get user content list:"+err.Error(), "userId", id)
			return err
		} else {
			for i, v := range resp.Id {
				list = append(list, []int64{id, v, resp.TimeStamp[i]})
				if !cache.IsBig(id) {
					build = append(build, strconv.FormatInt(resp.TimeStamp[i], 10))
					build = append(build, strconv.FormatInt(id, 10)+";"+strconv.FormatInt(v, 10))
				}
			}
		}
	}

	err := executor.Execute(ctx, luaZset.GetCreate(), []string{"inbox:" + strconv.FormatInt(userId, 10), "false", "864000"}, build).Err()
	if err != nil {
		logger.Error("execute create zset:" + err.Error())
	}

	return nil
}
