package gcache

// A ByteView holds an immutable view of bytes.
type ByteView struct {
	b []byte
}

// Len returns the view's length
func (v ByteView) Len() int {
	return len(v.b)
}

// 返回拷贝切片防止被修改
func (v *ByteView) ByteSlice() []byte {
	// return cloneBytes(v.b)
	return v.b
}

// String returns the data as a string, making a copy if necessary.
func (v *ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
