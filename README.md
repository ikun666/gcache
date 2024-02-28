# gcache模仿了 groupcache 的实现( Go 语言版的 memcached)


单机缓存和基于 HTTP/GRPC 的分布式缓存


最近最少访问(Least Recently Used, LRU) 缓存策略


使用 Go 锁机制防止缓存击穿


使用一致性哈希选择节点，虚拟节点实现负载均衡


使用 protobuf 优化节点间二进制通信

使用etcd cluster 实现服务注册、发现、动态上下线

