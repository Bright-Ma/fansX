package main

import (
	"fansX/internal/middleware/lua"
	"fansX/internal/model/database"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

type Consumer struct {
	db           *gorm.DB
	ch           chan *database.Comment
	close        chan bool
	executor     *lua.Executor
	commentList  []*database.Comment
	commentCount map[[2]int64]int64
}

func NewConsumer(db *gorm.DB, client *redis.Client) *Consumer {
	ch := make(chan *database.Comment, 1024*1024)
	return &Consumer{
		db:    db,
		ch:    ch,
		close: make(chan bool, 1),
	}
}

func (c *Consumer) Send(record *database.Comment) {
	c.ch <- record
}

func (c *Consumer) Close() {
	close(c.ch)
	c.close <- true
}

func (c *Consumer) Consume() {
	tick := time.Tick(time.Second)
	for {
		select {
		case <-tick:
			c.Update()
			clear(c.commentList)
			clear(c.commentCount)
		case msg := <-c.ch:
			c.InsertList(msg)
		case <-c.close:
			return
		}
	}
}
func (c *Consumer) InsertList(msg *database.Comment) {
	c.commentList = append(c.commentList, msg)
	c.commentCount[[2]int64{database.BusinessContent, msg.ContentId}]++
	if msg.RootId == 0 {
		c.commentCount[[2]int64{database.BusinessComment, msg.RootId}]++
	}
	return
}

func (c *Consumer) Update() {
	if len(c.commentList) == 0 {
		return
	}
	ma := min(999, len(c.commentList)-1)
	mi := 0
	db := c.db
	for {
		db.Create(c.commentList[mi:ma])
		mi = ma + 1
		if mi == len(c.commentList) {
			break
		}
		ma = mi + 999
		if ma >= len(c.commentList) {
			ma = len(c.commentList) - 1
		}
	}
	model := database.CommentCount{}
	for i, v := range c.commentCount {
		db.Model(&model).Where("business = ? and count_id = ?", i[0], i[1]).
			Update("count", gorm.Expr("count + ?", v))
	}

	return
}
