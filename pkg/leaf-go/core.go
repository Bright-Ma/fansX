package leaf_go

import (
	"context"
	"time"
)

type Core interface {
	// GetId 获取一个分布式唯一id，若可用则返回id+true，否则返回0+false
	// 虽然只尝试一次，但阻塞时间未必能忽略
	GetId() (int64, bool)
	// GetIdWithContext 内部循环调用GetId，context超时则返回err，请保证传入的ctx带有超时时间
	GetIdWithContext(ctx context.Context) (int64, error)
	// GetIdWithTimeout 内部调用GetIdWithContext
	GetIdWithTimeout(time.Duration) (int64, error)
}
