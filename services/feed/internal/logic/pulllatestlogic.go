package logic

import (
	"context"
	"errors"
	bigcache "fansX/internal/middleware/cache"
	"fansX/internal/middleware/lua"
	"fansX/internal/middleware/lua/script/zset"
	"fansX/internal/util"
	heapx "fansX/pkg/heapx"
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

	set := make(map[[3]int64]bool)
	inbox := "inbox:" + strconv.FormatInt(in.UserId, 10)

	inter, err := executor.Execute(timeout, interlua.GetRevRangeScript(), []string{inbox}, in.Limit).Result()
	if err != nil && !errors.Is(err, redis.Nil) {

		logger.Error("get inbox:" + err.Error())
		return nil, err

	} else if errors.Is(err, redis.Nil) {

		logger.Info("not find inbox")
		err = searchAll(timeout, logger, in.UserId, resp.UserId, set, cache, executor, contentClient)
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

func searchBig(arguments ...interface{}) {

	ctx := arguments[0].(context.Context)
	logger := arguments[1].(*slog.Logger)
	limit := arguments[2].(int64)
	followingId := arguments[3].([]int64)
	set := arguments[4].(map[[3]int64]bool)
	cache := arguments[5].(*bigcache.Cache)
	client := arguments[6].(publicContentRpc.PublicContentServiceClient)

	for _, id := range followingId {

		if cache.IsBig(id) {

			resp, err := client.GetUserContentList(ctx, &publicContentRpc.GetUserContentListReq{
				Id:        id,
				TimeStamp: time.Now().Unix(),
				Limit:     limit,
			})

			if err != nil {
				logger.Error("get user content list:"+err.Error(), "userId", id)
			} else {
				for i, v := range resp.Id {
					set[[3]int64{id, v, resp.TimeStamp[i]}] = true
				}
			}

		}

	}
	return
}

func searchAll(arguments ...interface{}) error {
	ctx := arguments[0].(context.Context)
	logger := arguments[1].(*slog.Logger)
	userId := arguments[2].(int64)
	followingId := arguments[3].([]int64)
	set := arguments[4].(map[[3]int64]bool)
	cache := arguments[5].(*bigcache.Cache)
	executor := arguments[6].(*lua.Executor)
	client := arguments[7].(publicContentRpc.PublicContentServiceClient)

	build := heapx.NewHeap[[]int64](func(a, b []int64) bool {
		return a[2] > b[2]
	})

	for _, id := range followingId {

		resp, err := client.GetUserContentList(ctx, &publicContentRpc.GetUserContentListReq{
			Id:        id,
			TimeStamp: time.Now().Unix(),
			Limit:     100,
		})

		if err != nil {
			logger.Error("get user content list:"+err.Error(), "userId", id)
			return err
		} else {
			for i, v := range resp.Id {
				set[[3]int64{id, v, resp.TimeStamp[i]}] = true
				if !cache.IsBig(id) {
					build.PushItem([]int64{id, v, resp.TimeStamp[i]})
				}
			}
		}

	}
	argv := make([]string, min(200, build.Len()))
	for i := 0; i < len(argv); i += 2 {
		item := build.PopItem()
		argv[i] = strconv.FormatInt(item[2], 10)
		argv[i+1] = strconv.FormatInt(item[0], 10) + ";" + strconv.FormatInt(item[1], 10)
	}

	err := executor.Execute(ctx, luaZset.GetCreate(), []string{"inbox:" + strconv.FormatInt(userId, 10), "false", "864000"}, argv).Err()
	if err != nil {
		logger.Error("execute create ZSet:" + err.Error())
	}

	return nil
}

func GetTopK(set map[[3]int64]bool, k int) [][3]int64 {
	heap := heapx.NewHeap[[3]int64](func(a, b [3]int64) bool {
		return a[2] < b[2]
	})
	for key := range set {
		heap.PushItem(key)
		if heap.Len() > k {
			heap.Pop()
		}
	}
	res := make([][3]int64, min(k, heap.Len()))
	for heap.Len() != 0 {
		res = append(res, heap.PopItem())
	}

	return res
}
