package gcache

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/ikun666/gcache/singleflight"
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
	//分布式节点
	peers PeerPicker
	//并发请求同一个key只执行一次
	loader *singleflight.Group
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
		loader:    &singleflight.Group{},
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

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		slog.Error("[RegisterPeerPicker called more than once]")
		return
	}
	g.peers = peers
}

// Get value for a key from cache
// 没有缓存会调用回调函数加载
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		slog.Info("[GCache] hit")
		return v, nil
	}

	return g.load(key)
}

// 没有缓存  可选本地和远端加载
func (g *Group) load(key string) (ByteView, error) {
	value, err := g.loader.Do(key, func() (any, error) {
		//优先从远端加载缓存
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				value, err := g.getFromPeer(peer, key)
				if err == nil {
					return value, nil
				}
				slog.Info("[GCache] Failed to get from peer", "err", err)
			}
		}
		return g.getLocally(key)
	})
	if err != nil {
		return ByteView{}, err
	}
	return value.(ByteView), err
}

// 从远端加载数据
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
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
