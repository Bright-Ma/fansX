package service

import (
	"bilibili/common/middleware/hotkeys/worker/group"
	"context"
	etcd "go.etcd.io/etcd/client/v3"
	"time"
)

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
		//panic("lease time out")
	}()

	return nil
}
