package group

import (
	"bilibili/pkg/hotkey-go/model"
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
	h := fnv.New64a()

	for i := 0; i < len(keys); i++ {
		// 计算bucketId
		_, _ = h.Write([]byte(keys[i]))
		value := h.Sum64() % uint64(g.bucketSize)

		// 这里先进行无锁的get，只是粗略的判断
		c, ok := g.bucket[value].Get(keys[i])
		if !ok {
			g.countMutex[value].Lock()
			// 这里使用有锁的get，保证并发安全
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
