package svc

import (
	"auth/internal/config"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"time"
)

type ServiceContext struct {
	Config config.Config
	RDB    *redis.Client
	Secret string
}

type JwtClaims struct {
	Userid int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type Session struct {
	Userid int64 `json:"user_id"`
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr: "",
		DB:   1,
	})
	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
	if err := rdb.Ping(timeout).Err(); err != nil {
		panic(err.Error())
	}
	cancel()

	return &ServiceContext{
		Config: c,
		RDB:    rdb,
	}
}
