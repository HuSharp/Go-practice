package consistenthash

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
// 函数类型 Hash，采取依赖注入的方式，允许用于替换成自定义的 Hash 函数，也方便测试时替换。
type Hash func(data []byte) uint32

// Map contains all hashed keys
type Map struct {
	hash		Hash
	replicas	int		// 虚拟节点倍数
	keys		[]int	// Sorted 哈希环
	hashMap		map[int]string	// 虚拟节点和真实节点的映射表
}

// New create a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE	// 默认为 crc32.ChecksumIEEE 算法
	}
	return m
}

// Add 传入 0 个或多个真实节点名称
func (m *Map) Add(nodeName ...string) {
	for _, key := range nodeName {
		// 对应创建虚拟节点
		for i := 0; i < m.replicas; i++ {
			// strconv包实现了基本数据类型和其字符串表示的相互转换。Itoa 是FormatInt(i, 10) 的简写。
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))	// 通过添加编号的方式区分不同虚拟节点
			m.keys = append(m.keys, hash)	// 增加虚拟节点 hash
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)	// 递增顺序
}

// Get  通过 key 获取匹配节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	// 通过计算哈希值来选择节点
	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica
	// 顺时针找到第一个匹配的虚拟节点的下标 idx
	searchIdx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[searchIdx%len(m.keys)]]// 映射真实节点
}

// Remove 删除只需要删除掉节点对应的虚拟节点和映射关系，至于均摊给其他节点，那是删除之后自然会发生的
func (m *Map) Remove(key string) {
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
		idx := sort.SearchInts(m.keys, hash)
		// 此处表示 slice 切片内部的元素被打散传入
		m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
		delete(m.hashMap, hash)
	}
}

func (m *Map) Show() string {
	var str string
	l := len(m.keys)
	for i := 0; i < l; i++ {
		k := m.keys[i]
		v := m.hashMap[k]
		s := fmt.Sprintf("%s %d\n", v, k)
		str += s
	}

	return str
}