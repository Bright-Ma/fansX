package main

import (
	"context"
	"encoding/json"
	"errors"
	bigcache "fansX/internal/middleware/cache"
	"fansX/internal/middleware/lua"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	"fansX/mq/content/script"
	leaf "fansX/pkg/leaf-go"
	"fansX/services/relation/proto/relationRpc"
	"github.com/IBM/sarama"
	"github.com/avast/retry-go"
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
	creator        leaf.Core
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.PublicContentCdcJson{}
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
		record := Translate(message)
		err = retry.Do(func() error {
			switch message.Type {
			case "INSERT":
				return h.handleInsert(record)
			case "UPDATE":
				if message.Data[0].Status == strconv.Itoa(database.ContentStatusPass) {
					return h.handleDelete(record)
				}
				return nil
			case "DELETE":
				return nil
			default:
				slog.Info("unsupported type", "type", message.Type)
				return nil
			}
		}, retry.Attempts(1000), retry.DelayType(retry.BackOffDelay), retry.MaxDelay(time.Second))
		if err == nil {
			session.MarkMessage(msg, "")
		}

	}
	return nil
}

func Translate(message *mq.PublicContentCdcJson) *database.VisibleContentInfo {
	if len(message.Data) == 0 {
		return nil
	}
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

// ToDo
func (h *Handler) handleDelete(record *database.VisibleContentInfo) error {
	h.client.Del(context.Background(), "ContentList:"+strconv.FormatInt(record.Id, 10))
	if err := h.UnLinkLike(record); err != nil {
		return err
	}
	if err := h.UnLinkComment(record); err != nil {
		return err
	}
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if h.bigCache.IsBig(record.Userid) {
		h.client.Del(context.Background(), "ContentList:"+strconv.FormatInt(record.Id, 10))
		slog.Info("is big user just update cache")
		return nil
	} else {
		listResp, err := h.relationClient.ListFollowing(timeout, &relationRpc.ListFollowingReq{
			UserId: record.Userid,
			All:    true,
		})
		if err != nil {
			slog.Error("get list following failed:" + err.Error())
			return err
		}
		member := strconv.FormatInt(record.Userid, 10) + ";" + strconv.FormatInt(record.Id, 10)
		for _, id := range listResp.UserId {
			key := "inbox:" + strconv.FormatInt(id, 10)
			err = h.executor.Execute(timeout, script.RemoveZSet, []string{key}, member).Err()
			if err != nil {
				slog.Error("del message in user inbox:" + err.Error())
				continue
			}
		}

	}
	return nil
}

func (h *Handler) handleInsert(record *database.VisibleContentInfo) error {
	if err := h.LinkLike(record); err != nil {
		return err
	}
	if err := h.LinkComment(record); err != nil {
		return err
	}
	is := h.bigCache.IsBig(record.Userid)
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if is {
		h.client.Del(timeout, "ContentList:"+strconv.FormatInt(record.Id, 10))
		slog.Info("is big user just update cache")
		return nil
	} else {
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
			err = h.executor.Execute(timeout, script.AddZSet, []string{key, "100"}, record.CreatedAt, member).Err()
			if err != nil {
				slog.Error("put message to user inbox:" + err.Error())
				continue
			}
		}
	}

	return nil
}

func (h *Handler) LinkLike(record *database.VisibleContentInfo) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx := h.db.WithContext(timeout).Begin()
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("business = ? and like_id = ?", database.BusinessContent, record.Id).Take(&database.LikeCount{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get like count record failed:" + err.Error())
		tx.Commit()
		return err
	} else if err == nil {
		slog.Info("like count record have been created")
		return nil
	}
	id, err := h.creator.GetIdWithContext(timeout)
	if err != nil {
		slog.Error("create like count id failed:" + err.Error())
		tx.Commit()
		return err
	}

	err = tx.Create(&database.LikeCount{
		Id:       id,
		Business: database.BusinessContent,
		LikeId:   record.Id,
		Status:   database.LikeCountStatusCommon,
		Count:    0,
	}).Error
	if err != nil {
		slog.Error("create like count failed:" + err.Error())
		tx.Rollback()
		return err
	}
	slog.Info("create like count success", "contentId", record.Id)
	tx.Commit()
	return nil
}

func (h *Handler) LinkComment(record *database.VisibleContentInfo) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx := h.db.WithContext(timeout).Begin()
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("business = ? and count_id = ?", database.BusinessContent, record.Id).Take(&database.CommentCount{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get comment count record failed:" + err.Error())
		tx.Commit()
		return err
	} else if err == nil {
		slog.Info("comment count record have been created")
		tx.Commit()
		return nil
	}
	id, err := h.creator.GetIdWithContext(timeout)
	if err != nil {
		slog.Error("create comment count id failed:" + err.Error())
		tx.Commit()
		return err
	}

	err = tx.Create(&database.CommentCount{
		Id:       id,
		Business: database.BusinessContent,
		CountId:  record.Id,
		Count:    0,
	}).Error
	if err != nil {
		slog.Error("create comment count failed:" + err.Error())
		tx.Rollback()
		return err
	}
	slog.Info("create comment count success", "contentId", record.Id)
	tx.Commit()
	return nil
}

func (h *Handler) UnLinkComment(record *database.VisibleContentInfo) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tx := h.db.WithContext(timeout).Begin()
	slog.Info("unlink comment count")

	err := tx.Where("business = ? and count_id = ?", database.BusinessComment, record.Id).Update("status", database.CommentCountStatusDelete).Error
	if err != nil {
		tx.Rollback()
		slog.Error("delete comment count failed:" + err.Error())
		return err
	}
	tx.Commit()
	return nil
}

func (h *Handler) UnLinkLike(record *database.VisibleContentInfo) error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tx := h.db.WithContext(timeout).Begin()
	slog.Info("unlink like count")

	err := tx.Where("business = ? and like_id = ?", database.BusinessComment, record.Id).Update("status", database.CommentCountStatusDelete).Error
	if err != nil {
		tx.Rollback()
		slog.Error("delete like count failed:" + err.Error())
		return err
	}
	tx.Commit()
	return nil
}
