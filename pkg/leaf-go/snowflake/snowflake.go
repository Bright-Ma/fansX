package snowflake

import (
	"context"
	"strconv"
	"time"
)

func (c *Creator) GetId() (int64, bool) {
	if c.working.Load() {
		return int64(c.snowNode.Generate()), true
	} else {
		return 0, false
	}
}

// heartCheck 心跳，定时上报时钟到本地和etcd
func (c *Creator) heartCheck() {
	ch := time.NewTicker(time.Millisecond * 200)
	key := "IdCreatorForever/" + c.name + "/" + c.addr

	for !c.local.Load() {
		select {
		case <-ch.C:
			timeout, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
			resp, err := c.client.Get(timeout, key)
			//当etcd请求失效时将本地存储的时钟作为依据
			var t int64

			if err != nil {
				t = c.lastTime
			} else {
				t, err = strconv.ParseInt(string(resp.Kvs[0].Value), 10, 64)
				if err != nil {
					t = c.lastTime
				}

				if time.Now().UnixMilli()-t <= 0 {
					//小步长
					if time.Now().UnixMilli()-t <= 500 {

						c.working.Store(false)
						time.Sleep(time.Millisecond * 1000)
						c.working.Store(true)
						//大步长
					} else {
						panic("clock failed")
					}
				}
			}
			_, _ = c.client.Put(timeout, key, strconv.FormatInt(time.Now().UnixMilli(), 10))
			c.lastTime = time.Now().UnixMilli()
			cancel()
		}
	}

	for range ch.C {
		t := c.lastTime

		if time.Now().UnixMilli()-t <= 0 {

			if time.Now().UnixMilli()-t <= 500 {

				c.working.Store(false)
				time.Sleep(time.Millisecond * 1000)
				c.working.Store(true)

			} else {
				panic("clock failed")
			}

		}
		c.lastTime = time.Now().UnixMilli()
	}
}
