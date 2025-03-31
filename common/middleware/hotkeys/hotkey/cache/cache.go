package cache

import "github.com/coocood/freecache"

type Cache struct {
	freeCache *freecache.Cache
	HotKey    *freecache.Cache
}

func (c *Cache) Get(key string) ([]byte, error) {
	val, err := c.freeCache.Get([]byte(key))
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (c *Cache) Set(key string, val []byte, ttl int) error {
	err := c.freeCache.Set([]byte(key), val, ttl)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) IsHotKey(key string) bool {
	_, err := c.HotKey.Get([]byte(key))
	if err != nil {
		return false
	}

	return true
}
