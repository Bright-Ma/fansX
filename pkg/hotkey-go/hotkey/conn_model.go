package hotkey

import (
	"net"
	"sync"
	"sync/atomic"
)

type conn struct {
	mutex  sync.Mutex
	conn   net.Conn
	closed *atomic.Bool
	addr   string
	core   *Core
	last   int64
}
