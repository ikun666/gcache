package consistentHash

import (
	"hash/crc32"
	"slices"
	"strconv"
)

// Hash maps bytes to uint32
// hash函数
type Hash func(data []byte) uint32

// Map constains all hashed keys
type Map struct {
	//hash函数
	hash Hash
	//虚拟节点倍数
	replicas int
	//有序哈希环
	hashRing []uint32
	//虚拟节点与真实节点的映射表
	hashMap map[uint32]string
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[uint32]string),
	}
	//不指定哈希函数使用默认哈希
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash.
// 添加keys到哈希环
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		//将自己加入
		hash := m.hash([]byte(key))
		m.hashRing = append(m.hashRing, hash)
		m.hashMap[hash] = key
		//每个key都要有虚拟节点 加入虚拟节点
		for i := 0; i < m.replicas; i++ {
			hash := m.hash([]byte(strconv.Itoa(i) + key))
			m.hashRing = append(m.hashRing, hash)
			m.hashMap[hash] = key
		}
	}
	slices.Sort(m.hashRing)
	// fmt.Println(m.hashRing)
}

// Remove use to remove a key and its virtual keys on the ring and map
func (m *Map) Remove(keys ...string) {
	for _, key := range keys {
		//将自己删除
		hash := m.hash([]byte(key))
		idx, _ := slices.BinarySearch(m.hashRing, hash)
		m.hashRing = append(m.hashRing[:idx], m.hashRing[idx+1:]...)
		delete(m.hashMap, hash)
		//删除虚拟节点
		for i := 0; i < m.replicas; i++ {
			hash := m.hash([]byte(strconv.Itoa(i) + key))
			// idx := sort.SearchInts(m.hashRing, hash)
			idx, _ := slices.BinarySearch(m.hashRing, hash)
			m.hashRing = append(m.hashRing[:idx], m.hashRing[idx+1:]...)
			delete(m.hashMap, hash)
		}
	}
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.hashRing) == 0 {
		return ""
	}

	hash := m.hash([]byte(key))
	//二叉搜索找第一个大于等于目标值的
	idx, ok := slices.BinarySearchFunc(m.hashRing, hash, func(e, t uint32) int {
		if e >= t {
			return 0
		}
		return -1
	})
	//没有找到大于等于的  顺时针看第一个就是他的值
	if !ok {
		idx = 0
	}
	//将虚拟节点映射成真实节点
	// fmt.Println("idx=", idx, "key=", m.hashRing[idx])
	// fmt.Println(m.hashRing)
	return m.hashMap[m.hashRing[idx]]
}
