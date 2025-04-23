package main

import (
	"context"
	"encoding/json"
	"errors"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

type Handler struct {
	db     *gorm.DB
	client *redis.Client
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	if claim.Topic() == "" {
		return h.CommentHandler(session, claim)

	}
	return h.CommentHandler(session, claim)
}

func (h *Handler) CommentHandler(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	consumer := NewConsumer(h.db, h.client)
	for msg := range claim.Messages() {
		message := mq.CommentKafkaMsg{}
		_ = json.Unmarshal(msg.Value, &message)
		record := database.Comment{
			Id:          message.Id,
			UserId:      message.UserId,
			ContentId:   message.ContentId,
			RootId:      message.RootId,
			Status:      database.CommentStatusCommon,
			Hot:         0,
			ParentId:    message.ParentId,
			UpdatedAt:   time.Time{},
			ShortText:   message.ShortText,
			LongTextUri: message.LongTextUri,
		}
		session.MarkMessage(msg, "")
		consumer.Send(&record)
	}
	return nil
}

func (h *Handler) DelHandler(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := mq.DelCommentKafkaMsg{}
		_ = json.Unmarshal(msg.Value, &message)
		err := h.DelComment(&message)
		if err != nil {
			continue
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *Handler) DelComment(message *mq.DelCommentKafkaMsg) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	tx := h.db.WithContext(timeout)
	record := database.Comment{}
	err := tx.Take(&record, message.CommentId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		record.Status = database.ContentStatusDelete
		err := tx.Create(record).Error
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	}
	if err != nil {
		tx.Commit()
		return err
	}
	if record.Status == database.ContentStatusDelete {
		tx.Commit()
		return nil
	}
	err = tx.Take(record).Update("status", database.ContentStatusDelete).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&database.CommentCount{}).
		Where("business_id = ? and count_id = ?", database.BusinessContent, record.ContentId).
		Update("count", gorm.Expr("count_id - 1")).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	if record.RootId != 0 {
		err = tx.Model(&database.CommentCount{}).
			Where("business_id = ? and count_id = ?", database.BusinessComment, record.RootId).
			Update("count", gorm.Expr("count_id - 1")).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
