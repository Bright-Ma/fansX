package connection

import (
	"fansX/pkg/hotkey-go/model"
	"github.com/panjf2000/gnet"
	"strconv"
	"time"
)

func (c *Conn) Close() {
	_ = c.conn.Close()
}
func (c *Conn) Ping() {
	_ = c.conn.AsyncWrite(model.ServerPingMessage)
}

func (c *Conn) Pong() {
	c.last = time.Now().Unix()
	_ = c.conn.AsyncWrite(model.ServerPongMessage)
}

func (c *Conn) ReSetTime() {
	c.last = time.Now().Unix()
}

func (c *Conn) Send(msg []byte) {
	_ = c.conn.AsyncWrite(msg)
}

// IsTimeout 该连接是否超时
func (c *Conn) IsTimeout() bool {
	if time.Now().Unix()-c.last > 60 {
		return true
	}
	return false
}

func (c *Conn) String() string {
	return c.id
}

func NewConn(conn gnet.Conn) *Conn {
	// 获取唯一连接id
	idMutex.Lock()
	next := nextId
	nextId++
	idMutex.Unlock()

	c := &Conn{
		id:   strconv.FormatInt(next, 10),
		last: time.Now().Unix(),
		conn: conn,
	}

	return c
}
