package snowflake

import (
	"github.com/bwmarrin/snowflake"
	etcd "go.etcd.io/etcd/client/v3"
	"sync/atomic"
)

type Creator struct {
	name string
	addr string
	// 当前节点时候可正常工作，小步长回拨时将该值设置为false
	// 这里虽然使用原子变量，但不能保证并发安全
	working  atomic.Bool
	client   *etcd.Client
	snowNode *snowflake.Node
	lease    etcd.LeaseID
	// 是否只上报时钟到本地，在etcd不可用时将该变量设置为true
	local atomic.Bool
	// 本地时钟(millisecond)
	lastTime int64
}

type Config struct {
	CreatorName string
	Addr        string
	EtcdAddr    []string
}
