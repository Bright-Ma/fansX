package svc

import (
	"context"
	"fansX/internal/middleware/lua"
	"fansX/internal/util"
	"fansX/pkg/hotkey-go/hotkey"
	"fansX/services/content/public/internal/config"
	"fansX/services/content/public/internal/script"
	"github.com/redis/go-redis/v9"
	etcd "go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log/slog"
	"time"
)

type ServiceContext struct {
	Config   config.Config
	Core     *hotkey.Core
	DB       *gorm.DB
	RClient  *redis.Client
	Executor *lua.Executor
	Logger   *slog.Logger
}

func NewServiceContext(c config.Config) *ServiceContext {
	dsn := "root:@tcp(linux.1jian10.cn:4000)/test?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	client := redis.NewClient(&redis.Options{
		Addr: "1jian10.cn:6379",
		DB:   0,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err.Error())
	}

	logger, err := util.InitLog("PublicContent.rpc", slog.LevelDebug)
	if err != nil {
		panic(err.Error())
	}

	e := lua.NewExecutor(client)
	_, err = e.Load(context.Background(), []*lua.Script{script.BuildZSet})
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
	core, err := hotkey.NewCore("", eClient,
		hotkey.WithCacheSize(1024*1024*1024),
		hotkey.WithChannelSize(1024*64),
	)
	if err != nil {
		panic(err.Error())
	}

	return &ServiceContext{
		Config:   c,
		Core:     core,
		RClient:  client,
		DB:       db,
		Executor: e,
		Logger:   logger,
	}
}
