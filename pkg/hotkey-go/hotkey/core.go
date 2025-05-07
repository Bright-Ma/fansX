package hotkey

import (
	"context"
	"encoding/json"
	cshash "fansX/pkg/consistenthash"
	"fansX/pkg/hotkey-go/model"
	"github.com/coocood/freecache"
	cmap "github.com/orcaman/concurrent-map/v2"
	etcd "go.etcd.io/etcd/client/v3"
	"log/slog"
	"net"
	"sync/atomic"
	"time"
)

// NewCore 使用with...来更改默认值
func NewCore(GroupName string, client *etcd.Client, options ...Option) (*Core, error) {
	c := &Core{
		cache:        freecache.NewCache(1024 * 1024 * 1024 * 4),
		hotkeys:      freecache.NewCache(1024 * 1024 * 128),
		group:        GroupName,
		client:       client,
		conn:         cmap.New[*conn](),
		hashMap:      cshash.NewMap(50),
		send:         make(chan kv, 1024*512),
		interval:     time.Millisecond * 100,
		observerList: make([]Observer, 0),
		ttl:          30,
	}

	for _, option := range options {
		option.Update(c)
	}

	err := c.init()
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

	go c.tick()
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

// 发送key的访问信息
func (c *Core) sendKey() {
	ticker := time.NewTicker(c.interval)
	list := make(map[string]int)

	for {
		select {
		case <-ticker.C:
			c.push(list)
			clear(list)
		case value := <-c.send:
			// 内存聚合
			list[value.key] += value.times
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
					Type:      model.AddKey,
					GroupName: c.group,
					Key:       make(map[string]int),
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

// 连接，添加连接进连接集合，从一致性hash环中移除对端
func (c *Core) connect(addr string) {
	con, err := net.Dial("tcp", addr)
	if err != nil {
		slog.Error("connect:" + err.Error())
		return
	}

	connection := &conn{closed: &atomic.Bool{}, conn: con, addr: addr, core: c, last: time.Now().Unix()}
	connection.closed.Store(false)

	c.conn.Set(addr, connection)
	c.hashMap.Update([]string{}, []string{addr})

	go connection.process()
}

// 关闭连接，在一致性hash环中和连接集合中移除对端
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

func (c *Core) register(ob Observer) {
	c.observerList = append(c.observerList, ob)
}

func (c *Core) notify(key string) {
	for _, ob := range c.observerList {
		ob.Do(key)
	}
}

// Get 从本地缓存中获取value，视为一次对key的访问
func (c *Core) Get(key string) ([]byte, bool) {
	c.send <- kv{key, 1}

	res, err := c.cache.Get([]byte(key))
	if err != nil {
		return nil, false
	}

	return res, true
}

// Set 将kv添加到本地缓存
func (c *Core) Set(key string, value []byte, ttl int) bool {
	err := c.cache.Set([]byte(key), value, ttl)
	if err != nil {
		return false
	}

	return true
}

// Del 将kv从本地缓存中移除
func (c *Core) Del(key string) bool {
	return c.cache.Del([]byte(key))
}

// IsHotKey 判断一个key是否为热key，视为对key的一次访问
func (c *Core) IsHotKey(key string) bool {
	c.send <- kv{key, 1}

	_, err := c.hotkeys.Get([]byte(key))
	if err != nil {
		return false
	}

	c.send <- kv{key, 1}
	return true
}

func (c *Core) tick() {
	time.Sleep(time.Second * 5)
	t := time.Now().Unix()
	mp := c.conn.Items()

	for _, v := range mp {
		if t-v.last >= 60 {
			c.closeConnect(v.addr)
			continue
		}
		v.write(model.ClientPingMessage)
	}
}
