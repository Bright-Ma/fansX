package main

import (
	"context"
	"encoding/json"
	"errors"
	bigcache "fansX/internal/middleware/cache"
	"fansX/internal/middleware/lua"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	"fansX/mq/relation/script"
	"fansX/services/content/public/proto/publicContentRpc"
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
	db                  *gorm.DB
	client              *redis.Client
	executor            *lua.Executor
	bigCache            *bigcache.Cache
	publicContentClient publicContentRpc.PublicContentServiceClient
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim 缓存更新，inbox更新，table更新
func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.FollowingCdcJson{}
		err := json.Unmarshal(msg.Value, message)

		if err != nil {
			slog.Error(err.Error())
			continue
		}
		if len(message.Data) == 0 || message.IsDdl {
			slog.Info("receive message:", "len data", len(message.Data), "is ddl", message.IsDdl)
			continue
		}

		data := &database.Follower{}
		data = Trans(&message.Data[0])
		key1 := "Following:" + message.Data[0].FollowerId
		key2 := "FollowingNums:" + message.Data[0].FollowerId
		h.client.Del(context.Background(), key1, key2)

		times := 0
		// 无限重试
		_ = retry.Do(func() error {
			err = h.process(data)
			if errors.Is(err, ErrNeedNotConsume) {
				slog.Info("message have been consumed or have been consumed latest message")
				return nil
			} else if err != nil {
				slog.Error(err.Error(), "times", times+1)
				times++
				return err
			}
			// 有限重试
			_ = retry.Do(func() error {
				return h.UpdateInbox(data)
			})
			return nil
		}, retry.Attempts(1000), retry.DelayType(retry.BackOffDelay), retry.MaxDelay(time.Second))
		session.MarkMessage(msg, "")

	}

	return nil
}

// process following->follower同步
func (h *Handler) process(data *database.Follower) error {
	var err error

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	tx := h.db.WithContext(timeout).Begin()
	record := &database.Follower{}
	// 查找关系记录
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(record, data.Id).Error
	// 错误
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Commit()
		return err
	}
	// 未找到记录，创建
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
	// 记录version<消息version并且状态不相同
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

	return ErrNeedNotConsume
}

// UpdateNums 计数表更新
func (h *Handler) UpdateNums(tx *gorm.DB, data *database.Follower) error {
	if data.Type == database.Followed {
		return tx.Take(&database.FollowerNums{}, data.FollowingId).Update("nums", gorm.Expr("nums + 1")).Error
	} else {
		return tx.Take(&database.FollowerNums{}, data.FollowingId).Update("nums", gorm.Expr("nums - 1")).Error
	}
}

// UpdateInbox inbox补偿
func (h *Handler) UpdateInbox(data *database.Follower) error {
	// 关注
	if data.Type == database.Followed {
		// 不是大v，向inbox中添加作品
		if !h.bigCache.IsBig(data.FollowingId) {
			// 作品列表获取
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
			// inbox更新
			err = h.executor.Execute(context.Background(), script.InsertZSetWithMa, []string{inbox, "100"}, sm).Err()
			if err != nil {
				slog.Error("add content to inbox:"+err.Error(), "followingId", data.FollowingId, "userId", data.FollowerId)
				return err
			}
		}
		// 取关
	} else {
		// 不是大v，inbox中移除作品
		if !h.bigCache.IsBig(data.FollowingId) {
			inbox := "inbox:" + strconv.FormatInt(data.FollowerId, 10)
			// inbox中member模式为{userId};{contentId}
			prefix := strconv.FormatInt(data.FollowingId, 10)
			err := h.executor.Execute(context.Background(), script.RemoveZSet, []string{inbox}, prefix).Err()
			if err != nil {
				slog.Error("del unfollowing user content in inbox:"+err.Error(), "followingId", data.FollowingId)
				return err
			}
		}
	}
	return nil
}

func Trans(f *mq.FollowingCdc) *database.Follower {
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

var ErrNeedNotConsume = errors.New("need not consume")
