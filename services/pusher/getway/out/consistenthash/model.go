package cshash

import "sync"

type Map struct {
	nodes []virtualNode
}

type virtualNode struct {
	virtualKey string
	key        string
	value      uint64
}

var OldMap *Map
var NewMap *Map
var virtualNums int
var rmu sync.RWMutex
