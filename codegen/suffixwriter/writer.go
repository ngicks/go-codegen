package suffixwriter

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Writer writes data to suffixed file.
type Writer struct {
	cwd         string
	suffix      string
	fileFactory func(name string) (io.WriteCloser, error)
	preProcess  PreProcess
	postProcess PostProcess
	logf        func(format string, args ...any)
}

type PreProcess func(name string) error

type PostProcess func(ctx context.Context, src []byte) ([]byte, error)

var checkGoimportsOnce = sync.OnceValue(CheckGoimports)

func New(suffix string, opts ...Option) *Writer {
	p := &Writer{
		cwd:    "", // where the command is invoked
		suffix: suffix,
		fileFactory: func(name string) (io.WriteCloser, error) {
			return os.Create(name)
		},
		preProcess:  func(name string) error { return checkGoimportsOnce() },
		postProcess: ApplyGoimports,
		logf:        func(format string, args ...any) {},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type Option func(p *Writer)

func WithCwd(cwd string) Option {
	return func(p *Writer) {
		p.cwd = cwd
	}
}

func WithFileFactory(fileFactory func(name string) (io.WriteCloser, error)) Option {
	return func(p *Writer) {
		p.fileFactory = fileFactory
	}
}

func WithPreProcess(preProcess PreProcess) Option {
	return func(p *Writer) {
		p.preProcess = preProcess
	}
}

func WithPostProcess(postProcess PostProcess) Option {
	return func(p *Writer) {
		p.postProcess = postProcess
	}
}

func WithLogf(logf func(format string, args ...any)) Option {
	return func(p *Writer) {
		p.logf = logf
	}
}

// openFile opens name suffixed with p.suffix.
// It returns an error if name is not under cwd.
func (p *Writer) openFile(name string) (w io.WriteCloser, filename string, err error) {
	if p.cwd == "" {
		p.cwd, err = os.Getwd()
		if err != nil {
			return nil, "", fmt.Errorf("getting cwd: %w", err)
		}
	}
	rel, err := filepath.Rel(p.cwd, name)
	if err != nil {
		return nil, "", err
	}

	if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return nil, "", fmt.Errorf("generated target file is not under cwd: %s", rel)
	}

	filename = suffixFilename(rel, p.suffix)

	p.logf("open: %s\n", filename)
	w, err = p.fileFactory(filename)
	return
}

func suffixFilename(f, suffix string) string {
	ext := filepath.Ext(filepath.Base(f))
	f, _ = strings.CutSuffix(f, ext)
	return f + suffix + ext
}

// Write write b into name but suffixed.
func (p *Writer) Write(ctx context.Context, name string, b []byte) error {
	err := p.preProcess(name)
	if err != nil {
		return fmt.Errorf("preprocessing %q: %w", name, err)
	}
	w, filename, err := p.openFile(name)
	if err != nil {
		return fmt.Errorf("opening %q(for %q): %w", filename, name, err)
	}
	processed, err := p.postProcess(ctx, b)
	if err != nil {
		return fmt.Errorf("postprocessing input for %q: %w", filename, err)
	}
	_, err = w.Write(processed)
	if err != nil {
		return fmt.Errorf("writing to %q: %w", filename, err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("closing %q: %w", filename, err)
	}
	return nil
}
