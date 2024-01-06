package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/ikun666/gcache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *gcache.Group {
	return gcache.NewGroup("scores", 2<<10, gcache.GetterFunc(
		func(key string) ([]byte, error) {
			slog.Info("[SlowDB] ", "search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 启动缓存服务器：创建 HTTPPool，添加节点信息，注册到 g 中，启动 HTTP 服务
func startCacheServer(addr string, addrs []string, g *gcache.Group) {
	peers := gcache.NewHTTPPool(addr)
	peers.Add(addrs...)
	g.RegisterPeers(peers)
	slog.Info("gcache is running at", "addr", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// 启动一个 API 服务
func startAPIServer(apiAddr string, g *gcache.Group) {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := g.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())

	})
	slog.Info("server is running at", "apiAddr", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Gcache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	g := createGroup()
	if api {
		go startAPIServer(apiAddr, g)
	}
	startCacheServer(addrMap[port], addrs, g)
}
