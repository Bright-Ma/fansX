package group

import (
	"fansX/pkg/hotkey-go/worker/config"
	"fansX/pkg/hotkey-go/worker/connection"
	"fansX/pkg/hotkey-go/worker/window"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type group struct {
	config        *config.Config
	keys          cmap.ConcurrentMap[string, *window.Window]
	connectionSet cmap.ConcurrentMap[*connection.Conn, bool]
}

var groupMap *Map

type Map struct {
	groups cmap.ConcurrentMap[string, *group]
}
