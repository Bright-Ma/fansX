package hotkey

import (
	"encoding/binary"
)

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

func (c *Core) EncodeSet(key string, value any, ttl int) bool {
	buf := make([]byte, 0)
	_, err := binary.Encode(buf, binary.LittleEndian, value)
	if err != nil {
		return false
	}
	return c.Set(key, buf, ttl)
}

func (c *Core) Del(key string) bool {
	return c.cache.Del([]byte(key))
}

func (c *Core) SendDel(key string) {
	c.send <- kv{key, 1}
	return
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
