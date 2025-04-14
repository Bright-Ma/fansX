# hotkey-go：自动热点发现，轻松解决热key问题
该中间件基于京东hotkey：[京东零售/hotkey](https://gitee.com/jd-platform-opensource/hotkey)  
只按照其架构实现了基础功能，超级青春版(源码9k+行)
## 概览
该包提供分为两个模块，嵌入go程序的接口，计算热key的worker节点  
api封装了go的本地缓存，每次获取本地缓存内容时，视为对key的一次访问，将其发送到worker节点进行计算  
worker节点计算认为该key为热key，将该key发回

## worker
热key计算节点，启动时在etcd中注册实例，go程序获取该实例并与之建立长连接(tcp)  
每个go程序在连接后会发送一个报文，表示该程序属于哪个group

### 热key计算
热key的计算采用滑动窗口算法，为每一个key维护一个时间窗口，当该窗口内访问次数超过一定数量，则视为热key  
当认定该key为热key，将该key发送到该key所属的group的所有go程序  
判断为热key的一定时间内，将不在发送热key到go程序，但仍会维护时间窗口

## go接口
该接口本质上是对本地缓存的封装，不过在对本地缓存进行操作时会上报key，同时提供热key判断功能(本质上是在本地缓存缓存了一份热key)  
未来将会添加获取热key的channel，使热key的处理更为迅速和灵活

### 性能优化
在对key进行访问时，不会立刻发送到worker节点，而是本地缓存一段时间，将请求进行聚合后发送  
在选择发送key到的worker节点时，选取一致性hash算法，使负载更均衡，同时使worker节点上下线时不会发生大量key的迁移

