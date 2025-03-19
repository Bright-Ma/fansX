package main

import (
	"log/slog"
	"os"
)

type Handler struct {
}

//func (h *Handler) Enabled(ctx, context.Context, level slog.Level) bool {
//	return true
//}

func main() {

	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
	slog.SetDefault(l)
	slog.Info("it is a test")

	return

}
