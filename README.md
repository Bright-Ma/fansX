# fansX
## 概览
这是一个微服务视频项目，思路源于一些书籍及国内技术团队的文章，目前正在开发
### 当前已完成部分

- leaf-go 分布式唯一id生成器
- hotkey-go 本地缓存与热点发现
- 关系服务
- 内容发布/管理
- feed流
- 消息推送系统
- 点赞系统
- 评论系统

### 接下来的计划  
- 重构与优化
- IM(x)
- 完善认证，权限控制，凭证系统

## 组件及框架
### 组件
#### 运行必备组件
- [kafka](https://kafka.apache.org/)
- [etcd](https://etcd.io/)
- [redis](https://github.com/redis/redis)
- [minio](https://min.io/)
- [TiDB](https://docs.pingcap.com/zh/)
> 最少部署测试集群+ticdc
- [ELK](https://www.elastic.co/)
>elasticsearch为必须部署，其他选择部署
#### 推荐部署的组件
- [filebeat](https://www.elastic.co/cn/beats/filebeat)
- [jaeger](https://jaeger.golang.ac.cn/)
- [prometheus](https://prometheus.ac.cn/)
- [grafana](https://grafana.com/)
### 框架
进入项目文件夹使用，将会自动下载运行所需框架
```shell
go mod tidy
```