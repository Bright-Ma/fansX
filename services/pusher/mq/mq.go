package mq

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	svc "puhser/internal/context"
	"puhser/route"
)

func Init(ctx *svc.Context) {
	config := nsq.NewConfig()
	// 创建消费者，其中topic固定，而channel随机生成，保证每个节点都会接收到topic中的消息
	c, err := nsq.NewConsumer(ctx.Config.NSQ.Topic, uuid.New().String(), config)
	if err != nil {
		panic(err.Error())
	}

	c.AddHandler(&MessageHandler{ctx: ctx})
	if err := c.ConnectToNSQD(ctx.Config.NSQ.Addr); err != nil {
		panic(err.Error())
	}
}

type MessageHandler struct {
	ctx *svc.Context
}
type Request struct {
	//消息体
	Msg route.Message `json:"message"`
	//用于指定全局消息发送到哪个bucket中
	BucketId int64 `json:"bucketId"`
}

func (h *MessageHandler) HandleMessage(m *nsq.Message) error {
	req := new(Request)
	if err := json.Unmarshal(m.Body, req); err != nil {
		m.Finish()
		return err
	}
	route.SendGlobalMessage(req.BucketId, &req.Msg)
	m.Finish()
	return nil
}
