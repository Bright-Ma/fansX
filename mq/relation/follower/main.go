package main

import (
	"context"
	lua2 "fansX/internal/middleware/lua"
	interlua "fansX/mq/relation/follower/lua"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
)

func main() {
	// TODO update config
	client := redis.NewClient(&redis.Options{
		Addr: "1jian10.cn:6379",
		DB:   0,
	})

	if err := client.Ping(context.Background()); err != nil {
		panic(err)
	}
	executor := lua2.NewExecutor(client)
	err := executor.LoadAll()
	if err != nil {
		panic(err.Error())
	}
	_, err = executor.Load(context.Background(), []lua2.Script{interlua.GetAdd()})
	if err != nil {
		panic(err.Error())
	}

	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test-group2", config)
	handler := Handler{
		client: client,
	}

	err = consumer.Consume(context.Background(), []string{"topic-test2"}, &handler)
	if err != nil {
		panic(err.Error())
	}

}
