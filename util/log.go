package util

import (
	"log/slog"
	"os"
)

func InitLog() *slog.Logger {
	//file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	//if err != nil {
	//	panic(err.Error())
	//}
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

	return l
}
