package logic

import (
	luaZset "bilibili/common/lua/script/zset"
	"bilibili/common/util"
	"bilibili/internal/model/database"
	"context"
	"errors"
	"gorm.io/gorm"
	"strconv"
	"time"

	"bilibili/services/relation/internal/svc"
	"bilibili/services/relation/proto/relationRpc"

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
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	key := "follower:zset:" + strconv.FormatInt(in.UserId, 10)
	fields, err := executor.Execute(timeout, luaZset.GetRevRange(), []string{key}, "false", in.Offset, in.Limit+in.Offset-1).Result()
	if err != nil {
		logger.Error("search follower from redis:" + err.Error())
		return nil, err
	}
	if fields != nil {
		res := make([]int64, len(fields.([]string)))
		for i, v := range fields.([]string) {
			res[i], err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				logger.Error("parse follower:"+err.Error(), "index", i, "num", v)
				return nil, err
			}
			return &relationRpc.ListFollowerResp{UserId: res}, nil
		}
	}

	record, err := l.svcCtx.Single.Do("ListFollower:"+strconv.FormatInt(in.UserId, 10), func() (interface{}, error) {
		record := make([]database.Follower, 0)
		err = db.Select("follower_id", "updated_at").
			Where("following_id = ? and type = ?", in.UserId, database.Followed).Limit(5000).Find(&record).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		go func() {
			data := make([]string, len(record)*2)
			for i, v := range record {
				data[i*2] = strconv.FormatInt(v.UpdatedAt, 10)
				data[i*2+1] = strconv.FormatInt(v.FollowerId, 10)
			}
			executor.Execute(context.Background(), luaZset.GetCreate(), []string{key, "false"}, data)
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
