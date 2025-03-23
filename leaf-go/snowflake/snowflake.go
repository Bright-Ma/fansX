package snowflake

import (
	"context"
	"strconv"
	"time"
)

func (c *Creator) GetId() int64 {
	return int64(c.snowNode.Generate())
}

func (c *Creator) Lock() {
	c.client.Txn(context.Background()).If()
}

func (c *Creator) heartCheck() {
	ch := time.NewTicker(time.Second)
	key := "IdCreatorForever/" + c.name + "/" + c.addr
	for {
		select {
		case <-ch.C:
			for {
				timeout, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
				resp, err := c.client.Get(timeout, key)
				if err != nil {
					cancel()
					continue
				} else {
					t, err := strconv.ParseInt(string(resp.Kvs[0].Value), 10, 64)
					if err != nil {
						cancel()
						continue
					}
					if time.Now().UnixMilli()-t <= 0 {
						panic("clock failed")
					}
					cancel()
					break
				}
			}
		}
	}
}
