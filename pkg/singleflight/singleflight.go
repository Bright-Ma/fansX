package singleflight

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

type Group struct {
	sync *redsync.Redsync
}

func NewGroup(client *redis.Client) *Group {
	pool := goredis.NewPool(client)
	return &Group{
		sync: redsync.New(pool),
	}
}

func (g *Group) Do(key string, fn func()) {
	if err := g.sync.NewMutex(key).TryLock(); err != nil {
		return
	}

	fn()

	return
}
