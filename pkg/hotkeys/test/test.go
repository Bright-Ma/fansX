package main

import (
	"bilibili/pkg/hotkeys/hotkey"
	"math/rand"
	"strconv"
)

func main() {
	for i := 0; i < 4; i++ {
		core, err := hotkey.NewCore(1024*1024*1024, []string{"1jian10.cn:4379"}, "test"+strconv.Itoa(i))
		if err != nil {
			panic(err.Error())
		}
		key := make([]string, 100000)
		for i := 0; i < 100000; i++ {
			key[i] = "key" + strconv.FormatInt(int64(i), 10)
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
