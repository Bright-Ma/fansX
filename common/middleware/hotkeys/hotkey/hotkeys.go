package hotkey

import (
	"context"
	"github.com/coocood/freecache"
	etcd "go.etcd.io/etcd/client/v3"
	"time"
)

type Core struct {
	cache  *freecache.Cache
	client *etcd.Client
}

func (c *Core) Get(key string) ([]byte, bool, error) {
	v, err := c.cache.Get([]byte(key))
	if err != nil {

	}
	return v, true
}

func (c *Core) Set(key string, value []byte, ttl int) error {
	return c.cache.Set([]byte(key), value, ttl)
}

func (c *Core) GetHotKeysChannel()

func NewCore(Size int) *Core {

	c := &Core{cache: freecache.NewCache(Size)}
	return c
}

func (c *Core) watch() error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	getResp, err := c.client.Get(timeout, "key", etcd.WithPrefix())
	if err != nil {
		return err
	}

	ch := c.client.Watch(timeout, "key", etcd.WithPrefix(), etcd.WithRev(getResp.Header.Revision))

	go func() {
		for resp := range ch {
			for _, v := range resp.Events {
				if v.Type == etcd.EventTypeDelete {

				} else {

				}
			}
		}
	}()

	return nil
}
