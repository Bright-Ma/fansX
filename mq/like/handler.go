package main

import (
	"context"
	"encoding/json"
	"errors"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	leaf "fansX/pkg/leaf-go"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log/slog"
	"time"
)

type Handler struct {
	db       *gorm.DB
	client   *redis.Client
	creator  leaf.Core
	consumer *Consumer
}

func (h *Handler) Setup(session sarama.ConsumerGroupSession) error {
	h.consumer = NewConsumer(session, h.db)
	go h.consumer.Consume()
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	h.consumer.Close()
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.Like{}
		err := json.Unmarshal(msg.Value, message)
		if err != nil {
			slog.Error("unmarshal json:" + err.Error())
			continue
		}

		err = h.process(message)
		if errors.Is(err, ErrNeedNotConsume) {
			session.MarkMessage(msg, "")
			continue
		} else if err != nil {
			continue
		}

	}

	return nil
}

func (h *Handler) process(message *mq.Like) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	tx := h.db.WithContext(timeout).Begin()
	var err error
	var status int
	if message.Cancel == true {
		status = database.LikeStatusUnlike
	} else {
		status = database.LikeStatusLike
	}
	record := &database.Like{}
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("business = ? and status in (0,1) and user_id = ? and like_id = ?", message.Business, message.UserId, message.LikeId).Take(record).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("search like record from tidb:" + err.Error())
		tx.Commit()
		return err
	} else if err == nil {
		if record.Status == status || record.UpdatedAt > message.TimeStamp {
			tx.Commit()
			return ErrNeedNotConsume
		}
		err = tx.Take(record).Update("status = ?", status).Update("updated_at", message.TimeStamp).Error
		if err != nil {
			tx.Rollback()
			slog.Error("update like record:" + err.Error())
		}
		return err
	}

	creator := h.creator
	var id int64
	var ok bool

	for {
		select {
		case <-timeout.Done():
			slog.Error("get id timeout")
			return context.DeadlineExceeded
		default:
			id, ok = creator.GetId()
			if !ok {
				time.Sleep(time.Millisecond * 100)
				continue
			}
		}
		break
	}
	record = &database.Like{
		Id:        id,
		Business:  int(message.Business),
		Status:    status,
		UserId:    message.UserId,
		LikeId:    message.LikeId,
		UpdatedAt: message.TimeStamp,
	}

	err = tx.Create(record).Error
	if err != nil {
		tx.Rollback()
		slog.Error("create like record:" + err.Error())
		return err
	}

	return nil
}

var (
	ErrNeedNotConsume = errors.New("need not consume")
)
