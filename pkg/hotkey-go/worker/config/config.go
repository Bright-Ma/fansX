package config

type Config struct {
	Group  GroupConfig
	Window WindowConfig
}

type GroupConfig struct {
	Name string
}

type WindowConfig struct {
	// 100ms
	Size int64
	// cnt 访问次数达到时为热key
	Threshold int64
	// second,发送间隔
	TimeWait int64
	// second，超时时间
	Timeout int64
}
