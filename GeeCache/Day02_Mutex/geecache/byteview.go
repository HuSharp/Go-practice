package geecache

// ByteView 抽象一个 只读数据结构， 用来表示缓存值
type ByteView struct {
	b []byte	// b 用来存储真实值
}

// Len 用来返回其所占的内存大小
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 用来返回一个拷贝， 防止缓存值被外部修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String returns the data as a string, making a copy if necessary.
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	newB := make([]byte, len(b))
	copy(newB, b)
	return newB
}
