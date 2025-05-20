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
	"log/slog"
	"strconv"
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
	consumer := NewConsumer(h.db)
	for msg := range claim.Messages() {
		message := mq.CommentKafkaJson{}
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
		message := mq.DelCommentKafkaJson{}
		_ = json.Unmarshal(msg.Value, &message)
		err := h.DelComment(&message)
		if err != nil {
			continue
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *Handler) DelComment(message *mq.DelCommentKafkaJson) error {
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
	err = tx.Take(&record).Update("status", database.ContentStatusDelete).Error
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
	h.DelUpdateRedis(timeout, &record)
	return nil
}

type ListRecord struct {
	CommentId   int64  `json:"comment_id"`
	UserId      int64  `json:"user_id"`
	ContentId   int64  `json:"content_id"`
	RootId      int64  `json:"root_id"`
	ParentId    int64  `json:"parent_id"`
	CreatedAt   int64  `json:"created_at"`
	ShortText   string `json:"short_text"`
	LongTextUri string `json:"long_text_uri"`
}

func (h *Handler) DelUpdateRedis(ctx context.Context, record *database.Comment) {
	key := "CommentListByTime:" + strconv.FormatInt(record.ContentId, 10)
	member := ListRecord{
		CommentId:   record.Id,
		UserId:      record.UserId,
		ContentId:   record.ContentId,
		RootId:      record.RootId,
		ParentId:    record.ParentId,
		CreatedAt:   record.CreatedAt,
		ShortText:   record.ShortText,
		LongTextUri: record.LongTextUri,
	}
	m, err := json.Marshal(member)
	if err != nil {
		slog.Error("marshal member json:" + err.Error())
		return
	}
	err = h.client.ZRem(ctx, key, string(m)).Err()
	if err != nil {
		slog.Error("delete member in redis ZSet:" + err.Error())
	}
}
