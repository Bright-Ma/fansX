package server

import (
	"encoding/json"
	"fansX/pkg/hotkey-go/model"
	"fansX/pkg/hotkey-go/worker/connection"
	"github.com/panjf2000/gnet"
	"log/slog"
)

func (h *Handler) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	c.SetContext(connection.NewConn(c))
	return
}

// React tcp报文的处理
func (h *Handler) React(packet []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	msg := &model.ClientMessage{}
	err := json.Unmarshal(packet, msg)
	if err != nil {
		slog.Warn("json unmarshal:" + err.Error())
		return nil, gnet.None
	}
	s := GetStrategy(msg.Type)
	if s == nil {
		slog.Warn("get strategy fail unknow type:" + msg.Type)
		return nil, gnet.Close
	}
	ctx := c.Context()
	_ = h.pool.Submit(func() {
		s.Handle(msg, ctx.(*connection.Conn))
	})

	return nil, gnet.None
}

func (h *Handler) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	if err != nil {
		slog.Error(err.Error())
	} else {
		slog.Info("connect closed")
	}

	return gnet.None
}
