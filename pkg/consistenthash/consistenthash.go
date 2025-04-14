package cshash

import (
	"hash/fnv"
	"sort"
	"strconv"
)

// NewMap 新建一致性hash表，nums表示每个真实节点的虚拟节点个数
func NewMap(nums int) *HashMap {
	return &HashMap{
		old:         []virtualNode{},
		new:         []virtualNode{},
		virtualNums: nums,
	}
}

// Update 更新一致性hash表，并发安全，内部使用一把锁保证原子性
func (hm *HashMap) Update(deleteKeys []string, insertKeys []string) {

	// 新虚拟节点的列表
	nodes := make([]virtualNode, 0)
	h := fnv.New64a()
	// 遍历插入key列表，并转换为虚拟节点
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

	// 遍历删除节点并标记，因为虚拟节点保存了真实节点key的信息，所以无需转换为虚拟节点
	DeleteKeys := make(map[string]bool)
	for _, v := range deleteKeys {
		DeleteKeys[v] = true
	}

	hm.rmu.Lock()
	// 将未删除(标记)的节点添加在新虚拟节点列表中
	for _, v := range hm.old {
		if !DeleteKeys[v.key] {
			nodes = append(nodes, v)
		}
	}
	// 对所有虚拟节点进行排序，便于二分查找
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].value < nodes[j].value
	})

	hm.new = nodes
	hm.old = hm.new

	hm.rmu.Unlock()
}

// Get 批量获取key所对应的真实节点的地址
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

// search 找到key的hash值在一致性hash环上下一个虚拟节点所对应的真实节点的key
func (hm *HashMap) search(hashValue uint64) string {
	left := 0
	right := len(hm.old) - 1
	// 二分查找
	for left <= right {
		mid := (left + right) / 2
		if hm.old[mid].value < hashValue {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	//left为首个大于等于hashValue的下标(若为则为len(old))
	return hm.next(left)
}

func (hm *HashMap) next(index int) string {
	// 无更大的
	if index == len(hm.old) {
		index = 0
	}
	//对于有重复hash值的虚拟节点，直接跳过
	for index < len(hm.old)-1 && hm.old[index].value == hm.old[index+1].value {
		index++
	}
	//若在hash环尾且上个节点和当前节点hash值一样，则返回hash环首部的节点
	if index == len(hm.old)-1 && hm.old[index-1].value == hm.old[index].value {
		return hm.old[0].key
	}
	return hm.old[index].key
}
