package main

import (
	"bilibili/common/lua"
	luaHash "bilibili/common/lua/script/hash"
	"bilibili/internal/model/mq"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
	"time"
)

type Handler struct {
	client   *redis.Client
	cache    map[int64]bool
	executor *lua.Executor
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
	if len(msg.List) == 0 {
		return
	}
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

	var res map[string]string
	var err error
	for {
		res, err = h.client.HGetAll(context.Background(), "BigHash").Result()
		if err != nil {
			slog.Error("get hash err:" + err.Error())
		} else {
			break
		}
	}

	for _, v := range res {
		for {
			err = h.client.Del(context.Background(), v).Err()
			if err != nil {
				slog.Error("del err:"+err.Error(), "key", v)
			} else {
				break
			}
		}
	}

	for k, v := range set {
		for {
			err = h.client.SAdd(context.Background(), k, v).Err()
			if err != nil {
				slog.Error("add err:"+err.Error(), "key", k)
			} else {
				break
			}
		}
	}

	create := make([]string, len(key)*2)
	for i := 0; i < len(create); i += 2 {
		create[i] = key[i/2]
		create[i+1] = key[i/2]
	}

	create = append(create, "version", strconv.FormatInt(time.Now().Unix(), 10))
	for {
		err = h.executor.Execute(context.Background(), luaHash.GetCreate(), []string{"BigHash", "true", "864000"}, create).Err()
		if err != nil {
			slog.Error("execute hash create:" + err.Error())
		} else {
			break
		}
	}

}
