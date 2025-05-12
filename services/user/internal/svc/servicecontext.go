package svc

import (
	leaf "fansX/pkg/leaf-go"
	"fansX/services/auth/proto/AuthRpc"
	"fansX/services/user/internal/config"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
	"log/slog"
)

type ServiceContext struct {
	Config     config.Config
	RClient    *redis.Client
	AuthClient AuthRpc.AuthServiceClient
	DB         *gorm.DB
	Creator    leaf.Core
	Logger     *slog.Logger
	Producer   sarama.SyncProducer
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := zrpc.MustNewClient()
	client := AuthRpc.NewAuthServiceClient()
	return &ServiceContext{
		Config: c,
	}
}
