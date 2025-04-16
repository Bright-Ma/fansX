package main

import (
	"context"
	lua2 "fansX/internal/middleware/lua"
	interlua "fansX/mq/relation/following/lua"
	"fansX/pkg/hotkey-go/hotkey"
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
		DB:   1,
	})
	e := lua2.NewExecutor(client)
	_, err = e.Load(context.Background(), []lua2.Script{interlua.GetDel(), interlua.GetAdd()})
	if err != nil {
		panic(err.Error())
	}

	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false

	core, err := hotkey.NewCore(hotkey.Config{
		Model:        hotkey.ModelConsumer,
		GroupName:    "",
		DelGroupName: "",
		CacheSize:    1024 * 1024 * 64,
		HotKeySize:   1024 * 1024 * 64,
		EtcdConfig: etcd.Config{
			Endpoints:   []string{"1jian10.cn:4379"},
			DialTimeout: time.Second * 3,
		},
	})
	if err != nil {
		panic(err.Error())
	}
	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test-group", config)
	handler := Handler{
		db:       db,
		client:   client,
		executor: e,
		core:     core,
	}

	err = consumer.Consume(context.Background(), []string{"topic-test"}, &handler)
	if err != nil {
		panic(err.Error())
	}
}
