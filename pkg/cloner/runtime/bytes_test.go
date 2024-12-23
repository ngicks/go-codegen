package cloneruntime

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestCloneBytesBuffer(t *testing.T) {
	b := new(bytes.Buffer)

	_, _ = io.CopyN(b, rand.Reader, 16)

	cloned := CloneBytesBuffer(b)

	if b.String() != cloned.String() {
		t.Fatalf("wrong clone")
	}
	if b.Cap() != cloned.Cap() {
		t.Fatalf("wrong clone")
	}
}
