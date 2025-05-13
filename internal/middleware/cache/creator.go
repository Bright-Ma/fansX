package bigcache

import (
	"context"
	"encoding/json"
	etcd "go.etcd.io/etcd/client/v3"
	"strconv"
	"time"
)

func NewCacheCreator(client *etcd.Client) *CacheCreator {
	return &CacheCreator{
		client: client,
	}
}

func (creator *CacheCreator) Update(id []int64) error {
	i := 0
	set := make([]BigSet, 0)
	for i < len(id) {
		set = append(set, BigSet{Id: id[i:min(len(id)-1, i+999)]})
		i += 1000
	}
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err := creator.client.Delete(timeout, "BigSet/", etcd.WithPrefix())
	if err != nil {
		return err
	}
	index := BigIndex{Key: make([]string, len(set))}
	for j, s := range set {
		index.Key[j] = "BigSet/" + strconv.FormatInt(int64(j), 10)
		value, _ := json.Marshal(s)
		_, err = creator.client.Put(timeout, "BigSet/"+strconv.FormatInt(int64(j), 10), string(value))
		if err != nil {
			return err
		}
	}
	value, _ := json.Marshal(index)
	_, err = creator.client.Put(timeout, "BigIndex/0", string(value))
	if err != nil {
		return err
	}
	return nil

}
