# leaf-go：go语言实现的分布式id生成器

目前提供两种id生成策略，数据库segment模式，snowflake模式  
其中segment模式需要使用mysql/tidb，snowflake需要使用etcd  
本文不提供详细的实现细节，请阅读下文实现参考
## 快速开始

### segment

首先在db中创建这样一张表  
```go
type IdTable struct {
	ID       int64     `gorm:"primary_key"`
	Tag      string    `gorm:"unique_index;not null;size:255"`
	MaxId    int64     `gorm:"not null;default:0"`
	Step     int64     `gorm:"not null;default:1024"`
	Desc     string    `gorm:"size:255"`
	UpdateAt time.Time `gorm:"autoUpdateTime"`
}
```
选取segment模式，填写配置，err==nil则可直接使用
```go
creator, err := leaf.NewCore(leaf.Config{
	Model: leaf.Segment,
	SegmentConfig: &leaf.SegmentConfig{
		Name:     "",
		UserName: "",
		Password: "",
		Address:  "",
	},
})
creator.GetId()
```

### snowflake

无需额外配置，选取snowflake模式，填写配置，err==nil即可直接使用
```go
creator, err := leaf.NewCore(leaf.Config{
	Model: leaf.Snowflake,
	SnowflakeConfig: &leaf.SnowflakeConfig{
		CreatorName: "",
		Addr:        "",
		EtcdAddr:    []string{},
	},
})
```

## api
该包只提供一个interface
```go
type Core interface {
	// GetId 获取一个分布式唯一id，若可用则返回id+true，否则返回0+false
	// 虽然只尝试一次，但阻塞时间未必能忽略
	GetId() (int64, bool)
	// GetIdWithContext 内部循环调用GetId，context超时则返回err，请保证传入的ctx带有超时时间
	GetIdWithContext(ctx context.Context) (int64, error)
	// GetIdWithTimeout 内部调用GetIdWithContext
	GetIdWithTimeout(time.Duration) (int64, error)
}
```

## 部分原理

### segment
segment主要采取预分配的方式来获取id，因此，他的id不会造成重复  
但同时，他又强依赖db作为id分发者，只要db宕机，id消耗完毕，就会阻塞直到db恢复  
这种预分配的思想在tidb的自增实现中也有所体现，参考：[AUTO_INCREMENT](https://docs.pingcap.com/zh/tidb/stable/auto-increment/)
#### step
在db中的table中，有一个step字段，默认值为1024，他表示每次向db申请的id的数量，若想提高性能，请将其设置为一个较大的值  
#### 阻塞预防 
一个原始的实现，每次id耗尽在向db中申请新一批的id，在id耗尽时可能造成阻塞  
因此，可以采取预申请的方案，在id耗尽之前向db中申请新一批的id，这样，在耗尽后，直接在内存中进行替换，可有效防止阻塞

### snowflake
snowflake是一个很知名的id生成算法，但在实践中会存在诸多问题，比如：时钟回拨造成的重复id，唯一worker_id如何获取，如何回收worker_id  
这里的实现并不能100%避免时钟回拨，只能尽量避免时钟回拨与时钟回拨造成的影响  

#### worker_id
在snowflake算法中，有10bit保留，由使用者指定的worker_id，在不同服务实例中，该id应为唯一值  
这里采取etcd作为协调中间件，在服务启动时，+etcd分布式锁，保证worker_id的获取不会被其他实例干扰

#### 降低时钟回拨概率 
在节点启动时，通过etcd获取所有相同的服务实例，向他们发送rpc请求，获取其时间戳，对他们求平均值  
若本地时钟较之有较大差距，则认为有较大概率发生时钟回拨，阻止本地节点启动  

#### 时钟回拨应对  
运行时，向etcd上报自身时钟信息，若发生小步长回拨，则阻塞一段时间后恢复服务，若为大步长回拨，则直接panic  
也有着不阻塞的实现方式，在snowflake中分出几bit，初始设置为0，在小步长回拨时，不进行阻塞，而是对分出的bit进行自增(该做法未实现)

#### 弱依赖etcd 
etcd只在启动时获取worker_id起着必要的作用，运行时并非必须，若etcd不可用，则将时钟信息上报切换为本地记录，防止etcd不可用对服务造成的影响  

## 实现参考
[Leaf——美团点评分布式ID生成系统](https://tech.meituan.com/2017/04/21/mt-leaf.html)


