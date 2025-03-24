package mlog

import (
	"log/slog"
	"os"
)

func Init(name string) *slog.Logger {
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
	return l.With("ServiceName", name).WithGroup("detail")
}
