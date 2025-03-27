package snowflake

import (
	leaf "bilibili/common/middleware/leaf-go"
	"bilibili/common/util"
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	etcd "go.etcd.io/etcd/client/v3"
	"strconv"
	"time"
)

func NewCreator(ctx context.Context, config *leaf.SnowflakeConfig) (*Creator, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   config.EtcdAddr,
		DialTimeout: time.Second,
	})
	if err != nil {
		return nil, err
	}

	lock, err := util.EtcdLock(ctx, client, "lock/"+config.CreatorName)
	if err != nil {
		return nil, err
	}
	defer lock.Unlock()

	res, err := client.Get(ctx, "IdCreator/"+config.CreatorName+"/"+config.Addr)
	if err != nil {
		return nil, err
	}

	if len(res.Kvs) == 1 {
		id, err := strconv.Atoi(string(res.Kvs[0].Value))
		if err != nil {
			return nil, err
		}

		return initCreator(ctx, client, config, int64(id))
	}

	res, err = client.Get(ctx, "IdCreator/"+config.CreatorName, etcd.WithPrefix())
	if err != nil {
		return nil, err
	}

	id := int64(len(res.Kvs))
	if id == 1024 {
		return nil, errors.New("worker id not enough")
	}

	_, err = client.Put(ctx, "IdCreator/"+config.CreatorName+"/"+config.Addr, strconv.FormatInt(id, 10))
	if err != nil {
		return nil, err
	}

	return initCreator(ctx, client, config, id)
}

func initCreator(ctx context.Context, client *etcd.Client, config *leaf.SnowflakeConfig, id int64) (*Creator, error) {
	node, err := snowflake.NewNode(id)
	if err != nil {
		return nil, err
	}

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

	go delResp(ch)
	_, err = client.Put(ctx, "IdCreatorTemporary/"+config.CreatorName+"/"+config.Addr, config.Addr, etcd.WithLease(leaseResp.ID))
	if err != nil {
		return nil, err
	}

	c := &Creator{
		name:     config.CreatorName,
		addr:     config.Addr,
		working:  true,
		client:   client,
		snowNode: node,
		lease:    leaseResp.ID,
	}
	go c.heartCheck()
	return c, nil
}

func delResp(ch <-chan *etcd.LeaseKeepAliveResponse) {
	for {
		select {
		case <-ch:
			continue
		}
	}
}
