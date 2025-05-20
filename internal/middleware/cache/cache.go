package bigcache

import (
	"context"
	"encoding/json"
	etcd "go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

type Cache struct {
	big    map[int64]bool
	rmu    sync.RWMutex
	client *etcd.Client
}

func NewCache(client *etcd.Client) *Cache {
	c := &Cache{
		big:    make(map[int64]bool),
		client: client,
	}
	go c.watch()
	return c
}

func (cache *Cache) watch() {
	ch := cache.client.Watch(context.Background(), "BigIndex/", etcd.WithPrefix())
	for resp := range ch {
		for _, ev := range resp.Events {
			if ev.Type == etcd.EventTypeDelete {
				continue
			} else {
				index := &BigIndex{}
				_ = json.Unmarshal(ev.Kv.Value, index)
			}
		}
	}
	panic("etcd watch err")
}
func (cache *Cache) update(ev *etcd.Event) error {
	index := &BigIndex{}
	_ = json.Unmarshal(ev.Kv.Value, index)
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	nmap := make(map[int64]bool)
	for _, v := range index.Key {
		resp, err := cache.client.Get(timeout, v)
		if err != nil {
			return err
		}
		set := &BigSet{}
		_ = json.Unmarshal(resp.Kvs[0].Value, set)
		for _, id := range set.Id {
			nmap[id] = true
		}
	}
	cache.rmu.Lock()
	defer cache.rmu.Unlock()
	cache.big = nmap
	return nil
}

func (cache *Cache) IsBig(id int64) bool {
	cache.rmu.RLock()
	defer cache.rmu.RUnlock()
	return cache.big[id]
}
