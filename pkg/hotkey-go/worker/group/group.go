package group

import (
	"encoding/json"
	"fansX/pkg/hotkey-go/model"
	"fansX/pkg/hotkey-go/worker/config"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/panjf2000/gnet"
	"strconv"
	"time"
)

func init() {
	Map = MapSubject{concurrentMap: cmap.New[Observer]()}
}

func GetGroupMap() *MapSubject {
	return &Map
}

func NewGroup(cf config.Config) *Group {
	g := &Group{
		config:        cf,
		nextId:        0,
		keys:          cmap.New[*count](),
		connectionSet: cmap.New[*Conn](),
	}
	return g
}

func (g *Group) AddConn(conn gnet.Conn) *Conn {
	g.idMutex.Lock()
	// 获取唯一id
	id := g.nextId
	g.nextId++
	g.idMutex.Unlock()

	c := &Conn{
		Last:  time.Now().Unix(),
		Id:    id,
		Conn:  conn,
		Group: g,
	}
	// 在连接集合中添加conn
	g.connectionSet.Set(strconv.FormatInt(id, 10), c)

	return c
}

// Send 广播，对该group中的所有连接发送消息
func (g *Group) Send(m string, key []string) {
	msg := &model.ServerMessage{
		Type: m,
		Keys: key,
	}

	s, err := json.Marshal(msg)
	if err != nil {
		return
	}

	mp := g.connectionSet.Items()
	for _, v := range mp {
		_ = v.Conn.AsyncWrite(s)
	}
}

func (g *Group) AddKey(keys []string, times []int64) {
	for i, v := range keys {
		c := g.keys.Upsert(v, nil, func(exist bool, in *count, new *count) *count {
			if exist {
				return in
			} else {
				return newCount(keys[i], 0, g)
			}
		})
		c.add(times[i])
	}
	return
}

func (g *Group) Tick() {
	mp := g.connectionSet.Items()
	t := time.Now().Unix()
	for _, v := range mp {
		if t-v.Last >= 30 {
			v.Close()
		}
	}

	return
}
