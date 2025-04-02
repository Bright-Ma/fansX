package hotkey

import (
	cshash "bilibili/pkg/consistenthash"
	"bilibili/pkg/hotkey-go/model"
	"context"
	"encoding/json"
	"github.com/coocood/freecache"
	cmap "github.com/orcaman/concurrent-map/v2"
	etcd "go.etcd.io/etcd/client/v3"
	"log/slog"
	"time"
)

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
