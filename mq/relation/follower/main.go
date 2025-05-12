package main

import (
	"context"
	"fansX/internal/middleware/lua"
	"fansX/mq/relation/script"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "1jian10.cn:6379",
		DB:   0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err.Error())
	}
	executor := lua.NewExecutor(client)
	_, err := executor.Load(context.Background(), []*lua.Script{
		script.IncrBy,
		script.InsertZSet,
		script.RemoveZSet,
		script.InsertZSetWithMa,
	})
	if err != nil {
		panic(err.Error())
	}

	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test_followers_group", config)
	handler := Handler{
		client: client,
	}

	err = consumer.Consume(context.Background(), []string{"test_relation_followers"}, &handler)
	if err != nil {
		panic(err.Error())
	}

}
