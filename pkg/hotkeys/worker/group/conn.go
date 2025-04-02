package group

import (
	"github.com/panjf2000/gnet"
	"strconv"
)

type Conn struct {
	Last  int64
	Id    int64
	Conn  gnet.Conn
	Group *Group
}

func (c *Conn) Close() {
	c.Group.connectionSet.Remove(strconv.FormatInt(c.Id, 10))
	_ = c.Conn.Close()
}
