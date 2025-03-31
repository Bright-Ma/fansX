package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	WatchPrefix string
	MRedis      struct {
		Host string
		DB   int
	}
	NSQ struct {
		Addr  string
		Topic string
	}
	Model       int
	VirtualNums int
}
