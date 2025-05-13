package bigcache

import (
	etcd "go.etcd.io/etcd/client/v3"
)

type CacheCreator struct {
	client *etcd.Client
}

type BigIndex struct {
	Key []string `json:"key"`
}

type BigSet struct {
	Id []int64 `json:"id"`
}
