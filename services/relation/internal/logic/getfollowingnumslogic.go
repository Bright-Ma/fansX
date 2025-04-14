package logic

import (
	"context"
	"encoding/binary"
	"errors"
	"fansX/common/util"
	"fansX/internal/model/database"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"

	"fansX/services/relation/internal/svc"
	"fansX/services/relation/proto/relationRpc"

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
	core := l.svcCtx.HotKey

	key := "FollowingNums:" + strconv.FormatInt(in.UserId, 10)

	logger.Info("GetFollowingNums", "userid", in.UserId)
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	v, ok := core.Get(key)
	if ok {
		var nums int64
		logger.Debug("get following nums from local cache")
		_, err := binary.Decode(v, binary.LittleEndian, &nums)
		if err == nil {
			return &relationRpc.GetFollowingNumsResp{Nums: nums}, nil
		}
		logger.Warn("binary decode:" + err.Error())
	} else {
		logger.Debug("not found following nums from local cache")
	}

	hot := core.IsHotKey(key)

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
		if hot {
			buf := make([]byte, 8)
			_, err = binary.Encode(buf, binary.LittleEndian, nums)
			if err != nil {
				logger.Error("binary encode following nums:" + err.Error())
				return nil, err
			}
			core.Set(key, buf, 60)
		}
		return &relationRpc.GetFollowingNumsResp{Nums: nums}, nil
	}

	logger.Debug("not found following nums from redis")

	record, err := l.svcCtx.Single.Do("GetFollowingNums:"+strconv.FormatInt(in.UserId, 10), func() (interface{}, error) {
		record := &database.FollowingNums{}
		timeout, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = db.WithContext(timeout).Take(record, in.UserId).Error
		if err != nil {
			return 0, err
		}

		err = client.Set(timeout, key, record.Nums, time.Minute*15).Err()
		if err != nil {
			logger.Warn("set following nums to redis:" + err.Error())
		}
		if hot {
			buf := make([]byte, 8)
			_, err = binary.Encode(buf, binary.LittleEndian, record.Nums)
			if err != nil {
				logger.Error("binary encode following nums:" + err.Error())
				return nil, err
			}
			core.Set(key, buf, 60)
		}
		return record, nil
	})

	if err != nil {
		logger.Error("get following nums from database:" + err.Error())
		return nil, err
	}

	return &relationRpc.GetFollowingNumsResp{Nums: record.(*database.FollowingNums).Nums}, nil
}
