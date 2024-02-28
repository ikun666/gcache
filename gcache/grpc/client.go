package grpcserver

import (
	"context"

	"github.com/ikun666/gcache"
	"github.com/ikun666/gcache/pb/gcachepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// client 模块实现了 groupcache 访问其他远程节点从而获取缓存的能力
type client struct {
	name string // 服务名称 ip:port
}

// Fetch 从 remote peer 获取对应的缓存值
func (c *client) Fetch(group string, key string) ([]byte, error) {
	conn, err := grpc.Dial(c.name, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	grpcClient := gcachepb.NewGroupCacheClient(conn)
	resp, err := grpcClient.Get(context.Background(), &gcachepb.Request{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, err
	}
	return resp.Value, err
}

func NewClient(service string) *client {
	return &client{name: service}
}

// 测试 client 是否实现了 Fetcher 接口
var _ gcache.Fetcher = (*client)(nil)

// // Fetcher 定义了从远端获取缓存的能力，所以每个 Peer 都应实现这个接口
// type Fetcher interface {
// 	Fetch(group string, key string) ([]byte, error)
// }
