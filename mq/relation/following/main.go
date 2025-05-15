package main

import (
	"context"
	bigcache "fansX/internal/middleware/cache"
	"fansX/internal/middleware/lua"
	"fansX/mq/relation/script"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	etcd "go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func main() {
	// TODO update config
	dsn := "root:@tcp(linux.1jian10.cn:4000)/test?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	client := redis.NewClient(&redis.Options{
		Addr: "1jian10.cn:6379",
		DB:   0,
	})
	e := lua.NewExecutor(client)
	_, err = e.Load(context.Background(), []*lua.Script{
		script.IncrBy,
		script.InsertZSet,
		script.InsertZSetWithMa,
		script.RemoveZSet,
	})
	if err != nil {
		panic(err.Error())
	}

	eClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{"1jian10.cn:4379"},
		DialTimeout: time.Second * 3,
	})
	if err != nil {
		panic(err.Error())
	}

	cache := bigcache.NewCache(eClient)

	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false

	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test_followings_group", config)
	handler := Handler{
		db:       db,
		client:   client,
		executor: e,
		bigCache: cache,
	}

	err = consumer.Consume(context.Background(), []string{"test_relation_followings"}, &handler)
	if err != nil {
		panic(err.Error())
	}
}
