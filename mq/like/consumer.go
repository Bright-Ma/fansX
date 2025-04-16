package main

import (
	"fansX/internal/model/mq"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"time"
)

type Consumer struct {
	session sarama.ConsumerGroupSession
	db      *gorm.DB
	ch      chan mq.Like
	close   chan int
	window  map[int64]map[int64][]int64
}

func NewConsumer(session sarama.ConsumerGroupSession, db *gorm.DB) *Consumer {
	ch := make(chan mq.Like, 1024*1024)
	return &Consumer{
		db:      db,
		session: session,
		ch:      ch,
	}
}

func (c *Consumer) Send(like mq.Like) {
	c.ch <- like
}

func (c *Consumer) Close() {
	close(c.ch)
	c.close <- 1
}

func (c *Consumer) Consume() {
	tick := time.Tick(time.Second * 10)
	for {
		select {
		case msg := <-c.ch:
			c.AddToWindow(&msg)
		case <-tick:

		case <-c.close:
			return

		}
	}

}

func (c *Consumer) AddToWindow(msg *mq.Like) {
	if msg.Cancel == true {
		msg.LikeId = -msg.LikeId
	}
	_, ok := c.window[msg.TimeStamp]
	if !ok {
		c.window[msg.TimeStamp] = make(map[int64][]int64)
	}
	list, ok := c.window[msg.TimeStamp][msg.LikeId]
	if !ok {
		c.window[msg.TimeStamp][msg.LikeId] = make([]int64, 0)
		list = c.window[msg.TimeStamp][msg.LikeId]
	}
	list = append(list, msg.LikeId)
}

func (c *Consumer) UpdateCountTable()
