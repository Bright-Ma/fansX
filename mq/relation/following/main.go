package main

import (
	"context"
	"encoding/json"
	"fansX/internal/model/mq"
	"fansX/pkg/hotkey-go/hotkey"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	etcd "go.etcd.io/etcd/client/v3"
	"log/slog"
	"time"
)

type Handler struct {
	client *redis.Client
	core   *hotkey.Core
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
		h.core.SendDel(key2)
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
	core, err := hotkey.NewCore(hotkey.Config{
		Model:        hotkey.ModelConsumer,
		GroupName:    "relation.rpc_consumer",
		DelGroupName: "relation.rpc",
		CacheSize:    1024 * 32,
		HotKeySize:   1024 * 32,
		EtcdConfig: etcd.Config{
			Endpoints:   []string{"http://127.0.0.1:4379"},
			DialTimeout: time.Second * 3,
		},
		DelChan: nil,
		HotChan: nil,
	})
	if err != nil {
		panic(err.Error())
	}
	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test-group2", config)
	handler := Handler{
		client: client,
		core:   core,
	}

	err = consumer.Consume(context.Background(), []string{"topic-test2"}, &handler)
	if err != nil {
		panic(err.Error())
	}

}
