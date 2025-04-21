package main

import (
	"fansX/internal/middleware/lua"
	"fansX/internal/model/mq"
	leaf "fansX/pkg/leaf-go"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	client   *redis.Client
	executor *lua.Executor
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	h.consumer = NewConsumer(h.db)
	go h.consumer.Consume()
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	h.consumer.Close()
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.Like{}
		err := json.Unmarshal(msg.Value, message)
		if err != nil {
			slog.Error("unmarshal json:" + err.Error())
			continue
		}

		need, err := h.process(message)
		if err == nil {
			session.MarkMessage(msg, "")
			if need {
				h.consumer.Send(message)
			}
		}
	}

	return nil
}
