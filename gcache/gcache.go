package gcache

import (
	"fmt"
	"log/slog"
	"sync"
)

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// 接口型函数 函数绑定方法实现接口调用自己
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// 分组cache
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache
// 没有缓存会调用回调函数加载
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		slog.Info("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

// 没有缓存  可选本地和远端加载
func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

// 从本地加载数据
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 加载到缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
