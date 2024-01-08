package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val any
	err error
}

type Group struct {
	mu sync.Mutex // protects m
	m  map[string]*call
}

//包装函数，多次请求执行一次
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	//延迟初始化
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	//如果请求存在，等待请求结果
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	//没有加入map，后续请求等待结果即可
	c := &call{}
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	//执行请求，执行完毕其余等待请求直接返回结果
	c.val, c.err = fn()
	c.wg.Done()
	//请求结束，删除请求
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
