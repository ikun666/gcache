package gcache

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/ikun666/gcache/consistentHash"
)

const (
	defaultBasePath = "/_gcache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	addr     string
	basePath string
	mu       sync.Mutex // guards peers and httpGetters
	peers    *consistentHash.Map
	//每一个远程节点对应一个 httpGetter
	httpGetters map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(addr string) *HTTPPool {
	return &HTTPPool{
		addr:        addr,
		basePath:    defaultBasePath,
		peers:       consistentHash.New(defaultReplicas, nil),
		httpGetters: make(map[string]*httpGetter),
	}
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Info("[Server]", "addr", p.addr, "method", r.Method, "url", r.URL.Path, "basePath", p.basePath)
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		// panic("HTTPPool serving unexpected path: " + r.URL.Path)
		slog.Error("HTTPPool serving unexpected path", "url", r.URL.Path)
		return
	}

	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.b)
}

// Set updates the pool's list of peers.
// 加入节点
func (p *HTTPPool) Add(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers.Add(peers...)
	for _, peer := range peers {
		//"http://10.0.0.2:8008/_gcache/"
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	peer := p.peers.Get(key)
	slog.Info("[PickPeer]", "peer", peer, "p.addr", p.addr, "p.httpGetters[peer]", p.httpGetters[peer])
	//选择的节点不能是空和自身 选自己机会一直调用自己
	if peer != "" && peer != p.addr {
		return p.httpGetters[peer], true
	}
	return nil, false
}

// HTTP 客户端类
type httpGetter struct {
	baseURL string
}

// HTTP 客户端类 httpGetter，实现 PeerGetter 接口。
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}
