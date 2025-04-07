package logic

import (
	"bilibili/common/util"
	"bilibili/internal/model/database"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"

	"bilibili/services/relation/internal/svc"
	"bilibili/services/relation/proto/relationRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFollowerNumsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFollowerNumsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFollowerNumsLogic {
	return &GetFollowerNumsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFollowerNumsLogic) GetFollowerNums(in *relationRpc.GetFollowerNumsReq) (*relationRpc.GetFollowerNumsResp, error) {
	db := l.svcCtx.DB
	logger := util.SetTrace(context.Background(), l.svcCtx.Logger)
	client := l.svcCtx.RClient

	key := "FollowerNums:" + strconv.FormatInt(in.UserId, 10)

	logger.Info("GetFollowerNums", "userid", in.UserId)
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.Get(timeout, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Error("get follower nums from redis:" + err.Error())
		return nil, err
	}
	if err == nil {
		nums, err := strconv.ParseInt(res, 10, 64)
		if err != nil {
			logger.Error("parse follower nums:" + err.Error())
			return nil, err
		}
		logger.Debug("get follower nums from redis")
		return &relationRpc.GetFollowerNumsResp{Nums: nums}, nil
	}

	logger.Debug("not found follower nums from redis")

	record, err := l.svcCtx.Single.Do("GetFollowerNums:"+strconv.FormatInt(in.UserId, 10), func() (interface{}, error) {
		record := &database.FollowerNums{}
		timeout, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = db.WithContext(timeout).Take(record, in.UserId).Error
		if err != nil {
			return 0, err
		}

		err = client.Set(timeout, key, record.Nums, time.Minute*5).Err()
		if err != nil {
			logger.Warn("set follower nums to redis:" + err.Error())
		}
		return record, nil
	})

	if err != nil {
		logger.Error("get follower nums from database:" + err.Error())
		return nil, err
	}

	return &relationRpc.GetFollowerNumsResp{Nums: record.(*database.FollowerNums).Nums}, nil
}
