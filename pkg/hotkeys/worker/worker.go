package main

import (
	"bilibili/pkg/hotkeys/worker/group"
	"bilibili/pkg/hotkeys/worker/server"
	"bilibili/pkg/hotkeys/worker/service"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"
)

func main() {
	err := service.RegisterService([]string{"1jian10.cn:4379"}, "1jian10.cn:23310", "worker/"+strconv.FormatInt(time.Now().UnixNano(), 10))
	if err != nil {
		panic(err.Error())
	}
	go func() {
		for {
			time.Sleep(1 * time.Second)
			fmt.Println(atomic.LoadUint32(&group.Nums))
			atomic.StoreUint32(&group.Nums, 0)
		}
	}()
	err = server.Serve("tcp://0.0.0.0:23310")
	if err != nil {
		panic(err.Error())
	}
}
