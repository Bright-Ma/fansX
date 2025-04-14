package group

import (
	"sync"
)

type count struct {
	mutex sync.Mutex
	key   string
	// 上次访问该key的时间戳(millisecond)
	lastTime int64
	// 上次访问该key的时间窗口
	lastIndex int64
	// 上次发送的时间，发送过一次，一段时间内不在发送
	lastSend int64
	bucketId int
	group    *Group
	//删除标记
	deleted bool
	// 滑动窗口，保存最近2s的访问情况，每个窗口100ms
	window [20]int64
	// 2s窗口内总共的访问次数
	total int64
}
