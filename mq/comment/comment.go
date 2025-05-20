package main

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	rClient := redis.NewClient(&redis.Options{
		Addr: "1jian10.cn:6379",
		DB:   0,
	})
	if err := rClient.Ping(context.Background()).Err(); err != nil {
		panic(err.Error())
	}

	dsn := "root:@tcp(linux.1jian10.cn:4000)/test?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "", config)
	handler := Handler{
		db:     db,
		client: rClient,
	}

	err = consumer.Consume(context.Background(), []string{}, &handler)
	if err != nil {
		panic(err.Error())
	}
}
