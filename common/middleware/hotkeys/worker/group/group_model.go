package group

import (
	cmap "github.com/orcaman/concurrent-map/v2"
	"sync"
)

type Group struct {
	idMutex    sync.Mutex
	countMutex []sync.Mutex
	nextId     int64
	bucket     []cmap.ConcurrentMap[string, *count]

	connectionSet cmap.ConcurrentMap[string, *Conn]
}

var groupMap cmap.ConcurrentMap[string, *Group]
