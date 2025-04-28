package leaf_go

type Config struct {
	Model           int
	SegmentConfig   *SegmentConfig
	SnowflakeConfig *SnowflakeConfig
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

const (
	Segment   = 1
	Snowflake = 2
)
