package group

import (
	"bilibili/common/middleware/hotkeys/model"
	"time"
)

func newCount(key string, id int, g *Group) *count {
	c := &count{
		key:      key,
		lastTime: time.Now().UnixMilli(),
		bucketId: id,
		group:    g,
		deleted:  false,
		window:   [20]int64{},
	}

	go c.check()
	return c
}

func (c *count) add(times int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.deleted {
		return
	}
	t := time.Now().UnixMilli()

	if t-c.lastTime >= 20000 {
		for i := 0; i < len(c.window); i++ {
			c.window[i] = 0
		}
		c.window[0] = times
		c.total = times
		c.lastTime = t

		c.send()
		return
	}

	for {
		if t/100 == c.lastTime/100 {
			break
		}
		c.lastTime += 100
		c.total -= c.window[c.lastIndex]
		c.lastIndex = (c.lastIndex + 1) % 20
	}
	c.total += times
	c.window[c.lastIndex] += times

	c.send()
	return
}

func (c *count) send() {
	if c.total >= 20 && time.Now().UnixMilli()-c.lastSend >= 65000 {
		c.group.Send(model.AddKey, []string{c.key})
		c.lastSend = time.Now().UnixMilli()
	}
}

func (c *count) check() {
	for {
		time.Sleep(time.Second * 30)

		c.mutex.Lock()

		if time.Now().UnixMilli()-c.lastTime > (time.Second * 20).Milliseconds() {
			c.delete()
			c.mutex.Unlock()
			return
		}

		c.mutex.Unlock()
	}
}

func (c *count) delete() {
	c.deleted = true
	c.group.countMutex[c.bucketId].Lock()
	c.group.bucket[c.bucketId].Remove(c.key)
	c.group.countMutex[c.bucketId].Unlock()
}
