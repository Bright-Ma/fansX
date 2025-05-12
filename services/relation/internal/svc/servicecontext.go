package svc

import (
	"context"
	"fansX/internal/middleware/lua"
	"fansX/internal/util"
	"fansX/pkg/hotkey-go/hotkey"
	leaf "fansX/pkg/leaf-go"
	"fansX/services/relation/internal/config"
	"fansX/services/relation/internal/script"
	"github.com/golang/groupcache/singleflight"
	"github.com/redis/go-redis/v9"
	etcd "go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log/slog"
)

type ServiceContext struct {
	Config   config.Config
	DB       *gorm.DB
	RClient  *redis.Client
	Creator  leaf.Core
	Logger   *slog.Logger
	Single   *singleflight.Group
	Executor *lua.Executor
	HotKey   *hotkey.Core
}

func NewServiceContext(c config.Config) *ServiceContext {
	dsn := "root:@tcp(linux.1jian10.cn:4000)/test?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	r := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   1,
	})
	creator, err := leaf.NewCore(leaf.Config{
		Model: leaf.Snowflake,
		SnowflakeConfig: &leaf.SnowflakeConfig{
			CreatorName: "relation.rpc",
			Addr:        "1jian10.cn:23010",
			EtcdAddr:    []string{"1jian10.cn:4379"},
		},
	})
	if err != nil {
		panic(err.Error())
	}

	logger, err := util.InitLog("relation.rpc", slog.LevelDebug)
	if err != nil {
		panic(err.Error())
	}

	e := lua.NewExecutor(r)
	_, err = e.Load(context.Background(), []*lua.Script{
		script.BuildZSet,
		script.RevRangeZSet,
		script.GetFiled,
	})
	if err != nil {
		panic(err.Error())
	}
	eClient, err := etcd.New(etcd.Config{Endpoints: []string{"1jian10.cn:4379"}})
	if err != nil {
		panic(err.Error())
	}

	core, err := hotkey.NewCore("relation.rpc", eClient, hotkey.WithCacheSize(1024*1024*1024))

	if err != nil {
		panic(err.Error())
	}

	svc := &ServiceContext{
		Config:   c,
		DB:       db,
		RClient:  r,
		Creator:  creator,
		Logger:   logger,
		Executor: e,
		Single:   &singleflight.Group{},
		HotKey:   core,
	}

	return svc
}
