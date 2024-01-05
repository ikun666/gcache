package lru

import "container/list"

type LRUCache struct {
	//maxBytes为0表示不限大小
	maxBytes int64
	nbytes   int64
	list     *list.List
	cache    map[string]*list.Element
	// 删除时的回调函数
	OnRemove func(key string, value Value)
}

//链表键值对
type entry struct {
	key   string
	value Value
}

// 获取value值内存大小
type Value interface {
	Len() int
}

func NewLRUCache(maxBytes int64, OnRemove func(key string, value Value)) *LRUCache {
	return &LRUCache{
		maxBytes: maxBytes,
		list:     list.New(),
		cache:    make(map[string]*list.Element),
		OnRemove: OnRemove,
	}
}
func (c *LRUCache) Get(key string) (value Value, ok bool) {
	//如果有将其移到队头  队头元素常用 队尾元素不常用
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)
		c.list.MoveToFront(ele)
		return kv.value, true
	}
	return
}
func (c *LRUCache) RemoveOldest() {
	//删除队尾元素 以其映射 修改容量 执行回调
	ele := c.list.Back()
	if ele != nil {
		kv := ele.Value.(*entry)
		c.list.Remove(ele)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnRemove != nil {
			c.OnRemove(kv.key, kv.value)
		}
	}
}

func (c *LRUCache) Add(key string, value Value) {
	//如同存在，就将其移到队头  更新新值大小
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)
		c.list.MoveToFront(ele)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else { //不存在插入队头
		ele := c.list.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	//插入后 如果超过容量 删除最近最久未使用的元素直到不超过容量
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}
func (c *LRUCache) Len() int {
	return c.list.Len()
}
