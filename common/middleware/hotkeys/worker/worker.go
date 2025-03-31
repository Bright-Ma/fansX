package main

import (
	"bilibili/common/middleware/hotkeys/worker/server"
	"bilibili/common/middleware/hotkeys/worker/service"
	"strconv"
	"time"
)

func main() {
	err := service.RegisterService([]string{"1jian10.cn:4379"}, "1jian10.cn:23310", "hotkeys/"+strconv.FormatInt(time.Now().UnixNano(), 10))
	if err != nil {
		panic(err.Error())
	}

	err = server.Serve("tcp://0.0.0.0:23310")
	if err != nil {
		panic(err.Error())
	}
}
