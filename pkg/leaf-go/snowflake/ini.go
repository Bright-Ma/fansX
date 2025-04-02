package snowflake

import (
	"bilibili/common/util"
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	etcd "go.etcd.io/etcd/client/v3"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"
)

func NewCreator(ctx context.Context, config *Config) (*Creator, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   config.EtcdAddr,
		DialTimeout: time.Second,
	})
	if err != nil {
		return nil, err
	}

	// 申请etcd分布式锁，因为我们需要获取唯一的workerId
	lock, err := util.EtcdLock(ctx, client, "lock/"+config.CreatorName)
	if err != nil {
		return nil, err
	}
	defer lock.Unlock()

	res, err := client.Get(ctx, "IdCreator/"+config.CreatorName+"/"+config.Addr)
	if err != nil {
		return nil, err
	}

	// 该服务之前申请过workerId，复用之前创建的workerId
	if len(res.Kvs) == 1 {
		id, err := strconv.Atoi(string(res.Kvs[0].Value))
		if err != nil {
			return nil, err
		}

		return initCreator(ctx, client, config, int64(id))
	}
	// 获取同一服务的个数，本地节点使用该个数作为workerId
	res, err = client.Get(ctx, "IdCreator/"+config.CreatorName, etcd.WithPrefix())
	if err != nil {
		return nil, err
	}

	id := int64(len(res.Kvs))
	// 达到雪花算法workerId上限
	if id == 1024 {
		return nil, errors.New("worker id not enough")
	}

	_, err = client.Put(ctx, "IdCreator/"+config.CreatorName+"/"+config.Addr, strconv.FormatInt(id, 10))
	if err != nil {
		return nil, err
	}

	return initCreator(ctx, client, config, id)
}

func initCreator(ctx context.Context, client *etcd.Client, config *Config, id int64) (*Creator, error) {
	node, err := snowflake.NewNode(id)
	if err != nil {
		return nil, err
	}

	// 该key永久存储在etcd中，定时上报时钟到该节点
	key := "IdCreatorForever/" + config.CreatorName + "/" + config.Addr
	res, err := client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(res.Kvs) == 0 {
		if _, err = client.Put(ctx, key, strconv.FormatInt(time.Now().UnixMilli(), 10)); err != nil {
			return nil, err
		}
	} else {
		num, err := strconv.ParseInt(string(res.Kvs[0].Value), 10, 64)
		if err != nil {
			return nil, err
		}
		// 时钟回拨
		if time.Now().UnixMilli()-num < 0 {
			return nil, errors.New("clock failed")
		}
	}

	lease := etcd.NewLease(client)
	leaseResp, err := lease.Grant(ctx, 10)
	if err != nil {
		return nil, err
	}

	ch, err := client.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return nil, err
	}

	// etcd中存储正在运行的服务实例地址，用于启动时判断时钟回拨的风险
	// 使用lease将节点和该key绑定到一起，保证节点活跃时自动续约，下线时自动删除
	_, err = client.Put(ctx, "IdCreatorTemporary/"+config.CreatorName+"/"+config.Addr, config.Addr, etcd.WithLease(leaseResp.ID))
	if err != nil {
		return nil, err
	}

	c := &Creator{
		name:     config.CreatorName,
		addr:     config.Addr,
		working:  atomic.Bool{},
		client:   client,
		snowNode: node,
		lease:    leaseResp.ID,
		local:    atomic.Bool{},
	}

	c.local.Store(false)
	c.working.Store(true)

	go c.heartCheck()
	go c.delResp(ch)

	return c, nil
}

func (c *Creator) delResp(ch <-chan *etcd.LeaseKeepAliveResponse) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprint(err))
			slog.Error("change to use local time record")

			c.working.Store(false)
			c.local.Store(true)

			return
		}
	}()
	for range ch {
		continue
	}
	//续约失效，认为etcd不可用
	panic("lease time out")
}
