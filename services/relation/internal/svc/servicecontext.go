package svc

import (
	"fansX/common/lua"
	"fansX/common/util"
	"fansX/pkg/hotkey-go/hotkey"
	leaf "fansX/pkg/leaf-go"
	"fansX/services/relation/internal/config"
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

	creator, err := leaf.Init(&leaf.Config{
		Model: leaf.Snowflake,
		SnowflakeConfig: &leaf.SnowflakeConfig{
			CreatorName: "relation.rpc",
			Addr:        "1jian10.cn:23010",
			EtcdAddr:    []string{"127.0.0.1:4379"},
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
	if err := e.LoadAll(); err != nil {
		panic(err.Error())
	}

	del := make(chan string, 1024*64)

	core, err := hotkey.NewCore(hotkey.Config{
		Model:      hotkey.ModelCache,
		GroupName:  "relation.rpc",
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

	go func() {
		for key := range del {
			core.Del(key)
		}
	}()

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
