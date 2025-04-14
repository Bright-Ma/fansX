package group

import (
	"fansX/pkg/hotkey-go/model"
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

// 计算时间窗口内的访问次数
func (c *count) add(times int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.deleted {
		return
	}

	t := time.Now().UnixMilli()
	// 距离上次访问过去2s，重置时间窗口
	if t-c.lastTime >= 20000 {
		for i := 0; i < len(c.window); i++ {
			c.window[i] = 0
		}

		c.window[0] = times
		c.total = times
		c.lastTime = t

		c.send()

	} else {
		for {
			if t/100 == c.lastTime/100 {
				break
			}
			c.lastTime += 100
			next := (c.lastIndex + 1) % 20
			// 擦除该段内的值
			if t/100 != c.lastTime/100 {
				c.total -= c.window[next]
				c.window[next] = 0
			}
			c.lastIndex = next
		}
		c.total += times
		c.window[c.lastIndex] += times

		c.send()
	}

	return
}

// 热key判断以及尝试发送
func (c *count) send() {
	if c.total >= 20 && time.Now().UnixMilli()-c.lastSend >= 65000 {
		c.group.Send(model.AddKey, []string{c.key})
		c.lastSend = time.Now().UnixMilli()
	}
}

// key的心跳检测，若一段时间内不访问，则删除该key
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
