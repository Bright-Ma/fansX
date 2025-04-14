package main

import (
	"context"
	"fansX/common/lua"
	luaHash "fansX/common/lua/script/hash"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
)

func main() {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.AutoCommit.Enable = false

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err.Error())
	}

	cache, err := Init(client)
	if err != nil {
		panic(err.Error())
	}

	executor := lua.NewExecutor(client)
	index, err := executor.Load(context.Background(), []lua.Script{luaHash.GetCreate()})
	if err != nil {
		panic(fmt.Sprintln(err.Error(), " index:", index))
	}

	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test_feed_group", config)
	handler := Handler{
		client:   client,
		cache:    cache,
		executor: executor,
	}

	err = consumer.Consume(context.Background(), []string{"test_feed"}, &handler)
	if err != nil {
		panic(err.Error())
	}

}
