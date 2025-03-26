package logic

import (
	"bilibili/common/util"
	"bilibili/model"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"

	"bilibili/services/relation/internal/svc"
	"bilibili/services/relation/proto/relationRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFollowingNumsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFollowingNumsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFollowingNumsLogic {
	return &GetFollowingNumsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFollowingNumsLogic) GetFollowingNums(in *relationRpc.GetFollowingNumsReq) (*relationRpc.GetFollowingNumsResp, error) {
	db := l.svcCtx.DB
	logger := util.SetTrace(context.Background(), l.svcCtx.Logger)
	client := l.svcCtx.RClient

	logger.Info("GetFollowingNums", "userid", in.UserId)
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	key := "FollowingNums:" + strconv.FormatInt(in.UserId, 10)

	res, err := client.Get(timeout, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Error("get following nums from redis:" + err.Error())
		return nil, err
	}
	if err == nil {
		nums, err := strconv.ParseInt(res, 10, 64)
		if err != nil {
			logger.Error("parse following nums:" + err.Error())
			return nil, err
		}
		logger.Debug("get following nums from redis")
		return &relationRpc.GetFollowingNumsResp{Nums: nums}, nil
	}

	logger.Debug("not found following nums from redis")

	record, err := l.svcCtx.Single.Do("GetFollowingNums:"+strconv.FormatInt(in.UserId, 10), func() (interface{}, error) {
		record := &model.FollowingNums{}
		err = db.Take(record, in.UserId).Error
		if err != nil {
			return 0, err
		}
		client.Set(context.Background(), key, record.Nums, time.Minute*15)
		return record, nil
	})

	if err != nil {
		logger.Error("get following nums from tidb:" + err.Error())
		return nil, err
	}

	return &relationRpc.GetFollowingNumsResp{Nums: record.(*model.FollowingNums).Nums}, nil
}
