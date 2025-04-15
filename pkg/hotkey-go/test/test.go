package main

import (
	"fansX/pkg/hotkey-go/hotkey"
	etcd "go.etcd.io/etcd/client/v3"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	key := make([]string, 500000)
	for i := 0; i < len(key); i++ {
		key[i] = "key" + strconv.FormatInt(int64(i), 10)
	}

	for i := 0; i < 4; i++ {
		config := hotkey.Config{
			Model:      hotkey.ModelCache,
			GroupName:  "test" + strconv.Itoa(i),
			CacheSize:  1024 * 1024 * 1024,
			HotKeySize: 1024 * 1024 * 128,
			EtcdConfig: etcd.Config{
				Endpoints:   []string{"127.0.0.1:4379"},
				DialTimeout: time.Second * 3,
			},
			DelChan: nil,
			HotChan: nil,
		}
		core, err := hotkey.NewCore(config)
		if err != nil {
			panic(err.Error())
		}

		go func() {
			for i := 0; i < 1; i++ {
				go func() {
					for {
						core.Get(key[rand.Intn(len(key))])
					}
				}()
			}
		}()

	}
	select {}
}
