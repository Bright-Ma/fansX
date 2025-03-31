package svc

import (
	"context"
	"github.com/redis/go-redis/v9"
	etcd "go.etcd.io/etcd/client/v3"
	"puhser/internal/config"
	"time"
)

// Context 程序运行的上下文，用于保存配置，redis连接等
type Context struct {
	Config  config.Config
	RDB     *redis.Client
	EClient *etcd.Client
}

func NewContext(file string) *Context {
	c := config.ReadConfig(file)
	rdb := redis.NewClient(&redis.Options{
		Addr: c.Redis.Addr,
		DB:   c.Redis.DB,
	})
	EClient, err := etcd.New(etcd.Config{
		Endpoints:   c.Etcd.EndPoints,
		DialTimeout: time.Duration(c.Etcd.DialTimeout) * time.Second,
	})

	if err != nil {
		panic(err)
	}
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err.Error())
	}

	return &Context{
		Config:  c,
		RDB:     rdb,
		EClient: EClient,
	}
}
