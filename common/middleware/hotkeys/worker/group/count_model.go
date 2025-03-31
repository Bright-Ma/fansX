package group

import (
	"sync"
)

type count struct {
	mutex     sync.Mutex
	key       string
	lastTime  int64 //time,millisecond
	lastIndex int64
	lastSend  int64
	bucketId  int
	group     *Group
	deleted   bool
	window    [20]int64 //2second,window 100ms
	total     int64
}
