package hotkey

import (
	cshash "fansX/pkg/consistenthash"
	"fansX/pkg/hotkey-go/model"
	"github.com/coocood/freecache"
	cmap "github.com/orcaman/concurrent-map/v2"
	etcd "go.etcd.io/etcd/client/v3"
)

type Core struct {
	cache    *freecache.Cache
	hotkeys  *freecache.Cache
	group    string
	delGroup string
	client   *etcd.Client
	conn     cmap.ConcurrentMap[string, *conn]
	hashMap  *cshash.HashMap

	send       chan kv
	del        chan string
	hot        chan string
	addProcess func(c *Core, message *model.ServerMessage)
	delProcess func(c *Core, message *model.ServerMessage)
}

type Config struct {
	Model        string
	GroupName    string
	DelGroupName string
	CacheSize    int
	HotKeySize   int
	EtcdConfig   etcd.Config

	DelChan chan string
	HotChan chan string
}

type kv struct {
	key   string
	times int
}

var (
	ModelConsumer = "consumer"
	ModelCache    = "cache"
)
