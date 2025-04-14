package cshash

import "sync"

type HashMap struct {
	old         []virtualNode
	new         []virtualNode
	rmu         sync.RWMutex
	virtualNums int
}

type virtualNode struct {
	virtualKey string
	key        string
	value      uint64
}
