package server

import "bilibili/common/middleware/hotkeys/worker/group"

type context struct {
	conn  *group.Conn
	group *group.Group
}
