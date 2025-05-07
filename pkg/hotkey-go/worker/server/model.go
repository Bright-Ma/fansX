package server

import (
	"fansX/pkg/hotkey-go/model"
	"fansX/pkg/hotkey-go/worker/connection"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
)

type Handler struct {
	gnet.EventServer
	pool *goroutine.Pool
}

// MessageStrategy 策略模式
type MessageStrategy interface {
	Handle(msg *model.ClientMessage, conn *connection.Conn)
}

type AddStrategy struct {
}
type PingStrategy struct {
}
type PongStrategy struct {
}
