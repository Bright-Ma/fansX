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
	"time"
)

type PublishLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishLogic {
	return &PublishLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PublishLogic) Publish(in *metaContentRpc.PublishReq) (*metaContentRpc.Empty, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	creator := l.svcCtx.Creator
	db := l.svcCtx.DB.WithContext(timeout)
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)

	logger.Info("user publish", "userId", in.UserId)

	id, ok := creator.GetId()
	if !ok {
		logger.Error("creator create id failed")
		return nil, errors.New("get id failed")
	}
	p, err := json.Marshal(in.PhotoUriList)
	if err != nil {
		logger.Error("json marshal photo uri list:"+err.Error(), "list", in.PhotoUriList)
		return nil, err
	}
	v, err := json.Marshal(in.VideoUriList)
	if err != nil {
		logger.Error("json marshal video uri list:"+err.Error(), "list", in.VideoUriList)
		return nil, err
	}

	record := database.InvisibleContentInfo{
		Id:           id,
		Version:      1,
		Status:       database.ContentStatusCheck,
		Userid:       in.UserId,
		Title:        in.Title,
		PhotoUriList: string(p),
		ShortText:    in.ShortText,
		LongTextUri:  in.LongTextUri,
		VideoUriList: string(v),
	}
	tx := db.Begin()
	err = tx.Create(&record).Error
	if err != nil {
		tx.Rollback()
		logger.Error("create invisible content record:" + err.Error())
		return nil, err
	}

	tx.Commit()

	return &metaContentRpc.Empty{}, nil

}
