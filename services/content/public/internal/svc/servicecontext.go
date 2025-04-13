package svc

import (
	"bilibili/common/lua"
	"bilibili/common/util"
	"bilibili/pkg/hotkey-go/hotkey"
	"bilibili/services/content/public/internal/config"
	"context"
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
	if err := e.LoadAll(); err != nil {
		panic(err.Error())
	}

	del := make(chan string, 1024*64)
	core, err := hotkey.NewCore(hotkey.Config{
		Model:      hotkey.ModelCache,
		GroupName:  "PublicContent.rpc",
		CacheSize:  1024 * 1024 * 512,
		HotKeySize: 1024 * 1024 * 64,
		EtcdConfig: etcd.Config{
			Endpoints:   []string{"127.0.0.1:4379"},
			DialTimeout: time.Second * 3,
		},
		DelChan: del,
		HotChan: nil,
	})
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
