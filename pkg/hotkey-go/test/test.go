package main

import (
	"fansX/pkg/hotkey-go/hotkey"
	etcd "go.etcd.io/etcd/client/v3"
	"log/slog"
	"math/rand"
	"strconv"
	"time"
)

type Observer struct {
}

func (ob *Observer) Do(key string) {
	slog.Info("key: " + key + " to hot key")
}

func main() {
	key := make([]string, 1)
	for i := 0; i < len(key); i++ {
		key[i] = "key" + strconv.FormatInt(int64(i), 10)
	}
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{"1jian10.cn:4379"},
		DialTimeout: time.Second * 3,
	})
	if err != nil {
		panic(err.Error())
	}
	core, err := hotkey.NewCore("test", client, hotkey.WithObserver(&Observer{}))
	for {
		k := key[rand.Intn(len(key))]
		core.Get(k)
		if core.IsHotKey(k) {
			slog.Info("hotkey:", k)
		}
		time.Sleep(time.Millisecond * 9)
	}

}
