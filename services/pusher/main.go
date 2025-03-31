package main

import (
	cshash "puhser/consistenthash"
	svc "puhser/internal/context"
	"puhser/mq"
	"puhser/route"
	"puhser/service"
)

func main() {
	ctx := svc.NewContext("./internal/config/config.yaml")
	cshash.Init(ctx.Config.VirtualNums)
	go service.Init(ctx)
	mq.Init(ctx)
	route.Init(ctx)
}
