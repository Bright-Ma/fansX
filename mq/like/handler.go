package main

import (
	"context"
	"encoding/json"
	"errors"
	"fansX/internal/middleware/lua"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	"fansX/internal/script/likeconsumerscript"
	leaf "fansX/pkg/leaf-go"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log/slog"
	"strconv"
	"time"
)

type Handler struct {
	db       *gorm.DB
	client   *redis.Client
	creator  leaf.Core
	consumer *Consumer
	executor *lua.Executor
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	h.consumer = NewConsumer(h.db)
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

		need, err := h.process(message)
		if err == nil {
			session.MarkMessage(msg, "")
			if need {
				h.consumer.Send(message)
			}
		}
	}

	return nil
}

func (h *Handler) process(message *mq.Like) (bool, error) {
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
		return false, err
	} else if err == nil {
		if record.UpdatedAt >= message.TimeStamp {
			tx.Commit()
			return false, nil
		}
		err = tx.Take(record).Update("status = ?", status).Update("updated_at", message.TimeStamp).Error
		if err != nil {
			tx.Rollback()
			slog.Error("update like record:" + err.Error())
			return false, err
		}
		tx.Commit()
		if status == record.Status {
			return false, nil
		}
		return true, nil
	}

	return true, h.CreateRecord(timeout, message, tx, status)

}

func (h *Handler) CreateRecord(timeout context.Context, message *mq.Like, tx *gorm.DB, status int) error {

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
	record := &database.Like{
		Id:        id,
		Business:  int(message.Business),
		Status:    status,
		UserId:    message.UserId,
		LikeId:    message.LikeId,
		UpdatedAt: message.TimeStamp,
	}

	err := tx.Create(record).Error
	if err != nil {
		tx.Rollback()
		slog.Error("create like record:" + err.Error())
		return err
	}
	tx.Commit()

	return nil
}

func (h *Handler) UpdateRedis(message *mq.Like) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	err := h.client.Del(timeout, "LikeNums:"+strconv.Itoa(int(message.Business))+":"+strconv.FormatInt(message.LikeId, 10)).Err()
	if err != nil {
		slog.Error("delete user like nums:" + err.Error())
	}

	err = h.client.Del(timeout, "UserLikeList:"+strconv.FormatInt(int64(message.Business), 10)+":"+strconv.FormatInt(message.UserId, 10)).Err()
	if err != nil {
		slog.Error("delete user like list:" + err.Error())
	}
	err = h.executor.Execute(timeout, likeconsumerscript.InsertScript, []string{
		"LikeList:" + strconv.Itoa(int(message.Business)) + ":" + strconv.FormatInt(message.LikeId, 10),
		"false",
	}, message.TimeStamp, message.UserId).Err()
	if err != nil {
		slog.Error("insert user to like list:" + err.Error())
	}
}
