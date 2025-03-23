package main

import (
	leaf "bilibili/leaf-go"
	"bilibili/leaf-go/segment"
	"bilibili/leaf-go/snowflake"
	"fmt"
	"sync"
	"time"
)

const totalGoroutines = 16
const idsPerGoroutine = 1000000

var wg sync.WaitGroup
var model = 1

func main() {

	startTime := time.Now()

	for i := 0; i < totalGoroutines; i++ {
		wg.Add(1)
		if model == 1 {
			go SnowFlakeGetIds(fmt.Sprintf("goroutine-%d", i))
		} else {
			go SegmentGetIds(fmt.Sprintf("goroutine-%d", i))
		}
	}

	wg.Wait()

	duration := time.Since(startTime)
	totalIds := totalGoroutines * idsPerGoroutine
	idsPerSecond := float64(totalIds) / duration.Seconds()

	fmt.Printf("Total IDs generated: %d\n", totalIds)
	fmt.Printf("Total time taken: %v\n", duration)
	fmt.Printf("Average IDs per second: %.2f\n", idsPerSecond)
}

func SegmentGetIds(name string) {
	creator, err := leaf.Init(&leaf.Config{
		Model: leaf.Segment,
		SegmentConfig: &segment.Config{
			Name:     "test",
			UserName: "root",
			Password: "",
			Address:  "linux.1jian10.cn:4000",
		},
	})
	if err != nil {
		panic(err.Error())
	}
	defer wg.Done()
	for i := 0; i < idsPerGoroutine; i++ {
		creator.GetId()
	}
}

func SnowFlakeGetIds(name string) {
	creator, err := leaf.Init(&leaf.Config{
		Model: leaf.Snowflake,
		SnowflakeConfig: &snowflake.Config{
			CreatorName: "test",
			Addr:        "127.0.0.1:23333" + name,
			EtcdAddr:    []string{"127.0.0.1:4379"},
		},
	})
	if err != nil {
		panic(err.Error())
	}
	defer wg.Done()
	for i := 0; i < idsPerGoroutine; i++ {
		creator.GetId()
	}

}
