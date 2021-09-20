package lru

import "container/list"

type Cache struct {
	maxBytes	int64		// 允许使用最大内存
	usedBytes	int64		// 已经使用的内存
	ll			*list.List	// 双向链表
	cache		map[string]*list.Element	// 队首为 最新的
	OnEvicted	func(key string, value Value)	//	如果回调函数 OnEvicted 不为 nil，则调用回调函数。
}

type entry struct {
	key		string
	value 	Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len()	int
}

// New 实例化
func New(maxBytes int64, OnEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

// Get ：find
func (c *Cache) Get(key string) (val Value, ok bool) {
	// 如果存在，就取出，并移到队首
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest ：删除, 即移除队尾频率最少元素
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)	// ll 中移除
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)	// map 中删除
		c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())	// 使用内存去除
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)	// 如果回调函数 OnEvicted 不为 nil，则调用回调函数。
		}
	}
}

// Add : add or modify
func (c *Cache) Add(key string, value Value) {
	// 判断是否存在
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)
		// 由于 val 值可能更新，因此需要更新使用内存
		c.usedBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		element := c.ll.PushFront(&entry{
			key:   key,
			value: value,
		})
		c.cache[key] = element
		c.usedBytes += int64(value.Len()) + int64(len(key))
	}
	// 由于更新可能会超过最大内存，因此需要删除，且可能删除一个还不够，所以采用 for
	for c.maxBytes != 0 && c.usedBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

// Len ：为了方便测试，我们实现 Len() 用来获取添加了多少条数据
func (c *Cache) Len() int {
	return c.ll.Len()
}