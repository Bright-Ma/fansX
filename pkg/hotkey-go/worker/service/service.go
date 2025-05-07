package service

import (
	"context"
	"fansX/pkg/hotkey-go/worker/config"
	"fansX/pkg/hotkey-go/worker/group"
	"fmt"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	etcd "go.etcd.io/etcd/client/v3"
	"log/slog"
	"strings"
	"time"
)

// RegisterService 将worker节点注册到etcd，同时监听配置的变化，host为本机ip+监听的端口号
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

	getResp, err := client.Get(timeout, "group/", etcd.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range getResp.Kvs {
		cf, err := ReadConfig(etcdAddr[0], string(kv.Key))
		if err != nil {
			return err

		}
		group.GetGroupMap().Update(cf)
	}

	go watchGroup(client, getResp.Header.Revision, etcdAddr[0])

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

// watchGroup 监听配置文件的变化
func watchGroup(client *etcd.Client, rev int64, addr string) {
	watch := client.Watch(context.Background(), "group/", etcd.WithRev(rev), etcd.WithPrefix())
	defer func() {
		if err := recover(); err != nil {
			slog.Error("watchGroup panic:" + fmt.Sprint(err))
			slog.Error("please try to restart worker")
		}
	}()
	for w := range watch {
		for _, ev := range w.Events {
			// 配置删除
			if ev.Type == etcd.EventTypeDelete {
				str := string(ev.Kv.Key)
				name, _ := strings.CutPrefix("group/", str)
				group.GetGroupMap().Delete(name)
				// 配置更新或新增
			} else if ev.Type == etcd.EventTypePut {
				cf, err := ReadConfig(addr, string(ev.Kv.Key))
				if err != nil {
					slog.Error("read config:" + err.Error())
					continue
				}
				group.GetGroupMap().Update(cf)
			} else {
				slog.Error("unKnow etcd.eventType")
			}
		}

	}
	panic("watch group time out")

}

// ReadConfig 在etcd中读取配置
func ReadConfig(addr string, path string) (config.Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.AddRemoteProvider("etcd3", addr, path)
	if err != nil {
		return config.Config{}, err
	}
	err = v.ReadRemoteConfig()
	if err != nil {
		return config.Config{}, err
	}
	cf := config.Config{}
	err = v.Unmarshal(&cf)
	if err != nil {
		return config.Config{}, err
	}
	return cf, nil
}
