package main

import (
	"context"
	"encoding/json"
	"errors"
	"fansX/common/lua"
	bigcache "fansX/internal/middleware/cache"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	interlua "fansX/mq/relation/following/lua"
	"fansX/pkg/hotkey-go/hotkey"
	"fansX/services/content/public/proto/publicContentRpc"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log/slog"
	"strconv"
	"time"
)

type Handler struct {
	db                  *gorm.DB
	client              *redis.Client
	executor            *lua.Executor
	core                *hotkey.Core
	bigCache            *bigcache.Cache
	publicContentClient publicContentRpc.PublicContentServiceClient
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.FollowingCanalJson{}
		err := json.Unmarshal(msg.Value, msg)

		if err != nil {
			slog.Error(err.Error())
			continue
		}
		if len(message.Data) == 0 {
			continue
		}

		data := &database.Follower{}
		data = Trans(&message.Data[0])
		key1 := "Following:" + message.Data[0].FollowerId
		key2 := "FollowingNums:" + message.Data[0].FollowerId
		h.client.Del(context.Background(), key1, key2)
		h.core.SendDel(key2)

		for i := 0; i < 3; i++ {
			err = h.process(data)

			if err != nil {
				slog.Error(err.Error(), "times", i+1)
				continue
			} else {
				session.MarkMessage(msg, "")
				break
			}
		}

		// error process
		_ = h.UpdateInbox(data)

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

func (h *Handler) UpdateNums(tx *gorm.DB, data *database.Follower) error {
	if data.Type == database.Followed {
		return tx.Take(&database.FollowerNums{}, data.FollowingId).Update("nums", gorm.Expr("nums + 1")).Error
	} else {
		return tx.Take(&database.FollowerNums{}, data.FollowingId).Update("nums", gorm.Expr("nums - 1")).Error
	}
}

func (h *Handler) UpdateInbox(data *database.Follower) error {
	if data.Type == database.Followed {
		if !h.bigCache.IsBig(data.FollowingId) {
			resp, err := h.publicContentClient.GetUserContentList(context.Background(), &publicContentRpc.GetUserContentListReq{
				Id:        data.FollowingId,
				TimeStamp: time.Now().Add(time.Hour).Unix(),
				Limit:     100,
			})
			if err != nil {
				slog.Error("get user content list:"+err.Error(), "userId", data.FollowingId)
				return err
			}
			inbox := "inbox:" + strconv.FormatInt(data.FollowerId, 10)

			sm := make([]interface{}, len(resp.Id)*2)
			for i := 0; i < len(resp.Id); i++ {
				sm[i*2] = strconv.FormatInt(resp.TimeStamp[i], 10)
				sm[i*2+1] = strconv.FormatInt(data.FollowingId, 10) + ";" + strconv.FormatInt(resp.Id[i], 10)
			}

			err = h.executor.Execute(context.Background(), interlua.GetAdd(), []string{inbox, "100"}, sm).Err()
			if err != nil {
				slog.Error("add content to inbox:"+err.Error(), "followingId", data.FollowingId, "userId", data.FollowerId)
				return err
			}
		}
	} else {
		if !h.bigCache.IsBig(data.FollowingId) {
			inbox := "inbox:" + strconv.FormatInt(data.FollowerId, 10)
			prefix := strconv.FormatInt(data.FollowingId, 10)
			err := h.executor.Execute(context.Background(), interlua.GetDel(), []string{inbox}, prefix).Err()
			if err != nil {
				slog.Error("del unfollowing user content in inbox:"+err.Error(), "followingId", data.FollowingId)
				return err
			}
		}
	}
	return nil
}

func Trans(f *mq.Following) *database.Follower {
	t, _ := strconv.Atoi(f.Type)
	id, _ := strconv.ParseInt(f.Id, 10, 64)
	u, _ := strconv.ParseInt(f.UpdatedAt, 10, 64)
	followerId, _ := strconv.ParseInt(f.FollowerId, 10, 64)
	followingId, _ := strconv.ParseInt(f.FollowingId, 10, 64)

	return &database.Follower{
		Id:          id,
		FollowerId:  followerId,
		Type:        t,
		FollowingId: followingId,
		UpdatedAt:   u,
	}
}
