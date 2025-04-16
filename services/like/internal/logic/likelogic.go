package logic

import (
	"context"
	"encoding/json"
	"fansX/internal/model/mq"
	"fansX/internal/util"
	"github.com/IBM/sarama"
	"strconv"
	"time"

	"fansX/services/like/internal/svc"
	"fansX/services/like/proto/likeRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeLogic {
	return &LikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LikeLogic) Like(in *likeRpc.LikeReq) (*likeRpc.Empty, error) {
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	timeout, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	logger.Info("user like", "business", in.BusinessId, "userid", in.UserId, "likeId", in.LikeId, "timestamp", in.TimeStamp)
	msg := mq.Like{
		TimeStamp: in.TimeStamp,
		Business:  in.BusinessId,
		UserId:    in.UserId,
		LikeId:    in.LikeId,
		Cancel:    false,
	}
	value, err := json.Marshal(msg)
	if err != nil {
		logger.Info("marshal json:" + err.Error())
		return nil, err
	}
	message := sarama.ProducerMessage{
		Topic: "",
		Key:   sarama.StringEncoder(strconv.FormatInt(in.LikeId, 10)),
		Value: sarama.ByteEncoder(value),
	}
	i := 0
	for {
		select {
		case <-timeout.Done():
			logger.Error("send message time out")
			return nil, context.DeadlineExceeded
		default:
			i++
			_, _, err = l.svcCtx.Producer.SendMessage(&message)
			if err != nil {
				logger.Error("send message to kafka:"+err.Error(), "times", i)
				time.Sleep(time.Millisecond * 100)
				continue
			}
		}
		break
	}

	return &likeRpc.Empty{}, nil
}
