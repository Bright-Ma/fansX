package connection

import (
	"github.com/panjf2000/gnet"
	"sync"
)

type Conn struct {
	id   string
	last int64
	conn gnet.Conn
}

var idMutex sync.Mutex
var nextId int64
