package server

import (
	group2 "fansX/pkg/hotkey-go/worker/group"
)

type context struct {
	conn  *group2.Conn
	group *group2.Group
}
