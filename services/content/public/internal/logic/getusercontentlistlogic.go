package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fansX/internal/middleware/lua"
	"fansX/internal/middleware/lua/script/zset"
	"fansX/internal/model/database"
	"fansX/internal/util"
	"fansX/pkg/hotkey-go/hotkey"
	interlua "fansX/services/content/public/internal/lua"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
	"strconv"
	"time"

	"fansX/services/content/public/internal/svc"
	"fansX/services/content/public/proto/publicContentRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserContentListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type ContentList struct {
	Id        []int64 `json:"id"`
	TimeStamp []int64 `json:"time_stamp"`
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
	cache := l.svcCtx.Core
	executor := l.svcCtx.Executor

	key := "ContentList:" + strconv.FormatInt(in.Id, 10)
	logger.Info("get user content list", "user", in.Id)
	v, ok := cache.Get(key)
	if ok {
		logger.Info("get user content list from local cache")
		var list ContentList
		err := json.Unmarshal(v, &list)
		if err != nil {
			logger.Error("unmarshal json:" + err.Error())
			return nil, err
		}
		return &publicContentRpc.GetUserContentListResp{
			Id:        list.Id,
			TimeStamp: list.TimeStamp,
		}, nil
	}
	hot := cache.IsHotKey(key)

	list, ok, err := contentListFromRedis(timeout, key, in.TimeStamp, executor, logger)
	if ok {
		if hot {
			contentListSet(list, logger, cache, key)
		}
		return &publicContentRpc.GetUserContentListResp{
			Id:        list[0],
			TimeStamp: list[1],
		}, nil
	}

	set := false
	if err == nil {
		set = true
	}

	list, err = contentListFromTiDB(in.Id, in.TimeStamp, in.Limit, logger, db)
	if err != nil {
		return nil, err
	}

	if set {
		data := make([]string, len(list)*2)
		for i := 0; i < len(list); i++ {
			data[i*2] = strconv.FormatInt(list[1][i], 10)
			data[i*2+1] = strconv.FormatInt(list[0][i], 10)
		}
		err = executor.Execute(timeout, luaZset.GetCreate(), []string{key, "false", "900"}, data).Err()
		if err != nil {
			logger.Warn("set content list to redis:" + err.Error())
		}
	}

	if hot {
		contentListSet(list, logger, cache, key)
	}
	return &publicContentRpc.GetUserContentListResp{
		Id:        list[0],
		TimeStamp: list[1],
	}, nil
}

func contentListFromRedis(arguments ...interface{}) ([][]int64, bool, error) {
	ctx := arguments[0].(context.Context)
	key := arguments[1].(string)
	timestamp := arguments[2].(int64)
	executor := arguments[3].(*lua.Executor)
	logger := arguments[4].(*slog.Logger)

	inter, err := executor.Execute(ctx, interlua.GetRevByScoreScript(), []string{key}, 0, timestamp).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Error("execute get rev by score:" + err.Error())
		return nil, false, err
	} else if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}

	logger.Info("get content list from redis")
	interSlice := inter.([]interface{})
	res := make([][]int64, 2)
	for i := 0; i < len(interSlice); i += 2 {
		value, _ := strconv.ParseInt(interSlice[i].(string), 10, 64)
		score, _ := strconv.ParseInt(interSlice[i+1].(string), 10, 64)
		res[0] = append(res[0], value)
		res[1] = append(res[1], score)
	}

	return res, true, nil
}

func contentListFromTiDB(arguments ...interface{}) ([][]int64, error) {
	id := arguments[0].(int64)
	timestamp := arguments[1].(int64)
	limit := arguments[2].(int64)
	logger := arguments[3].(*slog.Logger)
	db := arguments[4].(*gorm.DB)

	infos := make([]database.VisibleContentInfo, 0)
	err := db.Select("id", "created_at").
		Where("userid = ? and status = ? and created_at <= ?", id, database.ContentStatusPass, timestamp).
		Limit(int(limit)).Find(&infos).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("get content list from tidb:" + err.Error())
		return nil, err
	} else if err != nil {
		logger.Info("not content time < " + strconv.FormatInt(timestamp, 10))
		return make([][]int64, 2), nil
	}
	res := make([][]int64, 2)
	res[0] = make([]int64, len(infos))
	res[1] = make([]int64, len(infos))

	logger.Info("get content list from tidb")
	for i, v := range infos {
		res[0][i] = v.Id
		res[1][i] = v.CreatedAt
	}

	return res, nil
}

func contentListSet(arguments ...interface{}) {
	list := arguments[0].([][]int64)
	logger := arguments[1].(*slog.Logger)
	cache := arguments[2].(*hotkey.Core)
	key := arguments[3].(string)

	value, err := json.Marshal(ContentList{
		Id:        list[0],
		TimeStamp: list[1],
	})
	if err != nil {
		logger.Warn("marshal json:" + err.Error())
	} else {
		if !cache.Set(key, value, 60) {
			logger.Debug("set local cache failed")
		} else {
			logger.Debug("set local cache success")
		}
	}
}
