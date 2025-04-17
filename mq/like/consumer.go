package main

import (
	"encoding/json"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	leaf "fansX/pkg/leaf-go"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"log/slog"
	"time"
)

type Consumer struct {
	session sarama.ConsumerGroupSession
	db      *gorm.DB
	creator leaf.Core
	ch      chan SendMessage
	close   chan int
	window  map[int64]map[[2]int64][]int64
	msgSet  []*sarama.ConsumerMessage
}

type SendMessage struct {
	KafkaMessage *sarama.ConsumerMessage
	Like         *mq.Like
}

func NewConsumer(session sarama.ConsumerGroupSession, db *gorm.DB) *Consumer {
	ch := make(chan SendMessage, 1024*1024)
	return &Consumer{
		db:      db,
		session: session,
		ch:      ch,
		close:   make(chan int, 1),
		window:  make(map[int64]map[[2]int64][]int64),
		msgSet:  make([]*sarama.ConsumerMessage, 0),
	}
}

func (c *Consumer) Send(like *mq.Like, msg *sarama.ConsumerMessage) {
	c.ch <- SendMessage{
		KafkaMessage: msg,
		Like:         like,
	}
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
			c.AddToWindow(msg.Like)
			c.msgSet = append(c.msgSet, msg.KafkaMessage)
		case <-tick:
			err := c.UpdateCountTable()
			if err == nil {
				c.MarkMessage()
			}
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
		c.window[msg.TimeStamp] = make(map[[2]int64][]int64)
	}
	index := [2]int64{int64(msg.Business), msg.LikeId}
	list, ok := c.window[msg.TimeStamp][index]
	if !ok {
		c.window[msg.TimeStamp][index] = make([]int64, 0)
		list = c.window[msg.TimeStamp][index]
	}
	list = append(list, msg.UserId)
}

func (c *Consumer) UpdateCountTable() error {
	for window, list := range c.window {
		var id int64
		var ok bool
		for {
			id, ok = c.creator.GetId()
			if !ok {
				time.Sleep(time.Millisecond * 50)
				continue
			}
			break
		}
		record := database.TimeWindow{
			Id:     id,
			Window: window,
			Body:   nil,
		}
		body := database.WindowBody{
			Like: make(map[[2]int64][]int64),
		}

		for index, user := range list {
			body.Like[index] = user
		}
		b, err := json.Marshal(body)
		if err != nil {
			slog.Error("marshal json:" + err.Error())
			return err
		}
		record.Body = b
		tx := c.db.Begin()

		err = tx.Create(record).Error
		if err != nil {
			tx.Rollback()
			slog.Error("create window record:" + err.Error())
			return err
		}
		for like, user := range body.Like {
			if like[1] < 0 {
				like[1] = -like[1]
				err = tx.Where("business = ? and like_id = ?", like[0], like[1]).
					Update("count", gorm.Expr("count - ?", len(user))).Error
				if err != nil {
					slog.Error("update like count table(-):" + err.Error())
					tx.Rollback()
					return err
				}
			} else {
				err = tx.Where("business = ? and like_id = ?", like[0], like[1]).
					Update("count", gorm.Expr("count + ?", len(user))).Error
				if err != nil {
					slog.Error("update like count table(+):" + err.Error())
					tx.Rollback()
					return err
				}
			}
		}
		tx.Commit()
	}

	return nil

}

func (c *Consumer) MarkMessage() {
	for _, msg := range c.msgSet {
		c.session.MarkMessage(msg, "")
	}
}
