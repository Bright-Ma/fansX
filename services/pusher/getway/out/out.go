package main

import (
	svc "OutGetWay/context"
	"OutGetWay/route"
)

// 外部网关，路由到内部pusher节点
func main() {
	ctx := svc.NewContext("./config/config.yaml")
	route.Init(ctx)
}
