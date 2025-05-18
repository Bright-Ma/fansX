package main

import (
	"context"
	"encoding/json"
	"errors"
	bigcache "fansX/internal/middleware/cache"
	"fansX/internal/middleware/lua"
	"fansX/internal/middleware/lua/script/hash"
	"fansX/internal/model/mq"
	"fansX/mq/feed/lua"
	"fansX/services/content/public/proto/publicContentRpc"
	"fansX/services/relation/proto/relationRpc"
	"github.com/IBM/sarama"
	"github.com/avast/retry-go"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

type Handler struct {
	client              *redis.Client
	executor            *lua.Executor
	relationClient      relationRpc.RelationServiceClient
	publicContentClient publicContentRpc.PublicContentServiceClient
	cacheCreator        *bigcache.CacheCreator
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.FeedListKafkaJson{}
		err := json.Unmarshal(msg.Value, message)
		if err != nil {
			slog.Error("unmarshal json:" + err.Error())
			continue
		}
		h.process(message)
		session.MarkMessage(msg, "")
	}

	return nil
}

func (h *Handler) process(msg *mq.FeedListKafkaJson) {
	err := retry.Do(func() error {
		return h.cacheCreator.Update(msg.List)
	}, retry.Attempts(1000), retry.DelayType(retry.BackOffDelay), retry.MaxDelay(time.Second))

	if err != nil {
		panic(err.Error())
	}

	del, add := h.changeCompute(msg)
	wa := sync.WaitGroup{}
	wa.Add(2)
	go func() {
		h.handleAdd(add)
		time.Sleep(time.Minute * 5)
		h.handleAdd(add)
		wa.Done()
	}()
	go func() {
		h.handleDel(del)
		time.Sleep(time.Minute * 5)
		h.handleDel(del)
		wa.Done()
	}()
	wa.Wait()
}

func (h *Handler) changeCompute(msg *mq.FeedListKafkaJson) (del []int64, add []int64) {
	set := make(map[int64]bool)
	oSet := make(map[int64]bool)
	for _, v := range msg.List {
		set[v] = true
	}
	for _, v := range msg.OldList {
		oSet[v] = true
	}

	del = make([]int64, 0)
	add = make([]int64, 0)
	for key := range oSet {
		if set[key] == false {
			del = append(del, key)
		}
	}
	for key := range set {
		if oSet[key] == false {
			add = append(add, key)
		}
	}
	return del, add
}

func (h *Handler) handleDel(del []int64) (failed []int64) {
	failed = make([]int64, 0)

	for _, id := range del {
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		listResp, err := h.relationClient.ListAllFollower(timeout, &relationRpc.ListAllFollowerReq{
			UserId: id,
		})

		if err != nil {
			slog.Error("list del big follower" + err.Error())
			cancel()
			failed = append(failed, id)
			continue
		}

		getResp, err := h.publicContentClient.GetUserContentList(timeout, &publicContentRpc.GetUserContentListReq{
			Id:        id,
			TimeStamp: time.Now().Add(time.Hour).Unix(),
			Limit:     100,
		})

		if err != nil {
			slog.Error("get del big user content:" + err.Error())
			cancel()
			failed = append(failed, id)
			continue
		}

		mem := make([]string, len(getResp.Id))
		for i := 0; i < len(mem); i++ {
			mem[i] = strconv.FormatInt(id, 10) + ";" + strconv.FormatInt(getResp.Id[i], 10)
		}

		data := make([]interface{}, len(mem)*2)
		for i := 0; i < len(mem); i++ {
			data[i*2] = getResp.TimeStamp[i]
			data[i*2+1] = mem[i]
		}

		for _, v := range listResp.UserId {
			err = h.client.Get(timeout, "inbox:"+strconv.FormatInt(v, 10)).Err()
			if err != nil && !errors.Is(err, redis.Nil) {
				slog.Error("get user inbox:" + err.Error())
				continue
			} else if errors.Is(err, redis.Nil) {
				continue
			}
			err = h.executor.Execute(timeout, interlua.GetAdd(), []string{"inbox:" + strconv.FormatInt(v, 10), "100"}, data).Err()
			if err != nil {
				slog.Error("add content info to inbox:" + err.Error())
			}
		}

		cancel()
	}
	return failed
}

func (h *Handler) handleAdd(add []int64) (failed []int64) {
	failed = make([]int64, 0)
	for _, id := range add {
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		listResp, err := h.relationClient.ListAllFollower(timeout, &relationRpc.ListAllFollowerReq{
			UserId: id,
		})

		if err != nil {
			slog.Error("list add big follower:" + err.Error())
			cancel()
			failed = append(failed, id)
			continue
		}

		getResp, err := h.publicContentClient.GetUserContentList(timeout, &publicContentRpc.GetUserContentListReq{
			Id:        id,
			TimeStamp: time.Now().Add(time.Hour).Unix(),
			Limit:     100,
		})

		if err != nil {
			slog.Error("get add big user content:" + err.Error())
			failed = append(failed, id)
			cancel()
			continue
		}

		mem := make([]string, len(getResp.Id))
		for i := 0; i < len(mem); i++ {
			mem[i] = strconv.FormatInt(id, 10) + ";" + strconv.FormatInt(getResp.Id[i], 10)
		}

		for _, v := range listResp.UserId {
			err = h.client.Get(timeout, "inbox:"+strconv.FormatInt(v, 10)).Err()
			if err != nil && !errors.Is(err, redis.Nil) {
				slog.Error("get user online info" + err.Error())
				continue
			} else if errors.Is(err, redis.Nil) {
				continue
			}
			h.client.ZRem(timeout, "inbox:"+strconv.FormatInt(id, 10), mem)
		}
		cancel()
	}
	return failed
}
