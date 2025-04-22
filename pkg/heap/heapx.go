package heap

import (
	"container/heap"
)

type LessFunc[T any] func(a, b T) bool

type GenericHeap[T any] struct {
	data []T
	less LessFunc[T]
}

// NewHeap a<b 小根堆
func NewHeap[T any](less LessFunc[T]) *GenericHeap[T] {
	h := &GenericHeap[T]{
		data: make([]T, 0),
		less: less,
	}
	heap.Init(h)
	return h
}

func (h *GenericHeap[T]) Len() int {
	return len(h.data)
}

func (h *GenericHeap[T]) Less(i, j int) bool {
	return h.less(h.data[i], h.data[j])
}

func (h *GenericHeap[T]) Swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

func (h *GenericHeap[T]) Push(x any) {
	h.data = append(h.data, x.(T))
}

func (h *GenericHeap[T]) Pop() any {
	old := h.data
	n := len(old)
	item := old[n-1]
	h.data = old[0 : n-1]
	return item
}

func (h *GenericHeap[T]) PushItem(item T) {
	heap.Push(h, item)
}

func (h *GenericHeap[T]) PopItem() T {
	return heap.Pop(h).(T)
}

func (h *GenericHeap[T]) Peek() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}
	return h.data[0], true
}
