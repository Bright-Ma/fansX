package hotkey

import (
	"bilibili/pkg/hotkeys/model"
	"encoding/json"
	"log/slog"
	"net"
	"sync/atomic"
	"time"
)

func (c *Core) Tick() {
	time.Sleep(time.Second * 10)
	t := time.Now().Unix()
	mp := c.conn.Items()

	for _, v := range mp {
		if t-v.last >= 30 {
			c.closeConnect(v.addr)
			continue
		}
		v.write(model.ClientPingMessage)
	}
}

func (c *Core) connect(addr string) {
	con, err := net.Dial("tcp", addr)
	if err != nil {
		slog.Error("connect:" + err.Error())
	}

	connection := &conn{closed: &atomic.Bool{}, conn: con, addr: addr, core: c, last: time.Now().Unix()}
	connection.closed.Store(false)

	g, err := json.Marshal(model.ClientMessage{
		Type:      model.Group,
		GroupName: c.group,
	})
	if err != nil {
		slog.Warn("marshal json:" + err.Error())
	}
	connection.write(g)

	c.conn.Set(addr, connection)
	c.hashMap.Update([]string{}, []string{addr})

	go connection.process()
}

func (c *Core) closeConnect(addr string) {
	connection, ok := c.conn.Get(addr)
	if !ok {
		return
	}

	c.conn.Remove(addr)
	c.hashMap.Update([]string{connection.addr}, []string{})
	_ = connection.conn.Close()
	connection.closed.Store(true)
}
