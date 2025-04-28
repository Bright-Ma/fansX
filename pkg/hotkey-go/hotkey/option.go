package hotkey

import (
	cshash "fansX/pkg/consistenthash"
	"github.com/coocood/freecache"
	"time"
)

// WithCacheSize 设置本地缓存的大小(byte),默认值为4g
func WithCacheSize(size int) Option {
	return OptionFunc(func(core *Core) {
		core.cache = freecache.NewCache(size)
	})
}

// WithKeySize 设置热key缓存的大小(byte)，默认值为128m
func WithKeySize(size int) Option {
	return OptionFunc(func(core *Core) {
		core.hotkeys = freecache.NewCache(size)
	})
}

// WithVirtualNums 设置一致性hash虚拟节点的数量，默认值为50
func WithVirtualNums(nums int) Option {
	return OptionFunc(func(core *Core) {
		core.hashMap = cshash.NewMap(nums)
	})
}

// WithChannelSize 设置kv发送channel大小，默认值为1024*512
func WithChannelSize(size int) Option {
	return OptionFunc(func(core *Core) {
		core.send = make(chan kv, size)
	})
}

// WithSendInterval 设置kv发送间隔，默认值为100ms
func WithSendInterval(interval time.Duration) Option {
	if interval <= 0 {
		panic("invalid interval value")
	}
	return OptionFunc(func(core *Core) {
		core.interval = interval
	})
}

// WithObserver 将观察者加入观察者列表，在将热key加入缓存前通知观察者
func WithObserver(subjects ...Subject) Option {
	return OptionFunc(func(core *Core) {
		for _, sub := range subjects {
			core.register(sub)
		}
	})
}

// WithTTL 设置热key缓存时间(second)，默认值为30s
func WithTTL(ttl int) Option {
	return OptionFunc(func(core *Core) {
		core.ttl = ttl
	})
}
