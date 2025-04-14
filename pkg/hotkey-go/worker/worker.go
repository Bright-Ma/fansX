package main

import (
	"fansX/pkg/hotkey-go/worker/server"
	"fansX/pkg/hotkey-go/worker/service"
	"strconv"
	"time"
)

func main() {
	err := service.RegisterService([]string{"1jian10.cn:4379"}, "1jian10.cn:23030", "worker/"+strconv.FormatInt(time.Now().UnixNano(), 10))

	if err != nil {
		panic(err.Error())
	}

	err = server.Serve("tcp://0.0.0.0:23030")
	if err != nil {
		panic(err.Error())
	}
}
