package geecache

import (
	"fmt"
	"geecache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)


// 创建具体的 HTTP 客户端类 httpGetter
type httpGetter struct {
	baseURL string	// baseURL 表示将要访问的远程节点的地址，例如 http://example.com/_geecache/。
}

// HTTPPool implements PeerPicker for a pool of HTTP peers.
// HTTPPool，作为承载节点间 HTTP 通信的核心数据结构
type HTTPPool struct {
	self		string	// 用来记录自己的地址，包括主机名/IP 和端口。
	basePath	string	// 作为节点的通讯地址前缀，方便节点访问
	mu			sync.Mutex
	peers		*consistenthash.Map	// 一致性哈希的 map，通过 key 来选择节点
	// 每一个远程节点对应一个 httpGetter，因为 httpGetter 与远程节点的地址 baseURL 有关。
	httpGetters	map[string]*httpGetter	// 映射远程节点与对应的 httpGetter
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		// QueryEscape 函数对s进行转码使之可以安全的用在URL查询里。
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u) // 获取返回值
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

// Set 实例化一致性哈希， 添加传入节点
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 包装了一致性哈希算法的 Get() 方法，根据具体的 key，选择节点，返回节点对应的 HTTP 客户端。
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("[PickPeer] Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

const (
	defaultBasePath = "/geeCache/"
	defaultReplicas = 50
)

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:    	self,
		basePath: 	defaultBasePath,
	}
}

// Log server name info
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http request
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)// 返回前两个
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:" + groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}


