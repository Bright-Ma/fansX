package cshash

import (
	"hash/fnv"
	"sort"
	"strconv"
)

func Init(nums int) {
	OldMap = &Map{
		nodes: []virtualNode{},
	}
	virtualNums = nums
}

func Update(deleteKeys []string, insertKeys []string) {
	nodes := make([]virtualNode, 4096*16)
	h := fnv.New64a()
	for _, v := range insertKeys {
		for j := 0; j < virtualNums; j++ {
			virtualKey := v + "_" + strconv.FormatInt(int64(j), 10)
			_, _ = h.Write([]byte(virtualKey))
			nodes = append(nodes, virtualNode{
				virtualKey: virtualKey,
				key:        v,
				value:      h.Sum64(),
			})
		}
	}
	DeleteKeys := make(map[string]bool)
	for _, v := range deleteKeys {
		DeleteKeys[v] = true
	}
	for _, v := range OldMap.nodes {
		if !DeleteKeys[v.key] {
			nodes = append(nodes, v)
		}
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].value < nodes[j].value
	})
	NewMap = &Map{
		nodes: nodes,
	}
	rmu.Lock()
	OldMap = NewMap
	rmu.Unlock()
}

func Get(keys []string) []string {
	res := make([]string, len(keys))
	h := fnv.New64a()
	for index, key := range keys {
		_, _ = h.Write([]byte(key))
		rmu.RLock()
		res[index] = OldMap.search(h.Sum64())
		rmu.RUnlock()
	}

	return res
}

func (m *Map) search(hashValue uint64) string {
	left := 0
	right := len(m.nodes) - 1
	for left <= right {
		mid := (left + right) / 2
		if m.nodes[mid].value < hashValue {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return m.next(left)
}

func (m *Map) next(index int) string {
	if index == len(m.nodes) {
		index = 0
	}
	for index < len(m.nodes)-1 && m.nodes[index].value == m.nodes[index+1].value {
		index++
	}
	if index == len(m.nodes)-1 && m.nodes[index-1].value == m.nodes[index].value {
		return m.nodes[0].key
	}
	return m.nodes[0].key
}
