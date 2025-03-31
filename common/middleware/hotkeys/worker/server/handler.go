package server

import (
	"bilibili/common/middleware/hotkeys/model"
	"bilibili/common/middleware/hotkeys/worker/group"
	"encoding/json"
	"fmt"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
	"log/slog"
	"time"
)

type Handler struct {
	gnet.EventServer
	pool *goroutine.Pool
}

func (h *Handler) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	return
}

func (h *Handler) React(packet []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	msg := &model.ClientMessage{}
	err := json.Unmarshal(packet, msg)
	if err != nil {
		slog.Warn("json unmarshal:" + err.Error())
		return nil, gnet.None
	}
	if msg.Type == model.Ping {
		return model.ServerPongMessage, gnet.None

	} else if msg.Type == model.Pong {
		return model.ServerPingMessage, gnet.None

	} else if msg.Type == model.Group {
		g, ok := group.GetGroupMap().Get(msg.GroupName)
		if !ok {
			return nil, gnet.Close
		}
		c.SetContext(&context{
			conn:  g.AddConn(c),
			group: g,
		})

	} else if msg.Type == model.AddKey {
		inter := c.Context()
		ctx, ok := inter.(*context)
		if !ok {
			ctx.conn.Close()
			return nil, gnet.None
		}

		ctx.group.AddKey(msg.Keys, msg.Times)

	} else if msg.Type == model.DelKey {
		inter := c.Context()
		ctx, ok := inter.(*context)
		if !ok {
			ctx.conn.Close()
			return nil, gnet.None
		}

		ctx.group.Send(model.DelKey, msg.Keys)

	} else {
		slog.Warn("unKnow message:" + fmt.Sprintln(msg))
	}

	return nil, gnet.None
}

func (h *Handler) Tick() (time.Duration, gnet.Action) {

	mp := group.GetGroupMap().Items()
	for _, v := range mp {
		v.Tick()
	}

	return time.Second * 30, gnet.None
}
