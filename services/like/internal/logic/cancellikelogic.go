package logic

import (
	"context"
	"encoding/json"
	"fansX/internal/model/mq"
	"fansX/internal/util"
	"fansX/services/like/internal/svc"
	"fansX/services/like/proto/likeRpc"
	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"log/slog"
	"strconv"
	"time"
)

type CancelLikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelLikeLogic {
	return &CancelLikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelLikeLogic) CancelLike(in *likeRpc.CancelLikeReq) (*likeRpc.Empty, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	msg := mq.LikeKafkaJson{
		TimeStamp: in.TimeStamp,
		Business:  in.BusinessId,
		UserId:    in.UserId,
		LikeId:    in.LikeId,
		Cancel:    true,
	}
	logger.Info("user cancel like", "business", msg.Business, "userid", msg.UserId, "likeId", in.LikeId, "timeStamp", in.TimeStamp)
	value, err := json.Marshal(msg)
	if err != nil {
		slog.Error("marshal json:" + err.Error())
		return &likeRpc.Empty{}, nil
	}

	message := &sarama.ProducerMessage{
		Topic: "",
		Key:   sarama.StringEncoder(strconv.FormatInt(in.LikeId, 10)),
		Value: sarama.ByteEncoder(value),
	}
	i := 0
	for {
		select {
		case <-timeout.Done():
			logger.Error("produce message time out")
			return nil, context.DeadlineExceeded
		default:
			i++
			_, _, err = l.svcCtx.Producer.SendMessage(message)
			if err != nil {
				logger.Error("producer send message to kafka:"+err.Error(), "times", i)
				time.Sleep(time.Millisecond * 100)
				continue
			}
		}
		break

	}

	return &likeRpc.Empty{}, nil
}
