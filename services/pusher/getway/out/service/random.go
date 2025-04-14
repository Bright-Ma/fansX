package service

import (
	etcd "go.etcd.io/etcd/client/v3"
)

func randomWatch(WatchChan etcd.WatchChan) {
	for resp := range WatchChan {
		for _, ev := range resp.Events {
			if ev.Type == etcd.EventTypePut {
				Addr[string(ev.Kv.Value)] = "1"
			} else {
				delete(Addr, string(ev.Kv.Value))
			}
		}
		for k := range Addr {
			NewAddr = append(NewAddr, k)
		}
		mu.Lock()
		//更新地址，因slice为引用，速度较快
		OldAddr = NewAddr
		mu.Unlock()
	}
}

func randomInit(resp *etcd.GetResponse) {
	for _, v := range resp.Kvs {
		OldAddr = append(OldAddr, string(v.Value))
	}
}
