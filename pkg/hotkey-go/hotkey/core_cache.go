package hotkey

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
