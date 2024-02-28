package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ikun666/gcache"
	"github.com/ikun666/gcache/conf"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// 从etcd中获取配置项（服务注册发现）
func DiscoverPeers(prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.GConfig.Endpoints,
		DialTimeout: time.Duration(conf.GConfig.DialTimeout) * time.Second,
	})
	if err != nil {
		fmt.Println("create etcd client failed,err:", err)
		return []string{}, err
	}

	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Println("get peer addr list from etcd failed,err:", err)
		return []string{}, err
	}

	var peers []string
	for _, kv := range resp.Kvs {
		service := &Service{}
		err := json.Unmarshal(kv.Value, service)
		if err != nil {
			return []string{}, err
		}
		peers = append(peers, fmt.Sprintf("%v:%v", service.IP, service.Port))
	}
	fmt.Println("get peer addr list from etcd success,peers:", peers)
	return peers, nil
}

func WatchPeers(server gcache.Picker, prefix string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.GConfig.Endpoints,
		DialTimeout: time.Duration(conf.GConfig.DialTimeout) * time.Second,
	})
	if err != nil {
		fmt.Println("create etcd client failed,err:", err)
		return
	}
	//监听节点上线和下线
	for {
		wch := cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
		for ch := range wch {
			for _, event := range ch.Events {
				if event.Type == clientv3.EventTypeDelete {
					//   clusters/localhost:8002
					server.DelPeers(strings.Split(string(event.Kv.Key), "/")[1])

					fmt.Println("delete", string(event.Kv.Key))

				} else if event.Type == clientv3.EventTypePut {

					server.AddPeers(strings.Split(string(event.Kv.Key), "/")[1])

					fmt.Println("put", string(event.Kv.Key))
				} else {
					fmt.Println("watch")
				}
			}
		}
	}
}
