package gcache

import (
	"sync"

	"github.com/ikun666/gcache/lru"
)

type cache struct {
	mu         sync.Mutex //lru 读写都会修改内容
	lru        *lru.LRUCache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//延迟初始化
	if c.lru == nil {
		c.lru = lru.NewLRUCache(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
