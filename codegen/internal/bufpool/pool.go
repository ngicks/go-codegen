package bufpool

import (
	"bytes"
	"sync"
)

var bufPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(nil)
	},
}

func GetBuf() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func PutBuf(b *bytes.Buffer) {
	if b == nil {
		return
	}
	if b.Cap() > 512<<10 {
		// My biggest (in size) blog article is like 120KiB.
		// 512KiB maximum should be large enough.
		return
	}
	bufPool.Put(b)
}
