package logic

import (
	"context"
	"fansX/common/util"
	"fansX/internal/model/database"
	"fansX/services/relation/internal/svc"
	"fansX/services/relation/proto/relationRpc"
	"gorm.io/gorm"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelFollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelFollowLogic {
	return &CancelFollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelFollowLogic) CancelFollow(in *relationRpc.CancelFollowReq) (*relationRpc.Empty, error) {
	db := l.svcCtx.DB
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	logger.Info("user cancelFollowed", "userId", in.UserId, "followedId", in.FollowId)

	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx := db.WithContext(timeout).Begin()

	res := tx.Model(&database.Following{}).
		Where("follower_id = ? and type = ? and following_id = ?", in.UserId, database.Followed, in.FollowId).
		Update("type", database.UnFollowed)

	if res.Error != nil {
		logger.Error("update table-following:" + res.Error.Error())
		tx.Rollback()
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		logger.Info("also cancel following relation")
		tx.Commit()
		return &relationRpc.Empty{}, nil
	}

	logger.Debug("update table-following")

	err := tx.Take(&database.FollowingNums{}, in.UserId).Update("nums", gorm.Expr("nums - 1")).Error
	if err != nil {
		logger.Error("update table-following_nums:" + err.Error())
		tx.Rollback()
		return nil, err
	}

	logger.Debug("update table-following_nums")
	tx.Commit()
	return &relationRpc.Empty{}, nil
}
