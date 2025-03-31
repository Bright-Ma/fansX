package service

import (
	svc "OutGetWay/context"
	"context"
	etcd "go.etcd.io/etcd/client/v3"
	"math/rand"
	"sync"
)

// OldAddr 存储旧pusher节点地址的列表，实时性不强
var OldAddr []string

// NewAddr 用于更新时的中间存储
var NewAddr []string

// Addr 存储节点的最新地址
var Addr = make(map[string]string)

// mu 读写锁，更新时+写锁，否则+读锁
var mu sync.RWMutex

// Model 使用哪种路由模式，随机或者一致性hash
var Model int

// Watch 观测etcd中键值对的变化，对本地地址进行更新
func Watch(ctx *svc.Context) {
	watcher := etcd.NewWatcher(ctx.EClient)
	//根据前缀对key进行watch，返回值为一个channel，当有变化时，将会发送消息到该chan中
	WatchChan := watcher.Watch(context.Background(), ctx.Config.Etcd.WatchPrefix, etcd.WithPrefix())
	if Model == 1 {
		randomWatch(WatchChan)
	} else {
		consistentHashWatch(WatchChan)
	}
}

// InitService 初始化pusher节点列表
func InitService(ctx *svc.Context) {
	Model = ctx.Config.Model
	kv := etcd.NewKV(ctx.EClient)
	resp, err := kv.Get(context.Background(), ctx.Config.Etcd.WatchPrefix, etcd.WithPrefix())
	if err != nil {
		panic(err)
	}

	if Model == 1 {
		randomInit(resp)
	} else {
		consistentHashInit(resp)
	}
	go Watch(ctx)
}

// SelectService 随机路由，选取一个随机的路由地址
func SelectService() string {
	mu.RLock()
	defer mu.RUnlock()
	return OldAddr[rand.Intn(len(OldAddr))]
}
