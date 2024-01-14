package lru

import "container/list"

type MyCache struct {
	maxBytes   int64
	curBytes   int64 //当前缓存的字节
	linkedList *list.List
	cache      map[string]*list.Element
	onEvicted  func(key string, value Value) //驱逐策略
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	//返回所占用内存大小
	Len() int
}

func NewMyCache(maxBytes int64, onEvicted func(string, Value)) *MyCache {
	return &MyCache{
		maxBytes:   maxBytes,
		linkedList: list.New(),
		cache:      make(map[string]*list.Element),
		onEvicted:  onEvicted,
	}
}

func (c *MyCache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.linkedList.MoveToFront(ele)
		kv := ele.Value.(*entry) //断言
		return kv.value, true
	}
	return
}

func (c *MyCache) RemoveOldest() {
	ele := c.linkedList.Back()
	if ele != nil {
		c.linkedList.Remove(ele)
		kv := ele.Value.(*entry)

		//从字典中删除该节点的映射关系
		delete(c.cache, kv.key)
		c.curBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

func (c *MyCache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.linkedList.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.curBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.linkedList.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.curBytes += int64(len(key) + value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.curBytes {
		c.RemoveOldest()
	}
}
