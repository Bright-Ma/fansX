package main

import (
	"context"
	"fansX/internal/middleware/lua"
	"fansX/mq/comment/script"
	leaf "fansX/pkg/leaf-go"
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

	creator, err := leaf.NewCore(leaf.Config{
		Model: leaf.Snowflake,
		SnowflakeConfig: &leaf.SnowflakeConfig{
			CreatorName: "",
			Addr:        "",
			EtcdAddr:    nil,
		},
	})
	if err != nil {
		panic(err.Error())
	}
	executor := lua.NewExecutor(rClient)
	_, err = executor.Load(context.Background(), []*lua.Script{
		script.Add,
		script.Insert,
	})
	if err != nil {
		panic(err.Error())
	}

	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	kafkaConsumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "", config)
	handler := Handler{
		db:       db,
		client:   rClient,
		creator:  creator,
		executor: executor,
	}

	err = kafkaConsumer.Consume(context.Background(), []string{}, &handler)
	if err != nil {
		panic(err.Error())
	}
}
