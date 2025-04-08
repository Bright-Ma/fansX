package logic

import (
	"bilibili/common/util"
	"bilibili/internal/model/database"
	"bilibili/services/content/meta/internal/svc"
	"bilibili/services/content/meta/proto/metaContentRpc"
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteLogic) Delete(in *metaContentRpc.DeleteReq) (*metaContentRpc.Empty, error) {
	db := l.svcCtx.DB
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	logger.Info("user delete content", "userId", in.UserId, "contentId", in.ContentId)

	tx := db.Begin()

	record := &database.InvisibleContentInfo{}
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(record, in.ContentId).Error
	if err != nil {
		tx.Commit()
		logger.Error("select record err:" + err.Error())
		return nil, err
	}

	if in.UserId != record.Userid {
		tx.Commit()
		logger.Error("user delete content info:is not the publish user")
		return nil, errors.New("you can not do this it is not your content")
	}

	err = tx.Take(&database.InvisibleContentInfo{}, record.Id).
		Update("status", database.ContentStatusDelete).
		Update("version", gorm.Expr("version + 1")).Error
	if err != nil {
		tx.Rollback()
		logger.Error("update(delete) content record:" + err.Error())
		return nil, err
	}

	tx.Commit()
	return &metaContentRpc.Empty{}, nil
}
