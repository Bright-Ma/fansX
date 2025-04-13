package svc

import (
	"bilibili/common/lua"
	"bilibili/common/util"
	bigcache "bilibili/internal/cache"
	"bilibili/services/content/public/proto/publicContentRpc"
	"bilibili/services/feed/internal/config"
	"bilibili/services/relation/proto/relationRpc"
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log/slog"
)

type ServiceContext struct {
	Config         config.Config
	RClient        *redis.Client
	DB             *gorm.DB
	Logger         *slog.Logger
	Executor       *lua.Executor
	Cache          *bigcache.Cache
	RelationClient relationRpc.RelationServiceClient
	ContentClient  publicContentRpc.PublicContentServiceClient
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

	if err := r.Ping(context.Background()).Err(); err != nil {
		panic(err.Error())
	}

	logger, err := util.InitLog("FeedService", slog.LevelDebug)
	if err != nil {
		panic(err.Error())
	}

	e := lua.NewExecutor(r)
	if err = e.LoadAll(); err != nil {
		panic(err.Error())
	}

	cache, err := bigcache.Init(r)
	if err != nil {
		panic(err.Error())
	}

	conn := zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: []string{"127.0.0.1:4379"},
			Key:   "relation.rpc",
		},
	})
	relationClient := relationRpc.NewRelationServiceClient(conn.Conn())

	svc := &ServiceContext{
		Config:         c,
		DB:             db,
		RClient:        r,
		Logger:         logger,
		Executor:       e,
		Cache:          cache,
		RelationClient: relationClient,
	}

	return svc
}
