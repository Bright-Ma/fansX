package main

import (
	"bilibili/common/lua"
	"bilibili/mq/relation/follower/lua"
	"context"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:@tcp(linux.1jian10.cn:4000)/test?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	client := redis.NewClient(&redis.Options{
		Addr: "1jian10.cn:6379",
		DB:   1,
	})
	e := lua.NewExecutor(client)

	if err := e.LoadAll(); err != nil {
		panic(err.Error())
	}
	_, err = e.Load(context.Background(), []lua.Script{luaFollowing.GetAdd()})
	if err != nil {
		panic(err.Error())
	}

	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test-group", config)
	handler := Handler{
		db:       db,
		client:   client,
		executor: e,
	}

	err = consumer.Consume(context.Background(), []string{"topic-test"}, &handler)
	if err != nil {
		panic(err.Error())
	}
}
