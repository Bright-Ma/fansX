package main

import (
	"log/slog"
	"os"
	"time"
)

func main() {
	file, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		panic(err.Error())
	}
	l := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
	for i := 0; i < 10000; i++ {
		time.Sleep(time.Millisecond * 10)
		l.Info("it is a test")
	}
}
