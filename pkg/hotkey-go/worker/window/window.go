package window

import (
	"fansX/pkg/hotkey-go/worker/config"
	"time"
)

func NewWindow(cf *config.WindowConfig) *Window {
	w := &Window{
		config:   cf,
		lastTime: time.Now().UnixMilli(),
		window:   make([]int64, cf.Size),
	}

	return w
}

// Add 窗口计算,返回是为热key
func (w *Window) Add(times int64) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	t := time.Now().UnixMilli()
	//距离上次发送时间过短
	if t-w.lastSend <= w.config.TimeWait*1000 {
		return false
	}
	// 访问间隔超过时间窗口长度，重置窗口
	if t-w.lastTime > w.config.Size*100 {
		for i := 0; i < len(w.window); i++ {
			w.window[i] = 0
		}
		w.window[0] = times
		w.total = times
		return times >= w.config.Threshold
	}
	// 擦除时间窗口
	for t/100 != w.lastTime/100 {
		w.lastTime += 100
		next := (w.lastIndex + 1) % (int64(len(w.window)))
		w.total -= w.window[next]
		w.window[next] = 0
		w.lastIndex = next
	}
	// 添加至窗口
	w.total += times
	w.window[w.lastIndex] += times
	return w.total >= w.config.Threshold
}

// ResetSend 重设发送时间
func (w *Window) ResetSend() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.lastSend = time.Now().UnixMilli()
	return
}

// Timeout 返回当前时间窗口是否过长时间没有访问，即认为可以删除该窗口
func (w *Window) Timeout() bool {
	return time.Now().UnixMilli()-w.lastTime >= w.config.Timeout*1000
}
