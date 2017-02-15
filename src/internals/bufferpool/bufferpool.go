package bufferpool

import (
	"bytes"
	"sync"
)

var buffers sync.Pool

// Get fetches a buffer from the buffer pool.
func Get() *bytes.Buffer {
	buf := buffers.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// Put returns a buffer to the buffer pool.
func Put(buf *bytes.Buffer) {
	if buf != nil {
		buffers.Put(buf)
	}
}

func init() {
	buffers.New = func() interface{} {
		return &bytes.Buffer{}
	}
}
