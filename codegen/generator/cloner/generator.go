package cloner

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"iter"

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
}

func (c *Config) getMatcherConfig() *MatcherConfig {
	if c.MatcherConfig != nil {
		return c.MatcherConfig
	}
	return &MatcherConfig{}
}

func (c *Config) Generate(
	ctx context.Context,
	sourcePrinter *suffixwriter.Writer,
	pkgs []*packages.Package,
	extra []imports.TargetImport,
) error {
	parser := imports.NewParserPackages(pkgs)
	parser.AppendExtra(extra...)

	graph, err := typegraph.New(
		pkgs,
		c.getMatcherConfig().MatchType,
		codegen.ExcludeIgnoredGenDecl,
		codegen.ExcludeIgnoredTypeSpec,
	)
	if err != nil {
		return err
	}

	graph.MarkDependant(c.getMatcherConfig().MatchEdge)

	replacerData, err := graph.GatherReplaceData(
		parser,
		func(g *typegraph.Graph) iter.Seq2[typegraph.Ident, *typegraph.Node] {
			return g.EnumerateTypes()
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

		for _, node := range data.TargetNodes {
			err = generateMethod(c, buf, node, data)
			if err != nil {
				return err
			}
		}

		err = sourcePrinter.Write(ctx, data.Filename, buf.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}
