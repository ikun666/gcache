package etcd

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ikun666/gcache/conf"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Service struct {
	Addr     string
	IP       string
	Port     string
	Protocol string
}

// register 模块提供服务注册至 etcd 的能力
// var (
// 	// ip addr 查看虚拟机地址
// 	// DefaultEtcdConfig = clientv3.Config{
// 	// 	Endpoints:   conf.GConfig.Endpoints,
// 	// 	DialTimeout: time.Duration(conf.GConfig.DialTimeout) * time.Second,
// 	// }
// 	// Prefix = "clusters/"
// )

func Register(service *Service) {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.GConfig.Endpoints,
		DialTimeout: time.Duration(conf.GConfig.DialTimeout) * time.Second,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer cli.Close()
	jsonService, err := json.Marshal(service)
	if err != nil {
		log.Fatalln(err)
	}
	//租约
	var leaseID clientv3.LeaseID
	ctx := context.Background()
	resp, err := cli.Get(ctx, service.Addr, clientv3.WithCountOnly())
	if err != nil {
		log.Fatalln(err)
	}
	//如果etcd没有注册过这个服务才注册
	if resp.Count == 0 {
		resp, err := cli.Grant(ctx, int64(conf.GConfig.LeaseTTL))
		if err != nil {
			log.Fatalln(err)
		}
		leaseID = resp.ID
	}

	//事务
	kv := clientv3.NewKV(cli)
	txn := kv.Txn(ctx)
	_, err = txn.If(clientv3.Compare(clientv3.CreateRevision(service.Addr), "=", 0)).
		Then(
			clientv3.OpPut(conf.GConfig.Prefix+service.Addr, string(jsonService), clientv3.WithLease(leaseID)),
		).
		Else(
			clientv3.OpPut(conf.GConfig.Prefix+service.Addr, string(jsonService), clientv3.WithIgnoreLease()),
		).
		Commit()
	if err != nil {
		log.Fatalln(err)
	}
	liveChan, err := cli.KeepAlive(ctx, leaseID)
	if err != nil {
		log.Fatalln(err)
	}
	// for lease := range liveChan {
	// 	fmt.Printf("leaseID:%v,TTL:%v\n", lease.ID, lease.TTL)
	// }
	for {
		<-liveChan
	}

	//逻辑操作
}
