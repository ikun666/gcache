package gcache_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"github.com/ikun666/gcache"
)

// var db = map[string]string{
// 	"Tom":  "630",
// 	"Jack": "589",
// 	"Sam":  "567",
// }

func TestHttp(t *testing.T) {
	gcache.NewGroup("scores", 2<<10, gcache.GetterFunc(
		func(key string) ([]byte, error) {
			slog.Info("[SlowDB] search key", "key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			slog.Info("[not exist]", "key", key)
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := gcache.NewHTTPPool(addr)
	slog.Info("gcache is running at", "addr", addr)
	err := http.ListenAndServe(addr, peers)
	if err != nil {
		slog.Error("[Server]", "err", err)
	}
}
