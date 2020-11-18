package lru

/*
lru
Least Recently Used 最近最少使用
*/

import "container/list"

// Cache is a LRU cache. It is not safe for concurrent access.
// 非并发安全
type Cache struct {
	maxBytes int64                    // 允许使用最大内存
	nbytes   int64                    // 已使用内存大小
	ll       *list.List               // 双向链表，用来保存数据
	cache    map[string]*list.Element // 用来查找数据
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value) // 清除缓存的回调函数
}

// entry 双向链表每个节点缓存的数据类型
// 在链表中仍保存每个值对应的 key 的好处在于，淘汰队首节点时，需要用 key 从字典中删除对应的映射
type entry struct {
	key   string
	value Value // 接口类型
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// New is the Constructor of Cache
// 工厂函数，创建缓存实例
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted, // 清除缓存的回调函数
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 缓存key已存在
		c.ll.MoveToFront(ele)                                  // 将节点移到双链表头部
		kv := ele.Value.(*entry)                               // 获取原节点数据
		c.nbytes += int64(value.Len()) - int64(kv.value.Len()) // 更新已使用缓存的大小（新的大小-旧的大小）
		kv.value = value                                       // 更新节点的缓存数据
	} else {
		// 缓存key不存在
		ele := c.ll.PushFront(&entry{key, value})        // 直接添加数据到双链表头部
		c.cache[key] = ele                               // 更新缓存map
		c.nbytes += int64(len(key)) + int64(value.Len()) // 更新已使用缓存的大小（key+value）
	}

	// 如果已使用缓存大小超过设定的最大值，迭代删除旧数据
	// 如果最大值设为0,不会删除旧数据
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get look ups a key's value
// 获取缓存数据
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)    // 将数据移到双链表头部
		kv := ele.Value.(*entry) // 类型断言，获取数据
		return kv.value, true
	}
	// key不存在时，返回 nil,false
	return
}

// RemoveOldest removes the oldest item
// 删除缓存数据
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 获取双链表排最后的数据
	if ele != nil {
		c.ll.Remove(ele)                                       // 移除最后的数据
		kv := ele.Value.(*entry)                               // 获取移除的数据
		delete(c.cache, kv.key)                                // 维护map
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 维护已使用大小
		if c.OnEvicted != nil {                                // 回调函数
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len the number of cache entries
// 获取缓存实例的双链表的节点数量
func (c *Cache) Len() int {
	return c.ll.Len()
}
