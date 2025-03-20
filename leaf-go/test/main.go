package main

import (
	"fmt"
	leaf "leaf-go"
	"leaf-go/segment"
	"sync"
	"time"
)

func main() {
	const totalGoroutines = 1
	const idsPerGoroutine = 1000000

	var wg sync.WaitGroup

	startTime := time.Now()

	getIds := func(name string) {
		creator, err := leaf.InitLeaf(&leaf.Config{
			Model: leaf.Segment,
			SegmentConfig: segment.Config{
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

	for i := 0; i < totalGoroutines; i++ {
		wg.Add(1)
		go getIds(fmt.Sprintf("goroutine-%d", i))
	}

	wg.Wait()

	duration := time.Since(startTime)
	totalIds := totalGoroutines * idsPerGoroutine
	idsPerSecond := float64(totalIds) / duration.Seconds()

	fmt.Printf("Total IDs generated: %d\n", totalIds)
	fmt.Printf("Total time taken: %v\n", duration)
	fmt.Printf("Average IDs per second: %.2f\n", idsPerSecond)
}
