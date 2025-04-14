package service

import (
	"context"
	"fmt"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"net"
	cshash "puhser/consistenthash"
	svc "puhser/internal/context"
	"puhser/internal/server"
	"puhser/proto/push"
	"puhser/route"
	"strconv"
	"time"
)

func Init(ctx *svc.Context) {
	var err error
	_, err = RegisterService(ctx)
	if err != nil {
		panic(err.Error())
	}

	lis, err := net.Listen("tcp", ctx.Config.Etcd.Addr)
	if err != nil {
		panic(err.Error())
	}

	grpcServer := grpc.NewServer()
	push.RegisterPushMessageServiceServer(grpcServer, server.NewPushMessageServiceServer(ctx))
	fmt.Println("rpc begin listening:" + ctx.Config.Etcd.Addr)
	if ctx.Config.Model == 2 {
		InitHash(ctx)
		go Watch(ctx)
	}
	if err = grpcServer.Serve(lis); err != nil {
		panic(err.Error())
	}

}

func InitHash(ctx *svc.Context) {
	kv := etcd.NewKV(ctx.EClient)
	resp, err := kv.Get(context.Background(), ctx.Config.Etcd.WatchPrefix+"/", etcd.WithPrefix())
	if err != nil {
		panic(err.Error())
	}
	ins := make([]string, 0)
	for _, ev := range resp.Kvs {
		ins = append(ins, string(ev.Value))
	}
	cshash.Update([]string{}, ins)
}

// Watch 获取etcd中其他节点的变化，并对节点进行重新分配
func Watch(ctx *svc.Context) {
	watcher := etcd.NewWatcher(ctx.EClient)
	WatchChan := watcher.Watch(context.Background(), ctx.Config.Etcd.WatchPrefix, etcd.WithPrefix())

	for resp := range WatchChan {
		del := make([]string, 0) //删除节点列表
		ins := make([]string, 0) //增加节点列表
		for _, ev := range resp.Events {
			if ev.Type == etcd.EventTypePut {
				ins = append(ins, string(ev.Kv.Value))
			} else {
				del = append(del, string(ev.Kv.Value))
			}
			cshash.Update(del, ins)
			route.ReConn()
		}
	}
}

// RegisterService 向etcd中注册节点并自动续约
func RegisterService(ctx *svc.Context) (etcd.LeaseID, error) {
	EClient := ctx.EClient
	c := ctx.Config

	//创建租约
	grantResp, err := EClient.Grant(context.Background(), c.Etcd.TTL)
	if err != nil {
		return 0, err
	}

	//使用纳秒级别的时间戳创建key，ip+port为value，其和租约一同过期
	name := c.Etcd.Name + "/" + strconv.FormatInt(time.Now().UnixNano(), 10)
	_, err = EClient.Put(context.Background(), name, c.Etcd.Addr, etcd.WithLease(grantResp.ID))
	if err != nil {
		return 0, err
	}

	//自动续约
	ch, err := EClient.KeepAlive(context.Background(), grantResp.ID)
	if err != nil {
		return 0, err
	}

	//开启一个协程，自动处理续约报文
	go func() {
		for {
			select {
			case resp := <-ch:
				//续约过期，panic，需要重启节点
				if resp == nil {
					panic("grant timeout")
				}
			}
		}
	}()

	return grantResp.ID, nil
}
