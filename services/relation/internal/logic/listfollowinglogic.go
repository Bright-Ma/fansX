package logic

import (
	"bilibili/common/util"
	"bilibili/model"
	"context"
	"errors"
	"gorm.io/gorm"
	"strconv"
	"time"

	"bilibili/services/relation/internal/svc"
	"bilibili/services/relation/proto/relationRpc"

	"bilibili/common/lua/script/zset"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListFollowingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFollowingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowingLogic {
	return &ListFollowingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFollowingLogic) ListFollowing(in *relationRpc.ListFollowingReq) (*relationRpc.ListFollowingResp, error) {
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	db := l.svcCtx.DB
	executor := l.svcCtx.Executor

	logger.Info("listFollowing", "userid", in.UserId, "all", in.All, "limit", in.Limit, "offset", in.Offset)

	key := "following:zset:" + strconv.FormatInt(in.UserId, 10)
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	table, err := executor.Execute(timeout, luaZset.GetRevRange(), []string{key}, in.All, in.Offset, in.Limit+in.Offset-1).Result()
	if err != nil {
		logger.Error("execute lua zset-revrange:" + err.Error())
		return nil, err
	}

	if table != nil {
		str := table.([]string)
		res := make([]int64, len(str))

		for i, v := range str {
			res[i], err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				logger.Error("parse int:"+err.Error(), "index", i, "value", v)
				return nil, err
			}
		}

		logger.Info("get following list from redis", "nums", len(res))
		return &relationRpc.ListFollowingResp{UserId: res}, nil
	}

	logger.Info("following list not exists from redis")

	record, err := l.svcCtx.Single.Do("ListFollowing:"+strconv.FormatInt(in.UserId, 10), func() (interface{}, error) {

		record := make([]model.Following, 0)
		err = db.Select("following_id", "updated_at").
			Where("follower_id = ?  and type = ?", in.UserId, model.Followed).Find(&record).Error

		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		go func() {
			si := make([]string, len(record)*2)
			for i, v := range record {
				si[i*2] = strconv.FormatInt(v.UpdatedAt, 10)
				si[i*2+1] = strconv.FormatInt(v.FollowingId, 10)
			}
			executor.Execute(context.Background(), luaZset.GetCreate(), []string{key, "false"}, si)
		}()

		return record, nil
	})

	if err != nil {
		logger.Error("search table-following:" + err.Error())
		return nil, err
	}

	records := record.([]model.Following)
	start := min(len(records), int(in.Offset))
	end := min(len(records)-1, int(in.Limit+in.Offset-1))

	if start > end {
		logger.Debug("over page size")
		return &relationRpc.ListFollowingResp{UserId: make([]int64, 0)}, nil
	}

	res := make([]int64, end-start+1)
	for i := start; i <= end; i++ {
		res[i-start] = records[i].FollowingId
	}

	logger.Info("get following list from tidb", "nums", len(res))
	return &relationRpc.ListFollowingResp{UserId: res}, nil

}
