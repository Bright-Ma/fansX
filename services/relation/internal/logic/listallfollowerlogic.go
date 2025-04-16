package logic

import (
	"context"
	"fansX/internal/model/database"
	"fansX/internal/util"
	"time"

	"fansX/services/relation/internal/svc"
	"fansX/services/relation/proto/relationRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAllFollowerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAllFollowerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAllFollowerLogic {
	return &ListAllFollowerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAllFollowerLogic) ListAllFollower(in *relationRpc.ListAllFollowerReq) (*relationRpc.ListFollowerResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	db := l.svcCtx.DB.WithContext(timeout)
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)

	logger.Info("list all follower", "userId", in.UserId)

	record := make([]database.Follower, 0)
	err := db.Select("follower_id").
		Where("following_id = ? and type = ?", in.UserId, database.Followed).Find(&record).Error
	if err != nil {
		logger.Error("search follower id from TiDB:" + err.Error())
		return nil, err
	}

	res := &relationRpc.ListFollowerResp{UserId: make([]int64, len(record))}
	for i := 0; i < len(record); i++ {
		res.UserId[i] = record[i].FollowerId
	}

	return res, nil
}
