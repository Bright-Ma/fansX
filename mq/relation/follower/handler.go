package main

import (
	"context"
	"encoding/json"
	"errors"
	"fansX/common/lua"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
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
	executor *lua.Executor
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		msg := &mq.FollowingCanalJson{}
		err := json.Unmarshal(message.Value, msg)

		if err != nil {
			slog.Error(err.Error())
			continue
		}
		if len(msg.Data) == 0 {
			continue
		}

		data := &database.Follower{}
		data = Trans(&msg.Data[0])

		for i := 0; i < 3; i++ {
			err = h.process(data)

			if err != nil {
				slog.Error(err.Error(), "times", i+1)
				continue
			} else {
				session.MarkMessage(message, "")
				h.UpdateRedis(data)
				break
			}
		}

	}

	return nil
}

func (h *Handler) process(data *database.Follower) error {
	var err error

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	tx := h.db.WithContext(timeout).Begin()
	record := &database.Follower{}

	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(record, data.Id).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Commit()
		return err
	}

	if err != nil {
		err = tx.Create(data).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		err = h.UpdateNums(tx, data)
		if err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()
		return nil
	}

	if record.Type != data.Type && record.UpdatedAt < data.UpdatedAt {
		err = tx.Take(record).Update("type", data.Type).Update("updated_at", data.UpdatedAt).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		err = h.UpdateNums(tx, data)
		if err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()
		return nil
	}

	return nil
}
