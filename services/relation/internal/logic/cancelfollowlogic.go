package logic

import (
	"bilibili/common/util"
	"bilibili/internal/model/database"
	"bilibili/services/relation/internal/svc"
	"bilibili/services/relation/proto/relationRpc"
	"context"
	"errors"
	"gorm.io/gorm"

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
	tx := db.Begin()
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)

	logger.Info("user cancelFollowed", "userId", in.UserId, "followedId", in.FollowId)

	err := tx.Model(&database.Following{}).
		Where("follower_id = ? and type = ? and following_id = ?", in.UserId, database.Followed, in.FollowId).
		Update("type", database.UnFollowed).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("update table-following:" + err.Error())
		tx.Rollback()
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Debug("not found record from table-following")
		tx.Commit()
		return &relationRpc.Empty{}, nil
	}

	logger.Debug("update table-following")

	err = tx.Take(&database.FollowingNums{}, in.UserId).Where("nums", gorm.Expr("nums - 1")).Error
	if err != nil {
		logger.Error("update table-following_nums:" + err.Error())
		tx.Rollback()
		return nil, err
	}

	logger.Debug("update table-following_nums")
	tx.Commit()
	return &relationRpc.Empty{}, nil
}
