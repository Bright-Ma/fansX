package svc

import (
	"fansX/internal/middleware/lua"
	"fansX/pkg/hotkey-go/hotkey"
	"fansX/services/like/internal/config"
	"github.com/IBM/sarama"
	"github.com/golang/groupcache/singleflight"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
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
	return &ServiceContext{
		Config: c,
	}
}
