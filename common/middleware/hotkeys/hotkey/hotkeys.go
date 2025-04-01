package hotkey

import (
	"bilibili/common/middleware/hotkeys/model"
	"bilibili/pkg/consistenthash"
	"context"
	"encoding/json"
	"github.com/coocood/freecache"
	cmap "github.com/orcaman/concurrent-map/v2"
	etcd "go.etcd.io/etcd/client/v3"
	"log/slog"
	"net"
	"sync/atomic"
	"time"
)

type Core struct {
	cache   *freecache.Cache
	hotkeys *freecache.Cache
	group   string
	client  *etcd.Client
	version uint64
	conn    cmap.ConcurrentMap[string, *conn]
	hashMap *cshash.HashMap

	send chan kv
}

type kv struct {
	key   string
	times int
}

func NewCore(cacheSize int, etcdAddr []string, groupName string) (*Core, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   etcdAddr,
		DialTimeout: time.Second * 3,
	})
	if err != nil {
		return nil, err
	}
	c := &Core{
		cache:   freecache.NewCache(cacheSize),
		hotkeys: freecache.NewCache(cacheSize),
		group:   groupName,
		client:  client,
		version: 0,
		conn:    cmap.New[*conn](),
		hashMap: cshash.NewMap(50),
		send:    make(chan kv, 1024*1024*50),
	}

	err = c.init()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Core) init() error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	kvs, err := c.client.Get(timeout, "worker/", etcd.WithPrefix())
	if err != nil {
		return err
	}
	for _, v := range kvs.Kvs {
		c.connect(string(v.Value))
	}
	err = c.watch()
	if err != nil {
		return err
	}
	go c.Tick()
	go c.sendKey()

	return nil
}

func (c *Core) watch() error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	getResp, err := c.client.Get(timeout, "worker/", etcd.WithPrefix())
	if err != nil {
		return err
	}

	ch := c.client.Watch(timeout, "worker/", etcd.WithPrefix(), etcd.WithRev(getResp.Header.Revision))

	go func() {
		for resp := range ch {
			for _, v := range resp.Events {
				if v.Type == etcd.EventTypeDelete {
					c.closeConnect(string(v.Kv.Value))
				} else {
					c.connect(string(v.Kv.Value))
				}
			}
		}
	}()

	return nil
}

func (c *Core) sendKey() {
	ticker := time.NewTicker(time.Millisecond * 150)
	list := make(map[string]int)

	for {
		select {
		case <-ticker.C:
			c.push(list)
			clear(list)
		case value := <-c.send:
			_, ok := list[value.key]
			if !ok {
				list[value.key] = value.times
			} else {
				list[value.key] += value.times
			}
		}
	}
}

func (c *Core) push(list map[string]int) {

	mp := make(map[*conn]model.ClientMessage)
	for k, v := range list {
		addr := c.hashMap.Get([]string{k})
		connection, ok := c.conn.Get(addr[0])

		if ok {
			_, ok := mp[connection]
			if ok {
				mp[connection].Key[k] = v
			} else {
				mp[connection] = model.ClientMessage{
					Type: model.AddKey,
					Key:  make(map[string]int),
				}
				mp[connection].Key[k] = v
			}
		}
	}

	for connection, msg := range mp {
		body, err := json.Marshal(msg)
		if err != nil {
			slog.Error("marshal json:" + err.Error())
		}
		connection.write(body)
	}
}

func (c *Core) Tick() {
	time.Sleep(time.Second * 10)

	mp := c.conn.Items()
	for _, v := range mp {
		v.write(model.ClientPingMessage)
	}
}

func (c *Core) connect(addr string) {
	con, err := net.Dial("tcp", addr)
	if err != nil {
		slog.Error("connect:" + err.Error())
	}

	connection := &conn{closed: &atomic.Bool{}, conn: con, addr: addr, core: c}
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

func (c *Core) Get(key string) ([]byte, bool) {
	c.send <- kv{key, 1}

	res, err := c.cache.Get([]byte(key))
	if err != nil {
		return nil, false
	}

	return res, true
}

func (c *Core) Set(key string, value []byte, ttl int) bool {
	err := c.cache.Set([]byte(key), value, ttl)
	if err != nil {
		return false
	}

	return true
}

func (c *Core) Del(key string) bool {
	return c.cache.Del([]byte(key))
}

func (c *Core) IsHotKey(key string) bool {
	c.send <- kv{key, 1}

	_, err := c.hotkeys.Get([]byte(key))
	if err != nil {
		return false
	}

	c.send <- kv{key, 1}
	return true
}
