package bufferpool

import (
	"bytes"
	"sync"
)

var buffers sync.Pool

// Get fetches a buffer from the buffer pool.
func Get() *bytes.Buffer {
	return buffers.Get().(*bytes.Buffer)
}

// Put returns a buffer to the buffer pool.
func Put(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset()
		buffers.Put(buf)
	}
}

func init() {
	buffers.New = func() interface{} {
		return &bytes.Buffer{}
	}
}
