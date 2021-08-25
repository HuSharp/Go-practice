package geecache

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)	// 回调函数
}

/*GetterFunc 是 一个接口型函数
	接口型函数只能应用于接口内部只定义了一个方法的情况
	这种类型实现了 Get 方法（在 Get 方法中又调用了自身）
	也就是说这个类型的函数其实就是一个 Getter 类型的对象。
	利用这种类型转换，我们可以将此类型的函数转换为一个 Getter 对象，
	而不需要定义一个结构体，再让这个结构实现 Get 方法。
*/
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}


// Group 一个 Group 可以认为是一个缓存的命名空间
type Group struct {
	name		string
	getter		Getter		// 缓存未命中时获取源数据的回调(callback)。
	mainCache	cache
	peers		PeerPicker	// 用于 group 中选取 peer
}

var (
	mu	sync.RWMutex
	groups	= make(map[string]*Group)
)

// NewGroup create a new instance
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
	mu.RLock()	// 只读锁 因为不涉及任何冲突变量的写操作
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get val for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	log.Println("[Get] start Cache Hit!")
	if key == "" {
		return ByteView{}, fmt.Errorf("key is empty")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Get] Cache Hit!")
		return v, nil
	}
	// 如果缓存没有，就进行加载
	return g.load(key)
}

// 加载先看有没有节点，没有就在本地加载，否则去节点中调用 getFromPeer 函数
func (g *Group) load(key string) (val ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if val, err := g.getFromPeer(peer, key); err == nil {
				return val, nil
			}
			log.Println("[load] getFromPeer failed! err: ", err)
		}
	}
	return g.getLocally(key)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		log.Println("[getFromPeer] peer get error: ", err)
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	val := ByteView{b: cloneBytes(bytes)}
	g.addToCache(key, val)
	return val, nil
}

func (g *Group) addToCache(key string, val ByteView) {
	g.mainCache.add(key, val)
}

func (g *Group) RegisterPeers(picker PeerPicker) {
	if g.peers != nil {
		panic("[RegisterPeers] func called more than once!")
	}
	g.peers = picker
}