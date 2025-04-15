package main

import (
	"context"
	"fansX/common/lua"
	"fansX/common/util"
	bigcache "fansX/internal/middleware/cache"
	interlua "fansX/mq/content/public/lua"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log/slog"
)

func main() {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.AutoCommit.Enable = false

	dsn := "root:@tcp(linux.1jian10.cn:4000)/test?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	logger, err := util.InitLog("mq:PublicContent", slog.LevelDebug)
	if err != nil {
		panic(err.Error())
	}
	slog.SetDefault(logger)

	client := redis.NewClient(&redis.Options{
		Addr: "1jian10.cn:6379",
		DB:   0,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err.Error())
	}

	cache, err := bigcache.Init(client)
	if err != nil {
		panic(err.Error())
	}

	executor := lua.NewExecutor(client)
	_, err = executor.Load(context.Background(), []lua.Script{interlua.GetAdd()})
	if err != nil {
		panic(err.Error())
	}

	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test_public_content_group", config)
	handler := Handler{
		db:       db,
		client:   client,
		bigCache: cache,
	}

	err = consumer.Consume(context.Background(), []string{"test_public_content"}, &handler)
	if err != nil {
		panic(err.Error())
	}

}
