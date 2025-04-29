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

func (w *Window) Add(times int64) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	t := time.Now().UnixMilli()
	if t-w.lastSend <= w.config.TimeWait*1000 {
		return false
	}
	if t-w.lastTime > w.config.Size*100 {
		for i := 0; i < len(w.window); i++ {
			w.window[i] = 0
		}
		w.window[0] = times
		w.total = times
		return times >= w.config.Threshold
	}

	for t/100 != w.lastTime/100 {
		w.lastTime += 100
		next := (w.lastIndex + 1) % (int64(len(w.window)))
		w.total -= w.window[next]
		w.window[next] = 0
		w.lastIndex = next
	}
	w.total += times
	w.window[w.lastIndex] += times
	return w.total >= w.config.Threshold
}

func (w *Window) ResetSend() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.lastSend = time.Now().UnixMilli()
	return
}

func (w *Window) Timeout() bool {
	return time.Now().UnixMilli()-w.lastTime >= w.config.Timeout*1000
}
