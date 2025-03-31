package group

import (
	"github.com/panjf2000/gnet"
)

type Conn struct {
	Last  int64
	Id    int64
	Conn  gnet.Conn
	Group *Group
}

func (c *Conn) Close() {
	c.Group.connectionSet.Remove(c.Id)
	_ = c.Conn.Close()
}
