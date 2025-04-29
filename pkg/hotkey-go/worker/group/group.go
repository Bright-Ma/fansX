package group

import (
	"encoding/json"
	"fansX/pkg/hotkey-go/model"
	"fansX/pkg/hotkey-go/worker/config"
	"fansX/pkg/hotkey-go/worker/connection"
	"fansX/pkg/hotkey-go/worker/window"
	cmap "github.com/orcaman/concurrent-map/v2"
)

func newGroup(cf config.Config) *group {
	g := &group{
		config:        &cf,
		keys:          cmap.New[*window.Window](),
		connectionSet: cmap.NewStringer[*connection.Conn, bool](),
	}
	return g
}

// Send 广播，对该group中的所有连接发送消息
func (g *group) send(m string, key []string) {
	msg := &model.ServerMessage{
		Type: m,
		Keys: key,
	}

	s, err := json.Marshal(msg)
	if err != nil {
		return
	}

	mp := g.connectionSet.Items()
	for conn := range mp {
		conn.Send(s)
	}
}

func (g *group) addKey(keys []string, times []int64) {
	for i, v := range keys {
		w := g.keys.Upsert(v, nil, func(exist bool, in *window.Window, new *window.Window) *window.Window {
			if exist {
				return in
			} else {
				return window.NewWindow(&g.config.Window)
			}
		})
		ok := w.Add(times[i])
		if ok {
			w.ResetSend()
			g.send(model.AddKey, []string{v})
		}
	}
	return
}
