package undgen

import (
	"go/ast"
	"iter"

	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-codegen/codegen/typegraph"
	"golang.org/x/tools/go/packages"
)

func gatherPlainUndTypes(
	pkgs []*packages.Package,
	parser *imports.ImportParser,
	edgeFilter func(edge typegraph.TypeDependencyEdge) bool,
	seqFactory func(g *typegraph.TypeGraph) iter.Seq2[typegraph.TypeIdent, *typegraph.TypeNode],
) (data map[*ast.File]*typegraph.ReplaceData, err error) {
	graph, err := typegraph.NewTypeGraph(
		pkgs,
		isUndPlainTarget,
		excludeUndIgnoredCommentedGenDecl,
		excludeUndIgnoredCommentedTypeSpec,
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
	edgeFilter func(edge typegraph.TypeDependencyEdge) bool,
	seqFactory func(g *typegraph.TypeGraph) iter.Seq2[typegraph.TypeIdent, *typegraph.TypeNode],
) (data map[*ast.File]*typegraph.ReplaceData, err error) {
	graph, err := typegraph.NewTypeGraph(
		pkgs,
		isUndValidatorTarget,
		excludeUndIgnoredCommentedGenDecl,
		excludeUndIgnoredCommentedTypeSpec,
	)
	if err != nil {
		return nil, err
	}
	if edgeFilter != nil {
		graph.MarkDependant(edgeFilter)
	}
	return graph.GatherReplaceData(parser, seqFactory)
}
