package leaf_go

import (
	"context"
	"errors"
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

// Core leaf-go的核心，提供GetId接口，返回true时认为正常工作，false则认为存在错误
type Core interface {
	GetId() (int64, bool)
	// GetIdWithContext 内部循环调用GetId，context超时则返回err，请保证传入的ctx带有超时时间
	GetIdWithContext(ctx context.Context) (int64, error)
	// GetIdWithTimeout 内部调用GetIdWithContext
	GetIdWithTimeout(time.Duration) (int64, error)
}

// Init 创建一个实现GetId的实例，注意传入config时，需要指明开启哪种模式以及对应的config
func Init(c *Config) (Core, error) {
	if c.Model == Segment {
		return segment.NewCreator(&segment.Config{
			Name:     c.SegmentConfig.Name,
			UserName: c.SegmentConfig.UserName,
			Password: c.SegmentConfig.Password,
			Address:  c.SegmentConfig.Address,
		})
	} else if c.Model == Snowflake {
		return snowflake.NewCreator(context.Background(), &snowflake.Config{
			CreatorName: c.SnowflakeConfig.CreatorName,
			Addr:        c.SnowflakeConfig.Addr,
			EtcdAddr:    c.SnowflakeConfig.EtcdAddr,
		})
	}

	return nil, errors.New("please select id model")
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
