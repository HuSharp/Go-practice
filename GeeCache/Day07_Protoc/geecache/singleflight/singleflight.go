package singleflight

import "sync"

// call 表示正在进行中，或者已经结束的请求
type call struct {
	wg	sync.WaitGroup	// 避免重入
	val	interface{}
	err error
}

// Group 用于管理不同 key 的请求（call
type Group struct {
	mu	sync.Mutex	// g.mu 是保护 Group 的成员变量 m 不被并发读写而加上的锁。
	m	map[string]*call
}

// Do 方法，接收 2 个参数，第一个参数是 key，第二个参数是一个函数 fn。
// Do 的作用就是，针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误。
func (group *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	group.mu.Lock()
	if group.m == nil {
		group.m = make(map[string]*call)
	}
	if c, ok := group.m[key]; ok {
		group.mu.Unlock()
		c.wg.Wait()		// 如果请求正在进行中， 则等待
		return c.val, c.err
	}
	// 若之前没有请求过，那么构建新的
	c := new(call)
	c.wg.Add(1)		// 发请求前加锁
	group.m[key] = c
	group.mu.Unlock()

	c.val, c.err = fn()		// 调用 fn 。发起请求
	c.wg.Done()				// 请求结束  wg.Done() 锁减1。

	group.mu.Lock()
	delete(group.m, key)
	group.mu.Unlock()

	return c.val, c.err
}
