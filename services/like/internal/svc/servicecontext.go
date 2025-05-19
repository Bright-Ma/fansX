package svc

import (
	"context"
	"fansX/internal/middleware/lua"
	"fansX/internal/util"
	"fansX/pkg/hotkey-go/hotkey"
	"fansX/services/like/internal/config"
	"fansX/services/like/internal/script"
	"github.com/IBM/sarama"
	"github.com/golang/groupcache/singleflight"
	"github.com/redis/go-redis/v9"
	etcd "go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log/slog"
	"time"
)

type ServiceContext struct {
	Config   config.Config
	Producer sarama.SyncProducer
	Logger   *slog.Logger
	Client   *redis.Client
	Cache    *hotkey.Core
	DB       *gorm.DB
	Group    *singleflight.Group
	Executor *lua.Executor
}

func NewServiceContext(c config.Config) *ServiceContext {
	dsn := "root:@tcp(linux.1jian10.cn:4000)/test?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	rClient := redis.NewClient(&redis.Options{
		Addr: "1jian10.cn:6379",
		DB:   0,
	})
	if err := rClient.Ping(context.Background()).Err(); err != nil {
		panic(err.Error())
	}
	eClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{"1jian10.cn:4379"},
		DialTimeout: time.Second * 3,
	})
	cache, err := hotkey.NewCore("like.rpc", eClient,
		hotkey.WithCacheSize(1024*1024*1024),
		hotkey.WithChannelSize(1024*32),
	)
	if err != nil {
		panic(err.Error())
	}
	executor := lua.NewExecutor(rClient)
	_, err = executor.Load(context.Background(), []*lua.Script{
		script.List,
		script.Set,
		script.BuildList,
	})
	if err != nil {
		panic(err.Error())
	}

	logger, err := util.InitLog("like.rpc", slog.LevelDebug)
	if err != nil {
		panic(err.Error())
	}
	return &ServiceContext{
		Config:   c,
		Producer: nil,
		Logger:   logger,
		Client:   rClient,
		Cache:    cache,
		DB:       db,
		Executor: executor,
	}
}
