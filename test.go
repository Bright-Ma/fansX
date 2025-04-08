package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

func main() {
	nums := []int{1, 2, 3, 4, 5}
	buf, err := json.Marshal(nums)
	if err != nil {
		slog.Error(err.Error())
	}
	res := make([]int, 0)
	err = json.Unmarshal(buf, &res)
	if err != nil {
		slog.Error(err.Error())
	}
	fmt.Println(res)

}
