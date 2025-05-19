package logic

import (
	"context"
	"encoding/json"
	"fansX/internal/model/mq"
	"fansX/internal/util"
	"fansX/services/comment/internal/svc"
	"fansX/services/comment/proto/commentRpc"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

type DelCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDelCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DelCommentLogic {
	return &DelCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DelCommentLogic) DelComment(in *commentRpc.DelCommentReq) (*commentRpc.Empty, error) {
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	producer := l.svcCtx.Producer
	logger.Info("user del comment", "userId", in.UserId, "CommentId", in.CommentId)
	msg := mq.DelCommentKafkaJson{
		UserId:    in.UserId,
		CommentId: in.CommentId,
	}
	value, err := json.Marshal(msg)
	if err != nil {
		logger.Error("marshal msg to del comment kafka msg json:" + err.Error())
		return nil, err
	}
	message := sarama.ProducerMessage{
		Topic: "",
		Value: sarama.ByteEncoder(value),
	}
	_, _, err = producer.SendMessage(&message)
	if err != nil {
		logger.Error("send message to kafka:" + err.Error())
		return nil, err
	}

	return &commentRpc.Empty{}, nil
}
