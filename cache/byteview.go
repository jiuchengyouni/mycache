package cache

type ByteView struct {
	B []byte
}

func (v ByteView) Len() int {
	return len(v.B)
}

// 缓存值是只读的，返回一个拷贝防止被修改
func (v ByteView) ByteSlice() []byte {
	return CloneByte(v.B)
}

func (v ByteView) String() string {
	return string(v.B)
}

func CloneByte(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
