package main

import (
	"context"
	bigcache "fansX/internal/middleware/cache"
	"fansX/internal/middleware/lua"
	"fansX/mq/feed/script"
	"fansX/services/content/public/proto/publicContentRpc"
	"fansX/services/relation/proto/relationRpc"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
	etcd "go.etcd.io/etcd/client/v3"
	"time"
)

func main() {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.AutoCommit.Enable = false

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
	creator := bigcache.NewCacheCreator(eClient)

	executor := lua.NewExecutor(client)
	_, err = executor.Load(context.Background(), []*lua.Script{script.ZSetAdd})
	if err != nil {
		panic(err.Error())
	}

	zClient := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{"1jian10.cn:4379"},
	})
	publicContentClient := publicContentRpc.NewPublicContentServiceClient(zClient.Conn())
	relationClient := relationRpc.NewRelationServiceClient(zClient.Conn())

	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test_feed_group", config)
	handler := Handler{
		client:              client,
		cacheCreator:        creator,
		executor:            executor,
		publicContentClient: publicContentClient,
		relationClient:      relationClient,
	}

	err = consumer.Consume(context.Background(), []string{"test_feed"}, &handler)
	if err != nil {
		panic(err.Error())
	}

}
