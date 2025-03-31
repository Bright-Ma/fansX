package svc

import (
	"OutGetWay/config"
	etcd "go.etcd.io/etcd/client/v3"
	"time"
)

type Context struct {
	Config  config.Config
	EClient *etcd.Client
}

func NewContext(file string) *Context {
	c := config.ReadConfig(file)
	EClient, err := etcd.New(etcd.Config{
		Endpoints:   c.Etcd.Endpoints,
		DialTimeout: time.Duration(c.Etcd.DialTimeout) * time.Second,
	})

	if err != nil {
		panic(err)
	}
	return &Context{
		Config:  c,
		EClient: EClient,
	}
}
