package svc

import (
	"fansX/services/like/internal/config"
	"github.com/IBM/sarama"
	"log/slog"
)

type ServiceContext struct {
	Config   config.Config
	Producer sarama.SyncProducer
	Logger   *slog.Logger
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
