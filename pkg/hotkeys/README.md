# hotkey-go：自动热点发现，轻松解决热key问题
该中间件基于京东hotkey：[京东零售/hotkey](https://gitee.com/jd-platform-opensource/hotkey)  
只按照其架构实现了基础功能，超级青春版(源码9k+行)
## 概览
该包提供分为两个模块，嵌入go程序的接口，计算热key的worker节点  
api封装了go的本地缓存，每次获取本地缓存内容时，视为对key的一次访问，将其发送到worker节点进行计算  
worker节点计算认为该key为热key，将该key发回

## worker
热key计算节点，启动时在etcd中注册实例，go程序获取该实例并与之建立长连接(tcp)  


