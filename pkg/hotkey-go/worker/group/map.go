package group

import (
	"errors"
	"fansX/pkg/hotkey-go/worker/config"
	"fansX/pkg/hotkey-go/worker/connection"
	cmap "github.com/orcaman/concurrent-map/v2"
	"time"
)

func init() {
	groupMap = &Map{groups: cmap.New[*group]()}
	go groupMap.tick()
	go groupMap.checkWindow()
	go groupMap.checkConnection()
}

// GetGroupMap 饿汉单例模式
func GetGroupMap() *Map {
	return groupMap
}

func (m *Map) Delete(key string) {
	m.groups.Remove(key)
}

func (m *Map) Update(cf config.Config) {
	m.groups.Set(cf.Group.Name, newGroup(cf))
}

func (m *Map) AddKey(groupName string, conn *connection.Conn, keys []string, times []int64) error {
	g, found := m.groups.Get(groupName)
	if !found {
		return errors.New("group not found")
	}
	g.connectionSet.SetIfAbsent(conn, true)
	g.addKey(keys, times)
	return nil
}

// 心跳维持
func (m *Map) tick() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		for _, group := range m.groups.Items() {
			for conn := range group.connectionSet.Items() {
				conn.Ping()
			}
		}
	}
}

// 连接检测
func (m *Map) checkConnection() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		for _, group := range m.groups.Items() {
			for conn := range group.connectionSet.Items() {
				if conn.IsTimeout() {
					conn.Close()
					group.connectionSet.Remove(conn)
				}
			}
		}
	}
}

// 时间窗口检测
func (m *Map) checkWindow() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		for _, group := range m.groups.Items() {
			for key, w := range group.keys.Items() {
				if w.Timeout() {
					group.keys.Remove(key)
				}
			}
		}
	}
}
