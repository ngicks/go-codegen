package cloneruntime

import "bytes"

func CloneBytesBuffer(b *bytes.Buffer) *bytes.Buffer {
	if b == nil {
		return nil
	}
	buf := b.Bytes()
	newBuf := make([]byte, len(buf), cap(buf))
	copy(newBuf, buf)
	return bytes.NewBuffer(newBuf)
}
