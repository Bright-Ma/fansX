package group

import (
	"bilibili/pkg/hotkeys/model"
	"encoding/json"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/panjf2000/gnet"
	"hash/fnv"
	"log/slog"
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
	size := 32
	g := &Group{
		countMutex:    make([]sync.Mutex, size),
		bucket:        make([]cmap.ConcurrentMap[string, *count], size),
		connectionSet: cmap.New[*Conn](),
		bucketSize:    size,
	}
	for i := 0; i < size; i++ {
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
		value := h.Sum64() % uint64(g.bucketSize)

		c, ok := g.bucket[value].Get(keys[i])
		if !ok {
			g.countMutex[value].Lock()

			c, ok = g.bucket[value].Get(keys[i])
			if !ok {
				c = newCount(keys[i], int(value), g)
				g.bucket[value].Set(keys[i], c)
			}

			g.countMutex[value].Unlock()

		}
		if c == nil {
			slog.Error("nil c")
			h.Reset()
			continue
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
