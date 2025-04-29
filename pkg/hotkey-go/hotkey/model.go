package hotkey

import (
	cshash "fansX/pkg/consistenthash"
	"fansX/pkg/hotkey-go/model"
	"github.com/coocood/freecache"
	cmap "github.com/orcaman/concurrent-map/v2"
	etcd "go.etcd.io/etcd/client/v3"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Subject interface {
	register(ob Observer)
	notify(key string)
}

type Observer interface {
	Do(key string)
}

type Option interface {
	Update(core *Core)
}

type OptionFunc func(core *Core)

type conn struct {
	mutex  sync.Mutex
	conn   net.Conn
	closed *atomic.Bool
	addr   string
	core   *Core
	last   int64
}

type Core struct {
	cache   *freecache.Cache
	hotkeys *freecache.Cache
	ttl     int
	group   string
	client  *etcd.Client
	conn    cmap.ConcurrentMap[string, *conn]
	hashMap *cshash.HashMap

	send     chan kv
	interval time.Duration

	observerList []Observer
}

type kv struct {
	key   string
	times int
}

type MsgStrategy interface {
	Handle(msg *model.ServerMessage, conn *conn)
}

type MsgPingStrategy struct {
}

type MsgPongStrategy struct {
}

type MsgAddStrategy struct {
}
