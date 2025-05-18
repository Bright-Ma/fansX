package main

import (
	"context"
	bigcache "fansX/internal/middleware/cache"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	etcd "go.etcd.io/etcd/client/v3"
	"time"
)

func main() {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.AutoCommit.Enable = false

	client := redis.NewClient(&redis.Options{
		Addr: "1jian10.cn:6379",
		DB:   0,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err.Error())
	}

	eClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{"1jian10.cn:4379"},
		DialTimeout: time.Second * 3,
	})
	if err != nil {
		panic(err.Error())
	}
	creator := bigcache.NewCacheCreator(eClient)

	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test_feed_group", config)
	handler := Handler{
		client:       client,
		cacheCreator: creator,
	}

	err = consumer.Consume(context.Background(), []string{"test_feed"}, &handler)
	if err != nil {
		panic(err.Error())
	}

}
