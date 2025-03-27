package snowflake

import (
	"github.com/bwmarrin/snowflake"
	etcd "go.etcd.io/etcd/client/v3"
)

type Creator struct {
	name     string
	addr     string
	working  bool
	client   *etcd.Client
	snowNode *snowflake.Node
	lease    etcd.LeaseID
}
