package main

import (
	"encoding/json"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"log/slog"
	"strconv"
)

type Handler struct {
	db *gorm.DB
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	db := h.db
	for msg := range claim.Messages() {
		count := mq.LikeCountCdcJson{}
		err := json.Unmarshal(msg.Value, &count)
		if err != nil {
			slog.Error("unmarshal json to CommentCount:" + err.Error())
		}
		business, _ := strconv.ParseInt(count.Data[0].Business, 10, 64)
		if business != database.BusinessComment {
			session.MarkMessage(msg, "")
			continue
		}
		id, _ := strconv.ParseInt(count.Data[0].LikeId, 10, 64)
		now, _ := strconv.ParseInt(count.Data[0].Count, 10, 64)
		old, _ := strconv.ParseInt(count.Data[0].Count, 10, 64)
		err = db.Model(&database.Comment{}).Where("id = ?", id).Update("hot", gorm.Expr("hot + ?", now-old)).Error
		if err != nil {
			slog.Error("update comment count hot:" + err.Error())
			continue
		}
		session.MarkMessage(msg, "")
	}

	return nil
}
