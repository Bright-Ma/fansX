package server

import (
	"encoding/json"
	"fansX/pkg/hotkey-go/model"
	"fansX/pkg/hotkey-go/worker/group"
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

// React tcp报文的处理
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
	} else if msg.Type == model.DelKey {
		_ = h.pool.Submit(func() { h.HandDel(msg, c) })
	} else {
		slog.Warn("unKnow message:" + fmt.Sprintln(msg))
	}

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

// handPing 对ping包的处理，重置时间戳，回复pong
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

// handPong 对pong包的处理，重置时间戳，由于服务端不主动发送ping，故此函数一般不会调用
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

// HandGroup 处理设置group的报文，将该连接加入到groupMap中
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

// HandAdd 处理go程序发送的key以及访问次数
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
	g, ok := group.GetGroupMap().Get(msg.GroupName)
	if !ok {
		return nil, gnet.None
	}
	g.Send(model.DelKey, keys)

	return nil, gnet.None
}

// Tick 心跳处理
func (h *Handler) Tick() (time.Duration, gnet.Action) {

	mp := group.GetGroupMap().Items()
	for _, v := range mp {
		v.Tick()
	}

	return time.Second * 15, gnet.None
}
