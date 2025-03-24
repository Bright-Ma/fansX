package main

import (
	mlog "bilibili/log"
	"log/slog"
)

func main() {
	l := mlog.Init("test")
	slog.SetDefault(l)
	slog.Info("it is a test", "userid", 1)
}
