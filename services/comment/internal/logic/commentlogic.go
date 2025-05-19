package logic

import (
	"context"
	"encoding/json"
	"fansX/internal/model/mq"
	"fansX/internal/util"
	"github.com/IBM/sarama"
	"strconv"
	"time"

	"fansX/services/comment/internal/svc"
	"fansX/services/comment/proto/commentRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentLogic {
	return &CommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CommentLogic) Comment(in *commentRpc.CommentReq) (*commentRpc.Empty, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	creator := l.svcCtx.Creator
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	producer := l.svcCtx.Producer

	logger.Info("user comment", "user", in.UserId, "contentId", in.ContentId, "rootId", in.RootId, "parentId", in.ParentId)
	id, err := creator.GetIdWithContext(timeout)
	if err != nil {
		logger.Error("get id:" + err.Error())
		return nil, err
	}
	msg := mq.CommentKafkaJson{
		Id:          id,
		UserId:      in.UserId,
		ContentId:   in.ContentId,
		RootId:      in.RootId,
		ParentId:    in.ParentId,
		ShortText:   in.ShortText,
		LongTextUri: in.LongTextUri,
	}
	value, err := json.Marshal(msg)
	if err != nil {
		logger.Error("marshal msg to comment kafka msg json:" + err.Error())
	}
	message := sarama.ProducerMessage{
		Topic: "",
		Key:   sarama.StringEncoder(strconv.FormatInt(in.ContentId, 10)),
		Value: sarama.ByteEncoder(value),
	}
	_, _, err = producer.SendMessage(&message)
	if err != nil {
		logger.Error("send msg to kafka:" + err.Error())
		return nil, err
	}
	return &commentRpc.Empty{}, nil
}
