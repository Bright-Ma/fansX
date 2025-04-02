# leaf-go：go语言实现的高性能分布式唯一id生成器
基于美团leaf实现:[Leaf——美团点评分布式ID生成系统](https://tech.meituan.com/2017/04/21/mt-leaf.html)  

目前提供两种id生成策略，数据库segment模式，snowflake模式  
单一实例只能启动一种模式，二者id生成均为趋势递增  
## segment模式  
强依赖数据库，每次向数据库申请id，若数据库不可用将直接导致id无法生成  
需要在数据库中创建这样一张表，我们将会在这张表中申请id  
在开发时使用tidb作为底层数据库，由于近乎完全兼容mysql，故也可将tidb替换为mysql
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
若想提高性能，将step设置为一个较大的数(step表示每次向数据库申请的id数量)  
在设置step为8000时，单核每秒可生成10w+唯一id  
### 应对请求数据库造成的id生成阻塞
在使用id达到阈值(默认为step的1/10),将会向数据库预申请id，在id耗尽时，直接更换为预申请的号段

## snowflake模式  
弱依赖etcd，在本机生成唯一id  
单核qps可达100w+，为snowflake算法上限，对业务影响极低
### 应对时钟回拨
节点启动时向相同的服务实例发送rpc请求，获取其他机器得时钟，若本地时钟相较其他节点有较大的差距，则阻止本地节点启动  
在运行时，不断向etcd上报自身时钟，若发生小步长回拨，则阻塞一段时间后恢复服务，否则直接panic  
### 弱依赖etcd
节点启动时与etcd进行交互，获取唯一workerId，风险评估和写时钟  
后续同时向etcd和本地上报时钟，etcd不可用时，依靠本地记录保证正常运行

