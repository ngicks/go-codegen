package cloner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"io"
	"iter"
	"log/slog"
	"slices"

	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/codegen"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-codegen/codegen/typegraph"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
)

type Config struct {
	MatcherConfig *MatcherConfig
	Logger        *slog.Logger
}

func (c *Config) matcherConfig() *MatcherConfig {
	if c.MatcherConfig != nil {
		conf := *c.MatcherConfig
		conf.CustomHandlers = append(slices.Clip(conf.CustomHandlers), builtinCustomHandlers[:]...)
		conf.logger = c.logger()
		return &conf
	}
	return &MatcherConfig{CustomHandlers: builtinCustomHandlers[:], logger: c.logger()}
}

var (
	// use DiscardHandler after Go 1.24
	noopLogger = slog.New(slog.NewTextHandler(io.Discard, nil))
)

func (c *Config) logger() *slog.Logger {
	if c.Logger == nil {
		return noopLogger
	}
	return c.Logger
}

func (c *Config) Generate(
	ctx context.Context,
	sourcePrinter *suffixwriter.Writer,
	pkgs []*packages.Package,
) error {
	parser := imports.NewParserPackages(pkgs)
	parser.AppendExtra(c.matcherConfig().CustomHandlers.Imports()...)

	graph, err := typegraph.New(
		pkgs,
		c.matcherConfig().MatchType,
		codegen.ExcludeIgnoredGenDecl,
		codegen.ExcludeIgnoredTypeSpec,
		typegraph.WithPrivParser(parseNode),
	)
	if err != nil {
		return err
	}

	graph.MarkDependant(c.matcherConfig().MatchEdge)

	replacerData, err := graph.GatherReplaceData(
		parser,
		func(g *typegraph.Graph) iter.Seq2[typegraph.Ident, *typegraph.Node] {
			return g.IterUpward(true, c.matcherConfig().MatchEdge)
		},
	)
	if err != nil {
		return err
	}

	for _, data := range xiter.Filter2(
		func(f *ast.File, data *typegraph.ReplaceData) bool { return f != nil && data != nil },
		hiter.MapKeys(replacerData, pkgsutil.EnumerateFile(pkgs)),
	) {
		if len(data.TargetNodes) == 0 {
			continue
		}

		data.ImportMap.AddMissingImports(data.DstFile)
		res := decorator.NewRestorer()
		af, err := res.RestoreFile(data.DstFile)
		if err != nil {
			return fmt.Errorf("converting dst to ast for %q: %w", data.Filename, err)
		}

		buf := new(bytes.Buffer) // pool buf?

		if err := codegen.PrintFileHeader(buf, af, res.Fset); err != nil {
			return fmt.Errorf("%q: %w", data.Filename, err)
		}

		handled := 0
		for _, node := range data.TargetNodes {
			err = generateMethod(c, buf, graph, node, data)
			if err != nil {
				if errors.Is(err, errNotHandled) {
					continue
				}
				return err
			}
			handled++
		}

		if handled > 0 {
			err = sourcePrinter.Write(ctx, data.Filename, buf.Bytes())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
