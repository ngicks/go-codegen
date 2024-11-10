package undgen

import (
	"cmp"
	"maps"
	"slices"
	"testing"

	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

func compareGraphIdent(i, j TypeIdent) int {
	p := cmp.Compare(i.pkgPath, j.pkgPath)
	if p != 0 {
		return p
	}
	return cmp.Compare(i.typeName, j.typeName)
}

func Test_search_type_tree(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedExportFile |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes |
			packages.NeedModule |
			packages.NeedEmbedFiles |
			packages.NeedEmbedPatterns,
		Dir: "../",
	}

	pkgs, err := packages.Load(cfg, "./intenal/searchtypetree", "./intenal/searchtypetree/sub1")
	if err != nil {
		panic(err)
	}

	graph, err := NewTypeGraph(
		pkgs,
		isUndPlainTarget,
		nil,
		nil,
	)
	assert.NilError(t, err)

	matched := slices.SortedFunc(maps.Keys(graph.matched), compareGraphIdent)
	externals := slices.SortedFunc(maps.Keys(graph.external), compareGraphIdent)

	assert.Assert(
		t,
		slices.Equal(
			[]TypeIdent{
				{
					"github.com/ngicks/go-codegen/codegen/intenal/searchtypetree",
					"Bar",
				},
				{
					"github.com/ngicks/go-codegen/codegen/intenal/searchtypetree",
					"Foo",
				},
				{
					"github.com/ngicks/go-codegen/codegen/intenal/searchtypetree/sub1",
					"Bar",
				},
				{
					"github.com/ngicks/go-codegen/codegen/intenal/searchtypetree/sub1",
					"Foo",
				},
				{
					"github.com/ngicks/go-codegen/codegen/intenal/searchtypetree/sub1",
					"HasAliasToImplementor",
				},
			},
			matched,
		),
	)
	assert.Assert(
		t,
		slices.Equal(
			[]TypeIdent{
				{"github.com/ngicks/go-codegen/codegen/intenal/searchtypetree/sub2", "Foo"},
				{"github.com/ngicks/und/option", "Option"},
			},
			externals,
		),
	)

	graph.markTransitive(nil)

	transitive := slices.SortedFunc(
		hiter.OmitL(
			xiter.Filter2(
				func(_ TypeIdent, n *TypeNode) bool {
					return n.Matched.IsTransitive()
				},
				maps.All(graph.types),
			),
		),
		compareGraphIdent,
	)

	assert.Assert(
		t,
		slices.Equal(
			[]TypeIdent{
				{"github.com/ngicks/go-codegen/codegen/intenal/searchtypetree", "HasAlias"},
				{"github.com/ngicks/go-codegen/codegen/intenal/searchtypetree/sub1", "HasAlias"},
			},
			transitive,
		),
	)

	graph.markTransitive(func(edge TypeDependencyEdge) bool { return isUndAllowedPointer(edge.ChildType, edge.Stack) })

	transitive = slices.SortedFunc(
		hiter.OmitL(
			xiter.Filter2(
				func(_ TypeIdent, n *TypeNode) bool {
					return n.Matched.IsTransitive()
				},
				maps.All(graph.types),
			),
		),
		compareGraphIdent,
	)

	assert.Assert(
		t,
		slices.Equal(
			[]TypeIdent{
				{"github.com/ngicks/go-codegen/codegen/intenal/searchtypetree", "HasAlias"},
				{"github.com/ngicks/go-codegen/codegen/intenal/searchtypetree/sub1", "HasAlias"},
			},
			transitive,
		),
	)
}
