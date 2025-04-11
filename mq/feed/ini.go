package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

func Init(client *redis.Client) (map[int64]bool, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var res map[string]string
	var err error

	for {
		select {
		case <-timeout.Done():
			return nil, timeout.Err()
		default:
			res, err = client.HGetAll(context.Background(), "BigHash").Result()
			if err != nil {
				time.Sleep(time.Millisecond * 50)
				continue
			} else {
				break
			}
		}
		break
	}

	keys := make([]string, 0)
	for k, v := range res {
		if k == "version" {
			continue
		}
		keys = append(keys, v)
	}

	set := make(map[int64]bool)
	for i := 0; i < len(keys); i++ {
		for {
			select {
			case <-timeout.Done():
				return nil, timeout.Err()
			default:
				members, err := client.SMembers(context.Background(), keys[i]).Result()
				if err != nil {
					time.Sleep(time.Millisecond * 50)
					continue
				} else {
					for _, m := range members {
						member, _ := strconv.ParseInt(m, 10, 64)
						set[member] = true
					}
					break
				}
			}
			break
		}
	}

	return set, nil

}
