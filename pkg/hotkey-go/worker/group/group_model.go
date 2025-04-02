package group

import (
	cmap "github.com/orcaman/concurrent-map/v2"
	"sync"
)

type Group struct {
	idMutex    sync.Mutex
	bucketSize int
	countMutex []sync.Mutex
	// 对每个连接以一个唯一id进行标识，该id只在单机使用，故没有使用leaf包
	nextId int64
	// 对key进行分桶
	bucket []cmap.ConcurrentMap[string, *count]

	connectionSet cmap.ConcurrentMap[string, *Conn]
}

var groupMap cmap.ConcurrentMap[string, *Group]
