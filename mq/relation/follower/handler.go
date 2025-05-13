package main

import (
	"context"
	"encoding/json"
	"fansX/internal/middleware/lua"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	"fansX/mq/relation/script"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
	"time"
)

type Handler struct {
	client   *redis.Client
	executor *lua.Executor
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.FollowerCdcJson{}
		err := json.Unmarshal(msg.Value, message)
		if err != nil {
			slog.Info("receive message:", "len data", len(message.Data), "is ddl", message.IsDdl)
			slog.Error("unmarshal json:" + err.Error())
			continue
		}
		if len(message.Data) == 0 || message.IsDdl {
			continue
		}
		data := Trans(&message.Data[0])
		h.UpdateRedis(data)
		session.MarkMessage(msg, "")
	}

	return nil
}

func Trans(msg *mq.FollowerCdc) *database.Follower {
	t, _ := strconv.Atoi(msg.Type)
	id, _ := strconv.ParseInt(msg.Id, 10, 64)
	u, _ := strconv.ParseInt(msg.UpdatedAt, 10, 64)
	followerId, _ := strconv.ParseInt(msg.FollowerId, 10, 64)
	followingId, _ := strconv.ParseInt(msg.FollowingId, 10, 64)

	return &database.Follower{
		Id:          id,
		FollowerId:  followerId,
		Type:        t,
		FollowingId: followingId,
		UpdatedAt:   u,
	}
}

// UpdateRedis 增量更新redis
func (h *Handler) UpdateRedis(data *database.Follower) {
	e := h.executor
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := "Follower:" + strconv.FormatInt(data.FollowingId, 10)
	if data.Type == database.Followed {
		err := e.Execute(timeout, script.InsertZSet, []string{key}, strconv.FormatInt(data.UpdatedAt, 10), strconv.FormatInt(data.FollowerId, 10)).Err()
		if err != nil {
			slog.Error("insert user to follower list in redis:" + err.Error())
		}

		err = e.Execute(timeout, script.IncrBy, []string{"FollowerNums:" + strconv.FormatInt(data.FollowingId, 10)}, 1).Err()
		if err != nil {
			slog.Error("add follower nums to redis:" + err.Error())
		}
	} else {
		err := h.client.ZRem(timeout, key, data.FollowerId).Err()
		if err != nil {
			slog.Error("del user to follower list in redis:" + err.Error())
		}

		err = e.Execute(timeout, script.IncrBy, []string{"FollowerNums:" + strconv.FormatInt(data.FollowingId, 10)}, -1).Err()
		if err != nil {
			slog.Error("sub follower nums to redis:" + err.Error())
		}
	}
}
