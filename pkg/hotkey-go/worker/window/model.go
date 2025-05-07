package window

import (
	"fansX/pkg/hotkey-go/worker/config"
	"sync"
)

type Window struct {
	config *config.WindowConfig
	mutex  sync.Mutex
	// 上次访问该key的时间戳(millisecond)
	lastTime int64
	// 上次访问该key的时间窗口
	lastIndex int64
	// 上次发送的时间，发送过一次，一段时间内不在处理该key
	lastSend int64
	window   []int64
	// 窗口内访问次数总数
	total int64
}
