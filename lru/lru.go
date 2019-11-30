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

// LRU算法
package lru

import "container/list"

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	// 最大入口数，即最大缓存数量。
	// 如果为0的话，则表示不做限制
	MaxEntries int

	// 销毁前回调函数
	OnEvicted func(key Key, value interface{})

	// 链表
	ll *list.List

	// 缓存，指向列表中元素
	cache map[interface{}]*list.Element
}

// 缓存值，可为任意类型
type Key interface{}

// key-value的访问入口
type entry struct {
	key   Key
	value interface{}
}

// 本函数用于初始化一个Cache
func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),                          // 新建一个链表
		cache:      make(map[interface{}]*list.Element), // 新建映射
	}
}

// 向缓存中新增某一元素
func (c *Cache) Add(key Key, value interface{}) {
	// 如果缓存为空的话，则重新申请map以及list
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	// 判断key是否已经存在于cache中，如果是，则将其移动至列表头部
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value // 设置key的value
		return
	}
	// 如果key在list中不存在的话，则新建一个entry，并且将其插入链表头部
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele // 并且在cache中存放

	// 如果缓存此时已经满了，则需要对其进行淘汰
	// 使用RemoveOldest函数淘汰掉最久没有被使用的缓存
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

// 从缓存中获取一个value
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	// 为空测返回空
	if c.cache == nil {
		return
	}
	// 是否hit，如果hit则需要
	// 1. 将该缓存移动至list的头部
	// 2. 返回该key对应的value
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// 本函数用于删除key对应的cache
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}
	// 使用removeElement函数删除cache中对应项
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// 本函数用于淘汰最久未被使用的缓存
// 1. 从list中获取最久未被使用的对应cache
// 2. 利用removeElement函数淘汰对应cache
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

// 本函数用于淘汰缓存
// 即需要删除链表中对应节点，以及cache中对应项
func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// 本函数用于返回list长度
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// 清除所有缓存
func (c *Cache) Clear() {
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}
