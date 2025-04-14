package service

import (
	"OutGetWay/consistenthash"
	etcd "go.etcd.io/etcd/client/v3"
)

func consistentHashWatch(WatchChan etcd.WatchChan) {
	for resp := range WatchChan {
		del := make([]string, 0)
		ins := make([]string, 0)
		for _, ev := range resp.Events {
			if ev.Type == etcd.EventTypePut {
				ins = append(ins, string(ev.Kv.Value))
			} else {
				del = append(del, string(ev.Kv.Value))
			}
		}
		cshash.Update(del, ins)
	}
}

func consistentHashInit(resp *etcd.GetResponse) {
	for _, v := range resp.Kvs {
		OldAddr = append(OldAddr, string(v.Value))
	}
}
