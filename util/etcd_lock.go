package util

import (
	"context"
	"errors"
	etcd "go.etcd.io/etcd/clientv3"
	"time"
)

type Lock struct {
	client  *etcd.Client
	leaseId etcd.LeaseID
	closed  chan int
}

func EtcdLock(ctx context.Context, client *etcd.Client, key string) (*Lock, error) {
	lease := etcd.NewLease(client)
	leaseResp, err := lease.Grant(ctx, 10)
	if err != nil {
		return nil, err
	}

	lock := &Lock{
		client:  client,
		leaseId: leaseResp.ID,
		closed:  make(chan int),
	}

	ch, err := client.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		lock.Unlock()
		return nil, err
	}

	go func() {
		for {
			select {
			case <-ch:
				continue
			case <-lock.closed:
				close(lock.closed)
				return
			}
		}
	}()

	ok := false
	for i := 0; i < 50; i++ {
		res, err := client.Txn(ctx).
			If(etcd.Compare(etcd.CreateRevision(key), "=", 0)).
			Then(etcd.OpPut(key, "locked", etcd.WithLease(leaseResp.ID))).
			Commit()
		if err != nil || !res.Succeeded {
			time.Sleep(time.Millisecond * 15)
		} else {
			ok = true
			break
		}
	}
	if !ok {
		lock.Unlock()
		return nil, errors.New("lock timeout")
	}

	return lock, nil

}

func (l *Lock) Unlock() {
	l.closed <- 1
	_, _ = l.client.Revoke(context.Background(), l.leaseId)
}
