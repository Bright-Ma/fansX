package logic

import (
	"context"
	"errors"
	luaZset2 "fansX/internal/middleware/lua/script/zset"
	"fansX/internal/model/database"
	"fansX/internal/util"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"strconv"
	"time"

	"fansX/services/relation/internal/svc"
	"fansX/services/relation/proto/relationRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFollowerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFollowerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowerLogic {
	return &ListFollowerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFollowerLogic) ListFollower(in *relationRpc.ListFollowerReq) (*relationRpc.ListFollowerResp, error) {
	db := l.svcCtx.DB
	logger := util.SetTrace(context.Background(), l.svcCtx.Logger)
	executor := l.svcCtx.Executor

	logger.Info("ListFollower", "userId", in.UserId, "limit", in.Limit, "offset", in.Offset)
	if in.Limit+in.Offset > 5000 {
		logger.Info("page over")
		return nil, errors.New("page over")
	}
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := "Follower:" + strconv.FormatInt(in.UserId, 10)
	fields, err := executor.Execute(timeout, luaZset2.GetRevRange(), []string{key}, "false", in.Offset, in.Limit+in.Offset-1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Error("search follower from redis:" + err.Error())
		return nil, err
	}

	if err == nil {
		logger.Debug("search follower list from redis")

		res := make([]int64, len(fields.([]interface{})))
		for i, v := range fields.([]interface{}) {

			res[i], err = strconv.ParseInt(v.(string), 10, 64)
			if err != nil {
				logger.Error("parse follower:"+err.Error(), "index", i, "num", v)
				return nil, err
			}
		}
		return &relationRpc.ListFollowerResp{UserId: res}, nil
	}

	record, err := l.svcCtx.Single.Do("ListFollower:"+strconv.FormatInt(in.UserId, 10), func() (interface{}, error) {
		record := make([]database.Follower, 0)
		timeout, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = db.WithContext(timeout).Select("follower_id", "updated_at").
			Where("following_id = ? and type = ?", in.UserId, database.Followed).Limit(5000).Find(&record).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		go func() {
			data := make([]interface{}, len(record)*2)
			for i, v := range record {
				data[i*2] = strconv.FormatInt(v.UpdatedAt, 10)
				data[i*2+1] = strconv.FormatInt(v.FollowerId, 10)
			}
			timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			err := executor.Execute(timeout, luaZset2.GetCreate(), []string{key, "false", "300"}, data...).Err()
			if err != nil {
				logger.Warn("execute zset_create:" + err.Error())
			}
		}()
		return record, nil
	})

	records := record.([]database.Follower)
	start := min(len(records), int(in.Offset))
	end := min(len(records)-1, int(in.Limit+in.Offset-1))

	if start > end {
		logger.Debug("over page size")
		return &relationRpc.ListFollowerResp{UserId: make([]int64, 0)}, nil
	}

	res := make([]int64, end-start+1)
	for i := start; i <= end; i++ {
		res[i-start] = records[i].FollowerId
	}

	logger.Info("get follower list from database", "nums", len(res))
	return &relationRpc.ListFollowerResp{UserId: res}, nil
}
