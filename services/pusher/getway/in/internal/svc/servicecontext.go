package svc

import (
	"context"
	"github.com/nsqio/go-nsq"
	"github.com/redis/go-redis/v9"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	cshash "puhser/consistenthash"
	"puhser/getway/in/internal/config"
	"sync"
	"time"
)

type ServiceContext struct {
	Config   config.Config
	RDB      *redis.Client
	Services sync.Map
	EClient  *etcd.Client
	Producer *nsq.Producer
}

func NewServiceContext(c config.Config) *ServiceContext {
	EClient, err := etcd.New(etcd.Config{
		Endpoints:   c.Etcd.Hosts,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err.Error())
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: c.MRedis.Host,
		DB:   c.MRedis.DB,
	})
	if err = rdb.Ping(context.Background()).Err(); err != nil {
		panic(err.Error())
	}
	producer, err := nsq.NewProducer(c.NSQ.Addr, nsq.NewConfig())
	if err != nil {
		panic(err.Error())
	}

	svc := &ServiceContext{
		Config:   c,
		EClient:  EClient,
		Producer: producer,
	}
	InitService(svc)
	go Watch(svc)

	return svc
}

func Watch(svc *ServiceContext) {
	watcher := etcd.NewWatcher(svc.EClient)
	WatchChan := watcher.Watch(context.Background(), svc.Config.WatchPrefix, etcd.WithPrefix())
	for resp := range WatchChan {
		del := make([]string, 0)
		ins := make([]string, 0)
		for _, ev := range resp.Events {
			if ev.Type == etcd.EventTypePut {
				ConnService(svc, string(ev.Kv.Value))
				ins = append(ins, string(ev.Kv.Value))
			} else {
				del = append(del, string(ev.Kv.Value))
				value, ok := svc.Services.Load(ev.Kv.Value)
				if ok {
					c, _ := value.(*grpc.ClientConn)
					_ = c.Close()
				}
				svc.Services.Delete(string(ev.Kv.Value))
			}
		}
		cshash.Update(del, ins)
	}
}

func InitService(svc *ServiceContext) {
	cshash.Init(svc.Config.VirtualNums)
	kv := etcd.NewKV(svc.EClient)
	resp, err := kv.Get(context.Background(), svc.Config.WatchPrefix, etcd.WithPrefix())
	if err != nil {
		panic(err)
	}
	add := make([]string, 0)
	for _, v := range resp.Kvs {
		add = append(add, string(v.Key))
		ConnService(svc, string(v.Value))
	}
	cshash.Update([]string{}, add)
}

func ConnService(svc *ServiceContext, addr string) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return
	}
	svc.Services.Store(addr, conn)
	return
}
