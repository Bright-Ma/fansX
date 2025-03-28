package main

import (
	"bilibili/internal/model/mq"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type Handler struct {
	client *redis.Client
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.FollowingCanalJson{}
		err := json.Unmarshal(msg.Value, message)
		if err != nil {
			slog.Error("unmarshal json:" + err.Error())
			continue
		}
		key1 := "Following:" + message.Data[0].FollowerId
		key2 := "FollowingNums:" + message.Data[0].FollowerId
		h.client.Del(context.Background(), key1, key2)
		session.MarkMessage(msg, "")
	}

	return nil
}

func main() {

	client := redis.NewClient(&redis.Options{
		Addr: "1jian10.cn:6379",
		DB:   1,
	})

	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test-group2", config)
	handler := Handler{
		client: client,
	}

	err := consumer.Consume(context.Background(), []string{"topic-test2"}, &handler)
	if err != nil {
		panic(err.Error())
	}

}
