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
	"time"
)

func NewCore(config Config) (*Core, error) {
	client, err := etcd.New(config.EtcdConfig)
	if err != nil {
		return nil, err
	}

	c := &Core{
		cache:    freecache.NewCache(config.CacheSize),
		hotkeys:  freecache.NewCache(config.HotKeySize),
		group:    config.GroupName,
		delGroup: config.DelGroupName,
		client:   client,
		conn:     cmap.New[*conn](),
		hashMap:  cshash.NewMap(50),
		send:     make(chan kv, 1024*1024*50),
		del:      config.DelChan,
		hot:      config.HotChan,
	}

	err = c.init(config)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Core) init(config Config) error {
	if c.del == nil {
		c.delProcess = func(c *Core, msg *model.ServerMessage) {
			for _, v := range msg.Keys {
				_ = c.hotkeys.Del([]byte(v))
			}
		}
	} else {
		c.delProcess = func(c *Core, msg *model.ServerMessage) {
			for _, v := range msg.Keys {
				_ = c.hotkeys.Del([]byte(v))
				c.del <- v
			}
		}
	}

	if c.hot == nil {
		c.addProcess = func(c *Core, msg *model.ServerMessage) {
			for _, v := range msg.Keys {
				_ = c.hotkeys.Set([]byte(v), []byte{}, 60)
			}
		}
	} else {
		c.addProcess = func(c *Core, msg *model.ServerMessage) {
			for _, v := range msg.Keys {
				_ = c.hotkeys.Set([]byte(v), []byte{}, 60)
				c.hot <- v
			}
		}
	}

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
	if config.Model == ModelCache {
		go c.sendKey()
	} else {
		go c.sendDel()
	}

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

func (c *Core) sendDel() {
	ticker := time.NewTicker(time.Millisecond * 150)
	del := make(map[string]int)

	for {
		select {
		case <-ticker.C:
			msg := model.ClientMessage{
				Type:      model.DelKey,
				GroupName: c.delGroup,
				Key:       del,
			}
			buf, err := json.Marshal(msg)
			if err != nil {
				slog.Error("marshal json:" + err.Error())
				continue
			}
			mp := c.conn.Items()
			for _, v := range mp {
				v.write(buf)
				return
			}
		case value := <-c.send:
			del[value.key] = 1
		}
	}

}
