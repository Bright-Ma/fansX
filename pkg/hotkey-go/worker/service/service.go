package service

import (
	"bilibili/pkg/hotkey-go/worker/group"
	"context"
	"fmt"
	etcd "go.etcd.io/etcd/client/v3"
	"log/slog"
	"time"
)

// RegisterService 将worker节点注册到etcd，同时监听配置的变化(目前只提供group的变化)，host为本机ip+监听的端口号
func RegisterService(etcdAddr []string, Host string, key string) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	client, err := etcd.New(etcd.Config{
		Endpoints:   etcdAddr,
		DialTimeout: time.Second * 3,
	})
	if err != nil {
		return err
	}

	getResp, err := client.Get(timeout, "group", etcd.WithPrefix())
	if err != nil {
		return err
	}

	for _, v := range getResp.Kvs {
		g := group.NewGroup()
		group.GetGroupMap().Set(string(v.Value), g)
	}

	go watchGroup(client, getResp.Header.Revision)

	leaseResp, err := client.Grant(context.Background(), 10)
	if err != nil {
		return err
	}

	_, err = client.Put(timeout, key, Host, etcd.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	keepResp, err := client.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}

	go func() {
		for range keepResp {
		}
		panic("lease time out")
	}()

	return nil
}

// watchGroup 监听group的变化，未来考虑加入其他配置文件
func watchGroup(client *etcd.Client, rev int64) {
	watch := client.Watch(context.Background(), "group/", etcd.WithRev(rev))
	defer func() {
		if err := recover(); err != nil {
			slog.Error("watchGroup panic:" + fmt.Sprint(err))
			slog.Error("please try to restart worker")
		}
	}()
	for w := range watch {
		for _, ev := range w.Events {
			if ev.Type == etcd.EventTypeDelete {
				continue
			} else if ev.Type == etcd.EventTypePut {
				g := group.NewGroup()
				group.GetGroupMap().Set(string(ev.Kv.Value), g)
			} else {
				slog.Error("unKnow etcd.eventType")
			}
		}

	}
	panic("watch group time out")

}
