/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// 一致性哈希的实现
package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // 哈希算法
	replicas int            // 副本数量，其中副本即为虚拟节点
	keys     []int          // key，即为服务器的得到的hash key值，scice为升序排列。
	hashMap  map[int]string // 哈希表，即缓存服务器到key的一个映射
}

// 本函数用于新生成一个一致性哈希map
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn, // 哈希函数可以自己制定
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE // 如果输入没有哈希函数的话，则可使用默认哈希函数
	}
	return m
}

// 判断是否为空
func (m *Map) IsEmpty() bool {
	return len(m.keys) == 0
}

// 对一致性哈希环中新增节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 对其本身，以及副本使用编号+key的方式计算出其hash值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash) // 将计算出来的hash增加至keys的slice中
			m.hashMap[hash] = key         // 增加hash <-> key的一个映射
		}
	}
	sort.Ints(m.keys) // 重新排列
}

// 从哈希环中找到最合适的一个缓存节点
func (m *Map) Get(key string) string {
	// 首先判断是否为空，如果为空则返回空
	if m.IsEmpty() {
		return ""
	}

	// 使用同一哈希算法计算出key的对应哈希值
	hash := int(m.hash([]byte(key)))

	// 使用二分查找找到最合适的节点
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	// 如果没有找到，即到达了环末尾，则返回首节点即可
	if idx == len(m.keys) {
		idx = 0
	}

	// 返回对应的节点名称
	return m.hashMap[m.keys[idx]]
}
