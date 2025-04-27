package leaf_go

import (
	"context"
	"fansX/pkg/leaf-go/segment"
	"fansX/pkg/leaf-go/snowflake"
	"time"
)

type Config struct {
	Model           int
	SegmentConfig   *SegmentConfig
	SnowflakeConfig *SnowflakeConfig
}

var (
	Segment   = 1
	Snowflake = 2
)

type Core interface {
	// GetId 获取一个分布式唯一id，若可用则返回id+true，否则返回0+false
	GetId() (int64, bool)
	// GetIdWithContext 内部循环调用GetId，context超时则返回err，请保证传入的ctx带有超时时间
	GetIdWithContext(ctx context.Context) (int64, error)
	// GetIdWithTimeout 内部调用GetIdWithContext
	GetIdWithTimeout(time.Duration) (int64, error)
}

func NewSegment(c *SegmentConfig) (Core, error) {
	return segment.NewCreator(&segment.Config{
		Name:     c.Name,
		UserName: c.UserName,
		Password: c.Password,
		Address:  c.Address,
	})
}

func NewSnowflake(c *SnowflakeConfig) (Core, error) {
	return snowflake.NewCreator(context.Background(), &snowflake.Config{
		CreatorName: c.CreatorName,
		Addr:        c.Addr,
		EtcdAddr:    c.EtcdAddr,
	})
}

type SnowflakeConfig struct {
	// 使用的服务名称，同一服务保证不分发相同id，同一服务上限1024个节点
	CreatorName string
	// 该服务的ip+port，其他同一服务启动时获取该机器的时钟，验证时钟回拨的风险
	Addr string
	// etcd地址
	EtcdAddr []string
}

type SegmentConfig struct {
	// 服务名称，同一服务共享数据库的同一记录
	Name string
	// 数据库用户名
	UserName string
	// 数据库密码
	Password string
	// 数据库地址
	Address string
}
