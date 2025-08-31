package undgen

import (
	"go/ast"
	"iter"

	"github.com/ngicks/go-codegen/codegen/pkg/directive"
	"github.com/ngicks/go-codegen/codegen/pkg/imports"
	"github.com/ngicks/go-codegen/codegen/pkg/typegraph"
	"golang.org/x/tools/go/packages"
)

func gatherPlainUndTypes(
	pkgs []*packages.Package,
	parser *imports.ImportParser,
	edgeFilter func(edge typegraph.Edge) bool,
	seqFactory func(g *typegraph.Graph) iter.Seq2[typegraph.Ident, *typegraph.Node],
) (data map[*ast.File]*typegraph.ReplaceData, err error) {
	graph, err := typegraph.New(
		pkgs,
		isUndPlainTarget,
		directive.ExcludeIgnoredGenDecl,
		directive.ExcludeIgnoredTypeSpec,
	)
	if err != nil {
		return nil, err
	}
	if edgeFilter != nil {
		graph.MarkDependant(edgeFilter)
	}
	return graph.GatherReplaceData(parser, seqFactory)
}

func gatherValidatableUndTypes(
	pkgs []*packages.Package,
	parser *imports.ImportParser,
	edgeFilter func(edge typegraph.Edge) bool,
	seqFactory func(g *typegraph.Graph) iter.Seq2[typegraph.Ident, *typegraph.Node],
) (data map[*ast.File]*typegraph.ReplaceData, err error) {
	graph, err := typegraph.New(
		pkgs,
		isUndValidatorTarget,
		directive.ExcludeIgnoredGenDecl,
		directive.ExcludeIgnoredTypeSpec,
	)
	if err != nil {
		return nil, err
	}
	if edgeFilter != nil {
		graph.MarkDependant(edgeFilter)
	}
	return graph.GatherReplaceData(parser, seqFactory)
}
