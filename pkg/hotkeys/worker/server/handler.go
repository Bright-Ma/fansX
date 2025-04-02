package server

import (
	"bilibili/pkg/hotkeys/model"
	"bilibili/pkg/hotkeys/worker/group"
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
		return h.handPing(c)

	} else if msg.Type == model.Pong {
		return h.handPong(c)

	} else if msg.Type == model.Group {
		return h.HandGroup(msg, c)

	} else if msg.Type == model.AddKey {
		_ = h.pool.Submit(func() { h.HandAdd(msg, c) })
		return nil, gnet.None

	} else if msg.Type == model.DelKey {
		_ = h.pool.Submit(func() { h.HandDel(msg, c) })
		return nil, gnet.None

	} else {
		slog.Warn("unKnow message:" + fmt.Sprintln(msg))
	}

	return nil, gnet.None
}

func (h *Handler) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	if err != nil {
		slog.Error(err.Error())
	} else {
		slog.Debug("connection closed")
	}

	return gnet.None
}

func (h *Handler) handPing(c gnet.Conn) (out []byte, action gnet.Action) {
	inter := c.Context()
	if inter == nil {
		slog.Warn("nil context")
		return nil, gnet.None
	}

	ctx, ok := inter.(*context)
	if !ok {
		slog.Error("not found context")
		return nil, gnet.None
	}

	ctx.conn.Last = time.Now().Unix()

	return model.ServerPongMessage, gnet.None
}

func (h *Handler) handPong(c gnet.Conn) (out []byte, action gnet.Action) {
	inter := c.Context()
	if inter == nil {
		slog.Warn("nil context")
		return nil, gnet.None
	}

	ctx, ok := inter.(*context)
	if !ok {
		slog.Error("not found context")
		return nil, gnet.None
	}

	ctx.conn.Last = time.Now().Unix()

	return nil, gnet.None
}

func (h *Handler) HandGroup(msg *model.ClientMessage, c gnet.Conn) (out []byte, action gnet.Action) {
	g, ok := group.GetGroupMap().Get(msg.GroupName)
	if !ok {
		return nil, gnet.Close
	}
	c.SetContext(&context{
		conn:  g.AddConn(c),
		group: g,
	})

	return nil, gnet.None
}

func (h *Handler) HandAdd(msg *model.ClientMessage, c gnet.Conn) (out []byte, action gnet.Action) {
	inter := c.Context()
	if inter == nil {
		slog.Warn("nil context")
		return nil, gnet.None
	}

	ctx, ok := inter.(*context)

	if !ok {
		ctx.conn.Close()
		return nil, gnet.None
	}

	keys := make([]string, len(msg.Key))
	times := make([]int64, len(msg.Key))
	i := 0
	for k, v := range msg.Key {
		keys[i] = k
		times[i] = int64(v)
		i++
	}

	ctx.group.AddKey(keys, times)

	return nil, gnet.None
}

func (h *Handler) HandDel(msg *model.ClientMessage, c gnet.Conn) (out []byte, action gnet.Action) {
	inter := c.Context()
	if inter == nil {
		slog.Warn("nil context")
		return nil, gnet.None
	}

	ctx, ok := inter.(*context)
	if !ok {
		ctx.conn.Close()
		return nil, gnet.None
	}
	keys := make([]string, len(msg.Key))
	i := 0
	for k := range msg.Key {
		keys[i] = k
		i++
	}

	ctx.group.Send(model.DelKey, keys)
	return nil, gnet.None
}

func (h *Handler) Tick() (time.Duration, gnet.Action) {

	mp := group.GetGroupMap().Items()
	for _, v := range mp {
		v.Tick()
	}

	return time.Second * 30, gnet.None
}
