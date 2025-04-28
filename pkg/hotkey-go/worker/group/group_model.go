package group

import (
	"fansX/pkg/hotkey-go/worker/config"
	cmap "github.com/orcaman/concurrent-map/v2"
	"sync"
)

type Group struct {
	config     config.Config
	idMutex    sync.Mutex
	countMutex []sync.Mutex
	// 对每个连接以一个唯一id进行标识，该id只在单机使用，故没有使用leaf包
	nextId int64
	keys   cmap.ConcurrentMap[string, *count]

	connectionSet cmap.ConcurrentMap[string, *Conn]
}

func (g *Group) Name() string {
	return g.config.GroupName
}

type MapSubject struct {
	concurrentMap cmap.ConcurrentMap[string, Observer]
}

var Map MapSubject

func (m *MapSubject) Register(ob Observer) {
	m.concurrentMap.Set(ob.Name(), ob)
}

func (m *MapSubject) Remove(ob Observer) {
	m.concurrentMap.Remove(ob.Name())
}

func (m *MapSubject) Notify(args ...interface{}) {
	return
}

func (g *Group) Update() {

}

type Observer interface {
	Update()
	Name() string
}

type Subject interface {
	Register(ob Observer)
	Remove(ob Observer)
	Notify(config *config.Config)
}
