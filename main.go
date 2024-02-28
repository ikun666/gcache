package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ikun666/gcache"
	"github.com/ikun666/gcache/conf"
	"github.com/ikun666/gcache/db"
	"github.com/ikun666/gcache/etcd"
	grpcserver "github.com/ikun666/gcache/grpc"
	httpserver "github.com/ikun666/gcache/http"
)

// var db = map[string]string{
// 	"Tom":  "630",
// 	"Jack": "589",
// 	"Sam":  "567",
// 	"IKUN": "250",
// 	"CXK":  "2.5",
// }

//	func createGroup() *gcache.Group {
//		return gcache.NewGroup("scores", 2<<10, gcache.GetterFunc(
//			func(key string) ([]byte, error) {
//				slog.Info("[SlowDB] ", "search key", key)
//				if v, ok := db[key]; ok {
//					return []byte(v), nil
//				}
//				return nil, fmt.Errorf("%s not exist", key)
//			}))
//	}
func createGroup(name string) *gcache.Group {
	return gcache.NewGroup(name, int64(conf.GConfig.MaxBytes), gcache.GetterFunc(
		func(key string) ([]byte, error) {
			// 从后端数据库中查找
			fmt.Println("进入 GetterFunc，数据库中查询....")
			var scores []*db.Student
			db.DB.Where("name = ?", key).Find(&scores)
			if len(scores) == 0 {
				fmt.Println("后端数据库中也查询不到...")
				return []byte{}, errors.New("record not found")
			}

			fmt.Printf("成功从后端数据库中查询到学生 %s 的分数：%s\n", key, scores[0].Score)
			// fmt.Println([]byte(scores[0].Score))
			return []byte(scores[0].Score), nil
		}))
}

// 启动缓存服务器：创建 HTTPPool，添加节点信息，注册到 g 中，启动 HTTP 服务

func startCacheHTTPServer(addr, ip, port, protocol string, g *gcache.Group) {
	server := httpserver.NewHTTPPool(addr, ip, port, protocol)
	addrs, err := etcd.DiscoverPeers(conf.GConfig.Prefix)
	if err != nil {
		log.Println(err)
		return
	}
	// 将节点打到哈希环上
	server.AddPeers(addrs...)
	// 为 Group 注册服务 Picker
	g.RegisterServer(server)
	slog.Info("gcache is running at", "addr", addr)
	// 启动服务
	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}

// 启动缓存服务器：创建 GRPC，添加节点信息，注册到 g 中，启动 GRPC 服务
func startCacheGRPCServer(addr, ip, port, protocol string, g *gcache.Group) {
	server, err := grpcserver.NewServer(addr, ip, port, protocol)
	if err != nil {
		log.Println(err)
		return
	}
	addrs, err := etcd.DiscoverPeers(conf.GConfig.Prefix)
	if err != nil {
		log.Println(err)
		return
	}
	// 将节点打到哈希环上
	server.AddPeers(addrs...)
	// 为 Group 注册服务 Picker
	g.RegisterServer(server)
	log.Println("groupcache is running at ", fmt.Sprintf("%v:%v", ip, port))

	// 启动服务
	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}

// 启动一个 API 服务
func startAPIServer(apiAddr string, g *gcache.Group) {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		key := r.URL.Query().Get("key")
		view, err := g.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// fmt.Println("api", view.String(), view.ByteSlice())
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
		fmt.Printf("use time:%v um\n", time.Since(t1).Milliseconds())

	})
	slog.Info("server is running at", "apiAddr", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	conf.Init("./gcache/conf/conf.json")
	db.Init()
	var port int
	var api bool
	var grpc bool
	flag.IntVar(&port, "port", 8001, "Gcache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.BoolVar(&grpc, "grpc", true, "use grpc/http")
	flag.Parse()

	// apiAddr := "http://localhost:9999"

	g := createGroup("ikun666")
	if api {
		go startAPIServer(conf.GConfig.ApiAddr, g)
	}
	if grpc {
		startCacheGRPCServer("localhost:"+strconv.Itoa(port), "localhost", strconv.Itoa(port), "GRPC", g)
	} else {
		startCacheHTTPServer("localhost:"+strconv.Itoa(port), "localhost", strconv.Itoa(port), "HTTP", g)
	}

}
