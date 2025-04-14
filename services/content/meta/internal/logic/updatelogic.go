package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fansX/common/util"
	"fansX/internal/model/database"
	"fansX/services/content/meta/internal/svc"
	"fansX/services/content/meta/proto/metaContentRpc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateLogic) Update(in *metaContentRpc.UpdateReq) (*metaContentRpc.Empty, error) {
	db := l.svcCtx.DB
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)

	logger.Info("user update content info", "userId", in.UserId, "contentId", in.ContentId)

	tx := db.Begin()
	record := &database.InvisibleContentInfo{}

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("status <> ?", database.ContentStatusDelete).Take(record, in.ContentId).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Commit()
		logger.Error("search content failed:" + err.Error())
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Commit()
		logger.Info("record not found")
		return nil, err
	}

	if record.Userid != in.UserId {
		tx.Commit()
		logger.Error("user update content info:is not the publish user")
		return nil, errors.New("you can not do this it is not your content")
	}

	v, err := json.Marshal(in.VideoUriList)
	if err != nil {
		tx.Commit()
		logger.Error("json marshal video uri list:" + err.Error())
		return nil, err
	}
	p, err := json.Marshal(in.PhotoUriList)
	if err != nil {
		tx.Commit()
		logger.Error("json marshal photo uri list:" + err.Error())
		return nil, err
	}
	upd := &database.InvisibleContentInfo{
		Id:              in.ContentId,
		Version:         record.Version + 1,
		Status:          database.ContentStatusCheck,
		OldStatus:       record.Status,
		Userid:          record.Userid,
		Title:           in.Title,
		PhotoUriList:    string(p),
		ShortText:       in.ShortText,
		LongTextUri:     in.LongTextUri,
		VideoUriList:    string(v),
		OldPhotoUriList: record.PhotoUriList,
		OldShortText:    record.ShortText,
		OldLongTextUri:  record.LongTextUri,
		OldVideoUriList: record.VideoUriList,
	}

	err = tx.Take(&database.InvisibleContentInfo{}, in.ContentId).Updates(upd).Error
	if err != nil {
		tx.Rollback()
		logger.Error("update content failed:" + err.Error())
		return nil, err
	}

	tx.Commit()

	return &metaContentRpc.Empty{}, nil
}
