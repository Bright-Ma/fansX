package main

import (
	"context"
	"encoding/json"
	bigcache "fansX/internal/middleware/cache"
	"fansX/internal/middleware/lua"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	interlua "fansX/mq/content/public/lua"
	"fansX/services/relation/proto/relationRpc"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log/slog"
	"strconv"
	"time"
)

type Handler struct {
	db             *gorm.DB
	bigCache       *bigcache.Cache
	client         *redis.Client
	relationClient relationRpc.RelationServiceClient
	executor       *lua.Executor
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.PublicContentJson{}
		err := json.Unmarshal(msg.Value, message)
		if err != nil {
			slog.Error("unmarshal json:" + err.Error())
			continue
		}
		if message.IsDdl {
			session.MarkMessage(msg, "")
			continue
		}
		slog.Info("consume kafka message", "dml_type", message.Type)
		// need to update minio text db...
		if message.Type == "INSERT" {
			record := Translate(message)
			err = h.handleInsert(record)
			if err == nil {
				session.MarkMessage(msg, "")
			}
		} else if message.Type == "UPDATE" {
			if message.Data[0].Status == strconv.Itoa(database.ContentStatusDelete) {
				record := Translate(message)
				err = h.handleDelete(record)
				if err == nil {
					session.MarkMessage(msg, "")
				}
			} else {
				slog.Info("update message is not delete")
			}
		} else if message.Type == "DELETE" {
			slog.Info("delete message")
			session.MarkMessage(msg, "")
		} else {
			slog.Error("un supported type")
		}
	}
	return nil
}

func Translate(message *mq.PublicContentJson) *database.VisibleContentInfo {
	id, _ := strconv.ParseInt(message.Data[0].Id, 10, 64)
	version, _ := strconv.ParseInt(message.Data[0].Version, 10, 64)
	status, _ := strconv.ParseInt(message.Data[0].Status, 10, 64)
	createdAt, _ := strconv.ParseInt(message.Data[0].CreatedAt, 10, 64)
	userId, _ := strconv.ParseInt(message.Data[0].UserId, 10, 64)
	record := &database.VisibleContentInfo{
		Id:           id,
		Version:      version,
		Userid:       userId,
		Status:       int(status),
		Title:        message.Data[0].Title,
		PhotoUriList: message.Data[0].PhotoUriList,
		ShortText:    message.Data[0].ShortTextUri,
		LongTextUri:  message.Data[0].LongTextUri,
		VideoUriList: message.Data[0].VideoUriList,
		CreatedAt:    createdAt,
	}

	return record

}

func (h *Handler) handleDelete(record *database.VisibleContentInfo) error {
	h.client.Del(context.Background(), "ContentList:"+strconv.FormatInt(record.Id, 10))
	h.db.Delete(&database.LikeCount{}, "business = ? and like_id = ?", database.BusinessContent, record.Id)
	return nil
}

func (h *Handler) handleInsert(record *database.VisibleContentInfo) error {
	is := h.bigCache.IsBig(record.Userid)
	if is {
		h.client.Del(context.Background(), "ContentList:"+strconv.FormatInt(record.Id, 10))
		slog.Info("is big user just update cache")
		return nil
	} else {
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		listResp, err := h.relationClient.ListFollowing(timeout, &relationRpc.ListFollowingReq{
			UserId: record.Userid,
			All:    true,
		})
		if err != nil {
			slog.Error("get user following list:" + err.Error())
			return err
		}
		member := strconv.FormatInt(record.Userid, 10) + ";" + strconv.FormatInt(record.Id, 10)
		for _, id := range listResp.UserId {
			key := "inbox:" + strconv.FormatInt(id, 10)
			err = h.executor.Execute(timeout, interlua.GetAdd(), []string{key, "100"}, record.CreatedAt, member).Err()
			if err != nil {
				slog.Error("put message to user inbox:" + err.Error())
				continue
			}
		}
	}

	return nil
}

func (h *Handler) CreateLikeRecord(record *database.VisibleContentInfo) error {
	h.db.Model(&database.LikeCount{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("")
}
