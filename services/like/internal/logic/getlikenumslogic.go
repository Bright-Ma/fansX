package logic

import (
	"context"
	"encoding/binary"
	"errors"
	"fansX/internal/model/database"
	"fansX/internal/util"
	"fansX/pkg/hotkey-go/hotkey"
	"fansX/services/like/internal/script"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
	"strconv"
	"time"

	"fansX/services/like/internal/svc"
	"fansX/services/like/proto/likeRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLikeNumsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLikeNumsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikeNumsLogic {
	return &GetLikeNumsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLikeNumsLogic) GetLikeNums(in *likeRpc.GetLikeNumsReq) (*likeRpc.GetLikeNumsResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cache := l.svcCtx.Cache
	client := l.svcCtx.Client
	db := l.svcCtx.DB
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	executor := l.svcCtx.Executor

	key := "LikeNums:" + strconv.Itoa(int(in.BusinessId)) + ":" + strconv.FormatInt(in.LikeId, 10)
	body, ok := cache.Get(key)
	if ok {
		return &likeRpc.GetLikeNumsResp{Have: true, Nums: int64(binary.BigEndian.Uint64(body))}, nil
	}
	hot := cache.IsHotKey(key)
	res, err := SearchNumsFromRedis(timeout, client, logger, key)
	if err == nil {
		if hot {
			SetNumsCache(cache, res, key)
		}
		return &likeRpc.GetLikeNumsResp{Have: true, Nums: res}, nil
	}
	var rebuild bool
	if errors.Is(err, redis.Nil) {
		rebuild = true
	}

	resp, err := l.svcCtx.Group.Do(key, func() (interface{}, error) {
		resp, err := SearchNumsFromTiDB(db, client, key, cache, hot, in)
		if err != nil {
			return nil, err
		}
		if rebuild {
			executor.Execute(timeout, script.Set, []string{key}, resp.(*likeRpc.GetLikeNumsResp).Nums)
		}
		return resp, nil
	})

	if err == nil {
		return resp.(*likeRpc.GetLikeNumsResp), nil
	} else {
		logger.Error("search like count from tidb:" + err.Error())
		return nil, err
	}
}

func SearchNumsFromRedis(ctx context.Context, client *redis.Client, logger *slog.Logger, key string) (int64, error) {
	str, err := client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		logger.Debug("not found like count from redis:" + err.Error())
		return 0, err
	} else if err != nil {
		logger.Error("search like count from redis:" + err.Error())
		return 0, err
	}
	logger.Debug("found like count from redis", "res", str)
	res, _ := strconv.ParseInt(str, 10, 64)

	return res, nil
}

func SearchNumsFromTiDB(db *gorm.DB, client *redis.Client, key string, cache *hotkey.Core, hot bool, in *likeRpc.GetLikeNumsReq) (interface{}, error) {
	record := database.LikeCount{}
	timeout, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	err := db.Select("count").Where("business = ? and like_id = ?", in.BusinessId, in.LikeId).Take(&record).Error
	if err == nil {
		client.Set(timeout, key, record.Count, time.Second*30)
		if hot {
			SetNumsCache(cache, record.Count, key)
		}
		return &likeRpc.GetLikeNumsResp{Have: true, Nums: record.Count}, nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return &likeRpc.GetLikeNumsResp{Have: false}, nil
	} else {
		return nil, err
	}
}

func SetNumsCache(cache *hotkey.Core, nums int64, key string) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(nums))
	if nums < 10000 {
		cache.Set(key, buf, 5)
	} else {
		cache.Set(key, buf, 30)
	}
	return
}
