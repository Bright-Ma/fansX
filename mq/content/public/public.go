package main

import (
	"context"
	bigcache "fansX/internal/middleware/cache"
	"fansX/internal/middleware/lua"
	"fansX/internal/util"
	"fansX/mq/content/script"
	leaf "fansX/pkg/leaf-go"
	"fansX/services/relation/proto/relationRpc"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
	etcd "go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log/slog"
	"time"
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

	logger, err := util.InitLog("publiccontent.kafka", slog.LevelDebug)
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
	eClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{"1jian10.cn:4379"},
		DialTimeout: time.Second * 3,
	})
	if err != nil {
		panic(err.Error())
	}
	cache := bigcache.NewCache(eClient)

	executor := lua.NewExecutor(client)
	_, err = executor.Load(context.Background(), []*lua.Script{script.AddZSet})
	if err != nil {
		panic(err.Error())
	}

	creator, err := leaf.NewCore(leaf.Config{
		Model: leaf.Snowflake,
		SnowflakeConfig: &leaf.SnowflakeConfig{
			CreatorName: "publiccontent.kafka",
			Addr:        "1jian10.cn:20000", //未生效
			EtcdAddr:    []string{"1jian10.cn:4379"},
		},
	})
	if err != nil {
		panic(err.Error())
	}

	zClient := zrpc.MustNewClient(zrpc.NewEtcdClientConf([]string{"1jian10.cn:4379"}, "relation.rpc", "", ""))
	relationClient := relationRpc.NewRelationServiceClient(zClient.Conn())

	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test_content_public_group", config)
	handler := Handler{
		db:             db,
		client:         client,
		bigCache:       cache,
		creator:        creator,
		executor:       executor,
		relationClient: relationClient,
	}

	err = consumer.Consume(context.Background(), []string{"test_content_public"}, &handler)
	if err != nil {
		panic(err.Error())
	}

}
