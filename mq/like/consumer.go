package main

import (
	"context"
	"fansX/internal/middleware/lua"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	"fansX/internal/script/likeconsumerscript"
	"gorm.io/gorm"
	"log/slog"
	"strconv"
	"time"
)

type Consumer struct {
	db       *gorm.DB
	ch       chan *mq.Like
	close    chan bool
	change   map[[2]int64]int64
	executor *lua.Executor
}

func NewConsumer(db *gorm.DB) *Consumer {
	ch := make(chan *mq.Like, 1024*1024)
	return &Consumer{
		db: db,
		ch: ch,
	}
}

func (c *Consumer) Send(like *mq.Like) {
	c.ch <- like
}

func (c *Consumer) Close() {
	close(c.ch)
	c.close <- true
}

func (c *Consumer) Consume() {
	tick := time.Tick(time.Second * 10)
	for {
		select {
		case msg := <-c.ch:
			if msg.Cancel == true {
				c.change[[2]int64{int64(msg.Business), msg.LikeId}]--
			} else {
				c.change[[2]int64{int64(msg.Business), msg.LikeId}]++
			}
		case <-tick:
			c.Update()
		case <-c.close:
			return
		}
	}
}

func (c *Consumer) Update() {
	db := c.db
	executor := c.executor
	for k, v := range c.change {
		timeout, cancel := context.WithTimeout(context.Background(), time.Second)
		tx := db.WithContext(timeout).Begin()
		err := tx.Model(&database.LikeCount{}).Where("business = ? and like_id = ?", k[0], k[1]).
			Update("like_count", gorm.Expr("like_count + ?", v)).Error
		if err != nil {
			tx.Rollback()
			slog.Error("update like count:"+err.Error(), "business", k[0], "like_id", k[1])
		}
		executor.Execute(timeout, likeconsumerscript.AddScript, []string{
			"LikeNums:" + strconv.Itoa(int(k[0])) + ":" + strconv.FormatInt(k[1], 10),
			"false",
		}, v)

		cancel()
	}

	return
}
