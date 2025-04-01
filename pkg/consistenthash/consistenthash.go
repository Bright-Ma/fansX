package cshash

import (
	"hash/fnv"
	"sort"
	"strconv"
)

func NewMap(nums int) *HashMap {
	return &HashMap{
		old:         []virtualNode{},
		new:         []virtualNode{},
		virtualNums: nums,
	}
}

func (hm *HashMap) Update(deleteKeys []string, insertKeys []string) {
	nodes := make([]virtualNode, 0)
	h := fnv.New64a()
	for _, v := range insertKeys {
		for j := 0; j < hm.virtualNums; j++ {
			virtualKey := v + "_" + strconv.FormatInt(int64(j), 10)
			_, _ = h.Write([]byte(virtualKey))
			nodes = append(nodes, virtualNode{
				virtualKey: virtualKey,
				key:        v,
				value:      h.Sum64(),
			})
			h.Reset()
		}
	}

	DeleteKeys := make(map[string]bool)
	for _, v := range deleteKeys {
		DeleteKeys[v] = true
	}

	for _, v := range hm.old {
		if !DeleteKeys[v.key] {
			nodes = append(nodes, v)
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].value < nodes[j].value
	})

	hm.new = nodes
	hm.rmu.Lock()
	hm.old = hm.new
	hm.rmu.Unlock()
}

func (hm *HashMap) Get(keys []string) []string {
	res := make([]string, len(keys))
	h := fnv.New64a()

	hm.rmu.RLock()
	for index, key := range keys {
		_, _ = h.Write([]byte(key))
		res[index] = hm.search(h.Sum64())
		h.Reset()
	}
	hm.rmu.RUnlock()

	return res
}

func (hm *HashMap) search(hashValue uint64) string {
	left := 0
	right := len(hm.old) - 1
	for left <= right {
		mid := (left + right) / 2
		if hm.old[mid].value < hashValue {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return hm.next(left)
}

func (hm *HashMap) next(index int) string {
	if index == len(hm.old) {
		index = 0
	}
	for index < len(hm.old)-1 && hm.old[index].value == hm.old[index+1].value {
		index++
	}
	if index == len(hm.old)-1 && hm.old[index-1].value == hm.old[index].value {
		return hm.old[0].key
	}
	return hm.old[0].key
}
