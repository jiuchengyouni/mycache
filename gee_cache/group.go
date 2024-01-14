package gee_cache

import (
	"errors"
	"github.com/sirupsen/logrus"
	"mycache/cache"
	"mycache/gee_cachepb/pb"
	"mycache/singleflight"
	"sync"
)

// 缓存命名空间
type Group struct {
	name      string
	getter    Getter
	mainCache cache.Cache
	peers     PeerPicker

	loader *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache.Cache{CacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (cache.ByteView, error) {
	if key == "" {
		return cache.ByteView{}, errors.New("key is required")
	}
	if v, ok := g.mainCache.Get(key); ok {
		logrus.Info("缓存击中")
		return v, nil
	}
	return g.load(key)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value cache.ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				logrus.Info("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(cache.ByteView), err
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (cache.ByteView, error) {
	req := &pb.Request{Group: g.name, Key: key}
	res := &pb.Response{}
	bytes, err := peer.Get(req, res)
	if err != nil {
		return cache.ByteView{}, err
	}
	return cache.ByteView{B: bytes}, nil
}

func (g *Group) getLocally(key string) (cache.ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return cache.ByteView{}, err
	}
	value := cache.ByteView{B: cache.CloneByte(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value cache.ByteView) {
	g.mainCache.Add(key, value)
}
