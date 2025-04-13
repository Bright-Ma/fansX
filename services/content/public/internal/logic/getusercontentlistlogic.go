package logic

import (
	"bilibili/common/lua"
	"bilibili/common/util"
	servicemodel "bilibili/internal/model/services"
	interlua "bilibili/services/content/public/internal/lua"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"bilibili/services/content/public/internal/svc"
	"bilibili/services/content/public/proto/publicContentRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserContentListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserContentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserContentListLogic {
	return &GetUserContentListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserContentListLogic) GetUserContentList(in *publicContentRpc.GetUserContentListReq) (*publicContentRpc.GetUserContentListResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	db := l.svcCtx.DB.WithContext(timeout)
	client := l.svcCtx.RClient
	cache := l.svcCtx.Core
	executor := l.svcCtx.Executor

	key := "ContentList:" + strconv.FormatInt(in.Id, 10)
	logger.Info("get user content list", "user", in.Id)
	v, ok := cache.Get(key)
	if ok {
		logger.Debug("get user content list from local cache")
		var list servicemodel.ContentList
		err := json.Unmarshal(v, &list)
		if err != nil {
			logger.Error("json unmarshal:" + err.Error())
			return nil, err
		}
		index := sort.Search(len(list.TimeStamp), func(i int) bool {
			return list.TimeStamp[i] <= in.TimeStamp
		})
		if int64(len(list.TimeStamp)-index) >= in.Limit {
			return &publicContentRpc.GetUserContentListResp{Id: list.Id[index : int64(index)+in.Limit-1]}, nil
		}
	}

	id, ok, err := searchListFromRedis(timeout, logger, client, executor, key, in.Id, in.TimeStamp, int(in.Limit))
	if ok {
		return &publicContentRpc.GetUserContentListResp{Id: id}, nil
	}

	return nil, nil
}

func searchListFromRedis(ctx context.Context, logger *slog.Logger, client *redis.Client, executor *lua.Executor, key string, userId int64, timeStamp int64, limit int) ([]int64, bool, error) {
	executor.Execute(ctx, interlua.GetRevScript(), []string{key})

}

func searchListFromTiDB(ctx context.Context, logger *slog.Logger, db *gorm.DB, userid int64, timeStamp int64, limit int) ([]int64, error) {

}
