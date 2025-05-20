# fansX
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/1jian10/fansX)
## 概览
这是一个微服务视频项目，思路源于一些书籍及国内技术团队的文章，目前正在开发
### 中间件/pkg
- [一致性hash](pkg/consistenthash/README.md) 
- [redis分布式锁](pkg/sync/README.md) 
- [leaf-go](pkg/leaf-go/README.md)
- [hotkey-go](pkg/hotkey-go/README.md)
### 系统与服务
- [关系处理](services/relation/README.md)
- [内容](services/content/README.md)
- [feed流](services/feed/README.md)
- [点赞](services/like/README.md)
- [评论](services/comment/README.md)
- [全局消息](services/pusher/README.md)
- [用户](services/user/README.md)
- [认证](services/auth/README.md)

### 接下来的计划
- 实现细节补充  
- 测试，debug与文档编写
- 重写消息系统
### 放弃开发的系统
- IM
- 权限控制
- 搜索服务

## 组件及框架
### 组件
#### 运行必备组件
- [kafka](https://kafka.apache.org/)
- [etcd](https://etcd.io/)
- [redis](https://github.com/redis/redis)
- [TiDB](https://docs.pingcap.com/zh/)
> 最少部署测试集群+ticdc
#### 推荐部署的组件
- [ELK](https://www.elastic.co/)
- [filebeat](https://www.elastic.co/cn/beats/filebeat)
- [jaeger](https://jaeger.golang.ac.cn/)
- [prometheus](https://prometheus.ac.cn/)
- [grafana](https://grafana.com/)
### 框架
进入项目文件夹使用，将会自动下载运行所需框架
```shell
go mod tidy
```

## 推荐阅读
原理：[DDIA](https://book.douban.com/subject/30329536/)  
技术选型/原理/架构：[凤凰架构](https://book.douban.com/subject/35492898/)
