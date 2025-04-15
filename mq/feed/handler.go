package main

import (
	"context"
	"encoding/json"
	"errors"
	"fansX/common/lua"
	luaHash "fansX/common/lua/script/hash"
	"fansX/internal/model/mq"
	"fansX/mq/feed/lua"
	leaf "fansX/pkg/leaf-go"
	"fansX/services/content/public/proto/publicContentRpc"
	"fansX/services/relation/proto/relationRpc"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
	"time"
)

type Handler struct {
	client              *redis.Client
	executor            *lua.Executor
	creator             leaf.Core
	relationClient      relationRpc.RelationServiceClient
	publicContentClient publicContentRpc.PublicContentServiceClient
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.FeedListJson{}
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

func (h *Handler) process(msg *mq.FeedListJson) {
	var size int
	if len(msg.List)%2000 == 0 {
		size = len(msg.List) / 2000
	} else {
		size = len(msg.List)/2000 + 1
	}

	set := make(map[string][]interface{})
	key := make([]string, size)
	for i := 0; i < size; i++ {
		key[i] = "BigSet:" + strconv.Itoa(i)
		set["BigSet:"+strconv.Itoa(i)] = make([]interface{}, 0)
	}
	for i := 0; i < len(msg.List); i++ {
		set[key[i%size]] = append(set[key[i%size]], strconv.FormatInt(msg.List[i], 10))
	}
	del, add := h.changeCompute(msg)
	h.handleDel(del)
	for i := 1; i <= 10; i++ {
		err := h.updateRedis(set, key)
		if err != nil && i != 10 {
			time.Sleep(time.Millisecond * 100 * time.Duration(i*i))
		} else if err != nil {
			panic("update redis failed")
		}
	}
	go func() {
		time.Sleep(time.Minute * 5)
		h.handleAdd(add)
	}()

	go func() {
		time.Sleep(time.Minute * 5)
		h.handleDel(del)
	}()
}

func (h *Handler) changeCompute(msg *mq.FeedListJson) (del []int64, add []int64) {
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
			err = h.client.Get(timeout, "user:"+strconv.FormatInt(v, 10)).Err()
			if err != nil && !errors.Is(err, redis.Nil) {
				slog.Error("get user online info:" + err.Error())
				continue
			} else if errors.Is(err, redis.Nil) {
				continue
			}
			h.executor.Execute(timeout, interlua.GetAdd(), []string{"inbox:" + strconv.FormatInt(v, 10), "100"}, data)
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
			err = h.client.Get(timeout, "user:"+strconv.FormatInt(v, 10)).Err()
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

func (h *Handler) updateRedis(set map[string][]interface{}, key []string) error {

	res := make(map[string]string)
	var err error
	res, err = h.client.HGetAll(context.Background(), "BigHash").Result()
	if err != nil {
		slog.Error("get hash err:" + err.Error())
		return err
	}

	for _, v := range res {
		err = h.client.Del(context.Background(), v).Err()
		if err != nil {
			slog.Error("del err:"+err.Error(), "key", v)
			return err
		}
	}

	for k, v := range set {
		err = h.client.SAdd(context.Background(), k, v).Err()
		if err != nil {
			slog.Error("add err:"+err.Error(), "key", k)
			return err
		}
	}

	create := make([]string, len(key)*2)
	for i := 0; i < len(create); i += 2 {
		create[i] = key[i/2]
		create[i+1] = key[i/2]
	}
	id, ok := h.creator.GetId()
	if !ok {
		err = errors.New("id creator is not available")
		slog.Error(err.Error())
		return err
	}

	create = append(create, "version", strconv.FormatInt(id, 10))
	err = h.executor.Execute(context.Background(), luaHash.GetCreate(), []string{"BigHash", "true", "864000"}, create).Err()
	if err != nil {
		slog.Error("execute hash create:" + err.Error())
		return err
	}

	return nil
}
