package segment

func (c *Creator) GetId() int64 {
	for {
		c.mu.Lock()
		if c.old.nextId == c.old.maxId+1 && c.new == nil {
			c.mu.Unlock()
		} else {
			break
		}
	}
	defer c.mu.Unlock()
	if c.old.nextId == c.old.maxId+1 {
		c.old = c.new
		c.new = nil
	}
	if c.old.nextId == c.old.preIndex {
		c.ch <- 1
	}
	res := c.old.nextId
	c.old.nextId++
	return res
}
