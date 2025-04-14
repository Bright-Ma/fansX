package bigcache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
	"sync"
	"time"
)

type Cache struct {
	big     map[int64]bool
	version int64
	rmu     *sync.RWMutex
	client  *redis.Client
}

func Init(client *redis.Client) (*Cache, error) {
	c := &Cache{
		big:     make(map[int64]bool),
		version: 0,
		rmu:     nil,
		client:  client,
	}
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := c.check(timeout)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			time.Sleep(time.Minute * 5)
			for {
				timeout, cancel := context.WithTimeout(context.Background(), time.Second*20)
				err := c.check(timeout)
				if err != nil {
					cancel()
					continue
				} else {
					cancel()
					break
				}
			}
		}
	}()

	return c, nil

}

func (cache *Cache) check(ctx context.Context) error {
	var res map[string]string
	var err error

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			res, err = cache.client.HGetAll(context.Background(), "BigHash").Result()
			if err != nil {
				time.Sleep(time.Millisecond * 50)
				continue
			} else {
				break
			}
		}
		break
	}

	ver := res["version"]
	version, _ := strconv.ParseInt(ver, 10, 64)
	if version == cache.version {
		return nil
	}

	keys := make([]string, 0)
	for k, v := range res {
		if k == "version" {
			continue
		}
		keys = append(keys, v)
	}

	set := make(map[int64]bool)
	for i := 0; i < len(keys); i++ {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				members, err := cache.client.SMembers(context.Background(), keys[i]).Result()
				if err != nil {
					time.Sleep(time.Millisecond * 50)
					continue
				} else {
					for _, m := range members {
						member, _ := strconv.ParseInt(m, 10, 64)
						set[member] = true
					}
					break
				}
			}
			break
		}
	}

	cache.rmu.Lock()
	cache.big = set
	cache.version = version
	cache.rmu.Unlock()
	return nil
}

func (cache *Cache) IsBig(id int64) bool {
	cache.rmu.RLocker()
	defer cache.rmu.RUnlock()
	_, ok := cache.big[id]
	return ok
}
