package singleflight

import (
	"io"
	"log/slog"
	"net/http"
	"sync"
	"testing"
)

func TestSingleFlight(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := http.Get("http://localhost:9999/api?key=Tom")
			if err != nil {
				slog.Error("[Get]", "err", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				slog.Error("server returned", "res.Status", res.Status)
			}

			bytes, err := io.ReadAll(res.Body)
			if err != nil {
				slog.Error("reading response body", "err", err)
			}
			slog.Info("[data]", "data", string(bytes))
		}()
	}
	wg.Wait()
}
