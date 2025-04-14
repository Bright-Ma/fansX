package logic

import (
	"context"
	"errors"
	"fansX/common/util"
	"fansX/internal/model/database"
	"fansX/services/relation/internal/svc"
	"fansX/services/relation/proto/relationRpc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type FollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FollowLogic) Follow(in *relationRpc.FollowReq) (*relationRpc.Empty, error) {
	db := l.svcCtx.DB
	creator := l.svcCtx.Creator
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	logger.Info("user following", "userId", in.UserId, "followingId", in.FollowId)

	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tx := db.WithContext(timeout).Begin()

	nums := &database.FollowingNums{}
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(nums, in.UserId).Error
	if err != nil {
		logger.Error("lock table-following_nums:" + err.Error())
		tx.Commit()
		return nil, err
	} else if nums.Nums == 2000 {
		logger.Info("following nums is not enough", "nums", nums.Nums)
		tx.Commit()
		return nil, errors.New("following nums is not enough")
	}

	logger.Debug("lock table-following_nums")

	record := &database.Following{}
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("follower_id = ? and type in (0,1) and following_id = ?", in.UserId, in.FollowId).
		Take(record).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("search from table-followings:" + err.Error())
		tx.Commit()
		return nil, err

	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		if record.Type == database.Followed {
			logger.Info("also following")
			return &relationRpc.Empty{}, nil
		}

		err = tx.Take(record).Update("type", database.Followed).Error
		if err != nil {
			logger.Error("update table-followings:" + err.Error())
			tx.Rollback()
			return nil, err
		}

	} else {
		id, ok := creator.GetId()
		if !ok {
			logger.Error("get unique id failed")
			return nil, errors.New("id create failed")
		}

		err = tx.Create(&database.Following{
			Id:          id,
			FollowerId:  in.UserId,
			FollowingId: in.FollowId,
			Type:        database.Followed,
		}).Error
		if err != nil {
			logger.Error("create record from table-followings:" + err.Error())
			tx.Rollback()
			return nil, err
		}
	}
	logger.Debug("update table-followings")

	err = tx.Take(&database.FollowingNums{}, in.UserId).Update("nums", gorm.Expr("nums + 1")).Error
	if err != nil {
		logger.Error("update table-following_nums:" + err.Error())
		tx.Rollback()
		return nil, err
	}
	logger.Debug("update table-following_nums")

	tx.Commit()
	return &relationRpc.Empty{}, nil
}
