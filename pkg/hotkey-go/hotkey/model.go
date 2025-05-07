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

// Subject 观察者模式
type Subject interface {
	register(ob Observer)
	notify(key string)
}

type Observer interface {
	Do(key string)
}

// Option 函数选项模式
type Option interface {
	Update(core *Core)
}

type OptionFunc func(core *Core)

type conn struct {
	mutex sync.Mutex
	conn  net.Conn
	// 该连接是否关闭
	closed *atomic.Bool
	addr   string
	core   *Core
	// 上次该连接接收到消息的时间戳
	last int64
}

type Core struct {
	// 本地缓存
	cache *freecache.Cache
	// hotkey缓存
	hotkeys *freecache.Cache
	// hotkey缓存时间
	ttl    int
	group  string
	client *etcd.Client
	// 连接集合
	conn cmap.ConcurrentMap[string, *conn]
	// 一致性hash表，据此进行key发送的路由
	hashMap *cshash.HashMap
	// 内存聚合缓冲channel
	send chan kv
	// 连接超时时间
	interval time.Duration

	observerList []Observer
}

type kv struct {
	key   string
	times int
}

// MsgStrategy 策略模式
type MsgStrategy interface {
	Handle(msg *model.ServerMessage, conn *conn)
}

type MsgPingStrategy struct {
}

type MsgPongStrategy struct {
}

type MsgAddStrategy struct {
}
