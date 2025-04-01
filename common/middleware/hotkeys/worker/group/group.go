package group

import (
	"bilibili/common/middleware/hotkeys/model"
	"encoding/json"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/panjf2000/gnet"
	"hash/fnv"
	"strconv"
	"sync"
	"time"
)

func init() {
	groupMap = cmap.New[*Group]()
}

func GetGroupMap() *cmap.ConcurrentMap[string, *Group] {
	return &groupMap
}

func NewGroup() *Group {
	g := &Group{
		countMutex:    make([]sync.Mutex, 8),
		bucket:        make([]cmap.ConcurrentMap[string, *count], 8),
		connectionSet: cmap.New[*Conn](),
	}
	for i := 0; i < 8; i++ {
		g.bucket[i] = cmap.New[*count]()
	}

	return g
}

func (g *Group) AddConn(conn gnet.Conn) *Conn {
	g.idMutex.Lock()
	id := g.nextId
	g.nextId++
	g.idMutex.Unlock()

	c := &Conn{
		Last:  time.Now().Unix(),
		Id:    id,
		Conn:  conn,
		Group: g,
	}
	g.connectionSet.Set(strconv.FormatInt(id, 10), c)

	return c
}

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
	h := fnv.New64a()

	for i := 0; i < len(keys); i++ {
		_, _ = h.Write([]byte(keys[i]))
		value := h.Sum64() % 8

		c, ok := g.bucket[value].Get(keys[i])
		if !ok {
			g.countMutex[value].Lock()

			_, ok = g.bucket[value].Get(keys[i])
			if !ok {
				c = newCount(keys[i], int(value), g)
				g.bucket[value].Set(keys[i], c)
			}

			g.countMutex[value].Unlock()

		}

		c.add(times[i])
		h.Reset()
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
