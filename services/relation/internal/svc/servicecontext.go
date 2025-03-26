package svc

import (
	"bilibili/common/lua"
	luaHash "bilibili/common/lua/script/hash"
	leaf "bilibili/common/middleware/leaf-go"
	"bilibili/common/middleware/leaf-go/snowflake"
	"bilibili/common/util"
	"bilibili/services/relation/internal/config"
	"context"
	"github.com/golang/groupcache/singleflight"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log/slog"
	"strconv"
)

type ServiceContext struct {
	Config   config.Config
	DB       *gorm.DB
	RClient  *redis.Client
	Creator  leaf.Core
	Logger   *slog.Logger
	Single   *singleflight.Group
	Executor *lua.Executor
}

func NewServiceContext(c config.Config) *ServiceContext {
	dsn := "root:@tcp(linux.1jian10.cn:4000)/relation?charset=utf8mb4&parseTime=True"
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
		SnowflakeConfig: &snowflake.Config{
			CreatorName: "relation",
			Addr:        "addr",
			EtcdAddr:    []string{"127.0.0.1:4379"},
		},
	})
	if err != nil {
		panic(err.Error())
	}
	logger, err := util.InitLog("RelationService", slog.LevelInfo)
	if err != nil {
		panic(err.Error())
	}
	e := lua.NewExecutor(r)
	i, err := e.Load(context.Background(), []lua.Script{luaHash.GetGetField(), luaHash.GetCreate()})
	if err != nil {
		panic("load panic,index:" + strconv.Itoa(i) + "," + err.Error())
	}
	return &ServiceContext{
		Config:   c,
		DB:       db,
		RClient:  r,
		Creator:  creator,
		Logger:   logger,
		Executor: e,
	}
}
