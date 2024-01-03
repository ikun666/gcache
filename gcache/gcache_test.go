package gcache_test

import (
	"fmt"
	"testing"

	"log/slog"

	"github.com/ikun666/gcache"
)

// map 模拟耗时的数据库
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	g := gcache.NewGroup("scores", 2<<10, gcache.GetterFunc(
		func(key string) ([]byte, error) {
			slog.Info("[SlowDB] ", "search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range db {
		if view, err := g.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		} // load from callback function
		if _, err := g.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := g.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
	if _, err := g.Get("ikun"); err != nil {
		slog.Error("[Err]", "err", err)
	}
}
