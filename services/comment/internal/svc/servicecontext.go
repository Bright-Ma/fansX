package svc

import (
	"fansX/internal/middleware/lua"
	"fansX/pkg/hotkey-go/hotkey"
	leaf "fansX/pkg/leaf-go"
	"fansX/services/comment/internal/config"
	"github.com/IBM/sarama"
	"github.com/go-redsync/redsync/v4"
	"github.com/golang/groupcache/singleflight"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
)

type ServiceContext struct {
	Config   config.Config
	Client   *redis.Client
	Executor *lua.Executor
	DB       *gorm.DB
	Producer sarama.SyncProducer
	Logger   *slog.Logger
	Creator  leaf.Core
	Cache    *hotkey.Core
	RedSync  *redsync.Redsync
	Group    *singleflight.Group
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
