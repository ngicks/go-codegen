package suffixwriter

import (
	"bytes"
	"io"
	"maps"
	"sync"
)

type TestWriter struct {
	*Writer
	mu      sync.Mutex
	results map[string][]byte
}

func NewTestWriter(suffix string) *TestWriter {
	p := &TestWriter{
		results: make(map[string][]byte),
	}
	p.Writer = New(
		suffix,
		WithFileFactory(func(name string) (io.WriteCloser, error) {
			return &testPrinterWriter{
				p:    p,
				name: name,
				buf:  new(bytes.Buffer),
			}, nil
		}),
	)
	return p
}

type testPrinterWriter struct {
	p    *TestWriter
	name string
	buf  *bytes.Buffer
}

func (w *testPrinterWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *testPrinterWriter) Close() error {
	w.p.mu.Lock()
	defer w.p.mu.Unlock()
	w.p.results[w.name] = w.buf.Bytes()
	return nil
}

func (p *TestWriter) Results() map[string][]byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	return maps.Clone(p.results)
}
