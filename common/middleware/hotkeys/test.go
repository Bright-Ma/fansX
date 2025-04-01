package main

import (
	"bilibili/common/middleware/hotkeys/hotkey"
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	core, err := hotkey.NewCore(1024*1024*1024, []string{"1jian10.cn:4379"}, "test")
	if err != nil {
		panic(err.Error())
	}
	key := make([]string, 25)
	for i := 0; i < 25; i++ {
		key[i] = strconv.FormatInt(int64(i), 10)
	}

	for i := 0; i < 1000000; i++ {
		time.Sleep(time.Millisecond)
		j := rand.Intn(len(key))
		core.Get(key[j])
		if core.IsHotKey(key[j]) {
			slog.Info("hotkey:" + fmt.Sprintf("%d", j))
		}
	}
}
