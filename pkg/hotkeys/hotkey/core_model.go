package hotkey

import (
	cshash "bilibili/pkg/consistenthash"
	"github.com/coocood/freecache"
	cmap "github.com/orcaman/concurrent-map/v2"
	etcd "go.etcd.io/etcd/client/v3"
)

type Core struct {
	cache   *freecache.Cache
	hotkeys *freecache.Cache
	group   string
	client  *etcd.Client
	conn    cmap.ConcurrentMap[string, *conn]
	hashMap *cshash.HashMap

	send chan kv
}

type kv struct {
	key   string
	times int
}
