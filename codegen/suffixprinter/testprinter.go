package suffixprinter

import (
	"bytes"
	"io"
	"maps"
	"sync"
)

type TestPrinter struct {
	*Printer
	mu      sync.Mutex
	results map[string][]byte
}

func NewTestPrinter(suffix string) *TestPrinter {
	p := &TestPrinter{
		results: make(map[string][]byte),
	}
	p.Printer = New(
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
	p    *TestPrinter
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

func (p *TestPrinter) Results() map[string][]byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	return maps.Clone(p.results)
}
