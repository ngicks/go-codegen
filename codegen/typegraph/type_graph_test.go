package typegraph

import (
	"cmp"
	"go/ast"
	"go/types"
	"iter"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"github.com/ngicks/und/option"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

var pkgsMap map[string][]*packages.Package

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}

func init() {
	pkgsMap = make(map[string][]*packages.Package)
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes,
	}
	dirents := must(os.ReadDir("./testdata"))
	for _, dirent := range dirents {
		if !dirent.IsDir() {
			continue
		}
		pkgsMap[dirent.Name()] = must(packages.Load(cfg, "./testdata/"+dirent.Name()+"/..."))
	}
}

func compareGraphIdent(i, j Ident) int {
	p := cmp.Compare(i.PkgPath, j.PkgPath)
	if p != 0 {
		return p
	}
	return cmp.Compare(i.TypeName, j.TypeName)
}

func isFakeTargetType(n *types.Named) bool {
	pkg := n.Obj().Pkg()
	if pkg == nil {
		return false
	}
	return pkg.Path() == "github.com/ngicks/go-codegen/codegen/typegraph/testdata/faketarget" &&
		(n.Obj().Name() == "FakeTarget" || n.Obj().Name() == "FakeTarget2")
}

func collectIter(
	seq iter.Seq2[Ident, *Node],
	filter func(ident Ident, node *Node) bool,
) []Ident {
	return slices.SortedFunc(
		hiter.OmitL(
			xiter.Filter2(
				func(ident Ident, node *Node) bool {
					if filter == nil {
						return true
					}
					return filter(ident, node)
				},
				seq,
			),
		),
		compareGraphIdent,
	)
}

func firstElem[M ~map[K]V, K comparable, V any](m M) V {
	for _, v := range m {
		return v
	}
	return *new(V)
}

func Test_edges(t *testing.T) {
	pkgs := pkgsMap["edges"]
	graph, err := NewTypeGraph(
		pkgs,
		func(typeInfo *types.Named, external bool) (bool, error) {
			return isFakeTargetType(typeInfo), nil
		},
		func(gd *ast.GenDecl) (bool, error) {
			return !strings.Contains(gd.Doc.Text(), "filterGenDecl"), nil
		},
		func(ts *ast.TypeSpec, o types.Object) (bool, error) {
			return !strings.Contains(ts.Doc.Text(), "filterTypeSpec"), nil
		},
	)
	assert.NilError(t, err)

	testdataIdent := func(pkgName string, name string) Ident {
		return Ident{"github.com/ngicks/go-codegen/codegen/typegraph/testdata/" + pkgName, name}
	}

	node := graph.types[testdataIdent("edges", "MereArray")]
	child := firstElem(node.Children)[0]
	assert.Equal(t, testdataIdent("edges", "MereArray"), IdentFromTypesObject(child.ParentNode.Type.Obj()))
	assert.Equal(t, testdataIdent("faketarget", "FakeTarget2"), IdentFromTypesObject(child.ChildNode.Type.Obj()))
	assert.Equal(t, testdataIdent("faketarget", "FakeTarget2"), IdentFromTypesObject(child.ChildType.Obj()))

	assert.Assert(t, len(child.TypeArgs) == 2)

	arg0 := child.TypeArgs[0]
	assert.DeepEqual(t, []EdgeRouteNode(nil), arg0.Route)
	assert.Assert(t, arg0.Node == nil)
	assert.Assert(t, arg0.Ty == nil)
	assert.Equal(t, types.String, arg0.Org.(*types.Basic).Kind())

	arg1 := child.TypeArgs[1]
	assert.DeepEqual(t, []EdgeRouteNode{{Kind: EdgeKindPointer}}, arg1.Route)
	assert.Assert(t, arg1.Node == graph.types[testdataIdent("edges", "MereChan")])
	assert.Equal(t, testdataIdent("edges", "MereChan"), IdentFromTypesObject(arg1.Ty.Obj()))
	assert.Equal(t, testdataIdent("edges", "MereChan"), IdentFromTypesObject(arg1.Org.(*types.Pointer).Elem().(*types.Named).Obj()))

	type testCase struct {
		name   string
		assert func(*Node)
	}

	for _, tc := range []testCase{
		{
			name: "MereArray",
			assert: func(tn *Node) {
				assert.DeepEqual(t, []EdgeRouteNode{{Kind: EdgeKindArray}}, firstElem(tn.Children)[0].Stack)
			},
		},
		{
			name: "MereSlice",
			assert: func(tn *Node) {
				assert.DeepEqual(t, []EdgeRouteNode{{Kind: EdgeKindSlice}}, firstElem(tn.Children)[0].Stack)
			},
		},
		{
			name: "MereMap",
			assert: func(tn *Node) {
				assert.DeepEqual(t, []EdgeRouteNode{{Kind: EdgeKindMap}}, firstElem(tn.Children)[0].Stack)
			},
		},
		{
			name: "MereChan",
			assert: func(tn *Node) {
				assert.DeepEqual(t, []EdgeRouteNode{{Kind: EdgeKindChan}}, firstElem(tn.Children)[0].Stack)
			},
		},
		{
			name: "MereStruct",
			assert: func(tn *Node) {
				a := tn.Children[testdataIdent("faketarget", "FakeTarget")][0].Stack
				b := tn.Children[testdataIdent("faketarget", "FakeTarget2")][0].Stack
				assert.DeepEqual(t, []EdgeRouteNode{{Kind: EdgeKindStruct, Pos: option.Some(0)}, {Kind: EdgeKindPointer}}, a)
				assert.DeepEqual(t, []EdgeRouteNode{{Kind: EdgeKindStruct, Pos: option.Some(1)}}, b)
			},
		},
		{
			name: "Complex",
			assert: func(tn *Node) {
				assert.DeepEqual(
					t,
					[]EdgeRouteNode{
						{Kind: EdgeKindStruct, Pos: option.Some(0)},
						{Kind: EdgeKindPointer},
						{Kind: EdgeKindMap},
						{Kind: EdgeKindSlice},
						{Kind: EdgeKindPointer},
						{Kind: EdgeKindArray},
						{Kind: EdgeKindMap},
					},
					tn.Children[testdataIdent("edges", "MereArray")][0].Stack,
				)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.assert(graph.types[Ident{"github.com/ngicks/go-codegen/codegen/typegraph/testdata/edges", tc.name}])
		})
	}
}

func Test_filterast(t *testing.T) {
	pkgs := pkgsMap["filterast"]
	graph, err := NewTypeGraph(
		pkgs,
		func(typeInfo *types.Named, external bool) (bool, error) {
			return false, nil
		},
		func(gd *ast.GenDecl) (bool, error) {
			return !strings.Contains(gd.Doc.Text(), "filterGenDecl"), nil
		},
		func(ts *ast.TypeSpec, o types.Object) (bool, error) {
			return !strings.Contains(ts.Doc.Text(), "filterTypeSpec"), nil
		},
	)
	assert.NilError(t, err)

	types := slices.SortedFunc(maps.Keys(graph.types), compareGraphIdent)

	assert.DeepEqual(
		t,
		[]Ident{
			{
				PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filterast",
				TypeName: "Decl2",
			},
			{
				PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filterast",
				TypeName: "Decl5",
			},
		},
		types,
	)
}

func Test_filteredge(t *testing.T) {
	pkgs := pkgsMap["filteredge"]
	graph, err := NewTypeGraph(
		pkgs,
		func(typeInfo *types.Named, external bool) (bool, error) {
			return isFakeTargetType(typeInfo), nil
		},
		nil,
		nil,
	)
	assert.NilError(t, err)

	isMatchedStruct := func(n *types.Named) bool {
		pkg := n.Obj().Pkg()
		if pkg == nil {
			return false
		}
		return pkg.Path() == "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge" && n.Obj().Name() == "MatchedStruct"
	}

	type testCase struct {
		name       string
		edgeFilter func(edge Edge) bool
		expected   []Ident
	}

	for _, tc := range []testCase{
		{
			name: "only direct",
			edgeFilter: func(edge Edge) bool {
				return isMatchedStruct(edge.ParentNode.Type) ||
					(len(edge.Stack) == 1 &&
						edge.Stack[0].Kind == EdgeKindStruct)
			},
			expected: []Ident{
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "A",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "D",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "MatchedStruct",
				},
			},
		},
		{
			name: "only slice",
			edgeFilter: func(edge Edge) bool {
				return isMatchedStruct(edge.ParentNode.Type) ||
					(len(edge.Stack) == 2 &&
						edge.Stack[0].Kind == EdgeKindStruct &&
						edge.Stack[1].Kind == EdgeKindSlice)
			},
			expected: []Ident{
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "B",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "D",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "MatchedStruct",
				},
			},
		},
		{
			name: "slice and map",
			edgeFilter: func(edge Edge) bool {
				return isMatchedStruct(edge.ParentNode.Type) ||
					(len(edge.Stack) == 2 &&
						edge.Stack[0].Kind == EdgeKindStruct &&
						(edge.Stack[1].Kind == EdgeKindSlice || edge.Stack[1].Kind == EdgeKindMap))
			},
			expected: []Ident{
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "B",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "C",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "D",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "F",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "G",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "MatchedStruct",
				},
			},
		},
		{
			name: "anything but C is not allowed",
			edgeFilter: func(edge Edge) bool {
				obj := edge.ParentNode.Type.Obj()
				return obj.Name() != "C"
			},
			expected: []Ident{
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "A",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "B",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "D",
				},
				{
					PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/filteredge",
					TypeName: "MatchedStruct",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			graph.MarkDependant(tc.edgeFilter)
			idents := collectIter(
				graph.IterUpward(false, nil),
				func(ident Ident, node *Node) bool { return node.Matched.IsDependant() },
			)
			assert.DeepEqual(
				t,
				tc.expected,
				idents,
			)
			idents = collectIter(
				graph.IterUpward(true, tc.edgeFilter),
				nil,
			)
			assert.DeepEqual(
				t,
				tc.expected,
				idents,
			)
		})
	}
}
func Test_loop(t *testing.T) {
	pkgs := pkgsMap["loop"]
	graph, err := NewTypeGraph(
		pkgs,
		func(typeInfo *types.Named, external bool) (bool, error) {
			return isFakeTargetType(typeInfo), nil
		},
		nil,
		nil,
	)
	assert.NilError(t, err)

	allTypes := []Ident{
		{
			PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/loop",
			TypeName: "LoopEmbedded",
		},
		{
			PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/loop",
			TypeName: "Tree",
		},
		{
			PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/loop",
			TypeName: "recursion1",
		},
		{
			PkgPath:  "github.com/ngicks/go-codegen/codegen/typegraph/testdata/loop",
			TypeName: "recursion2",
		},
	}

	types := slices.SortedFunc(maps.Keys(graph.types), compareGraphIdent)
	assert.DeepEqual(
		t,
		allTypes,
		types,
	)

	graph.MarkDependant(nil)
	types = collectIter(
		maps.All(graph.types),
		func(ident Ident, node *Node) bool { return node.Matched.IsDependant() },
	)
	assert.DeepEqual(
		t,
		allTypes,
		types,
	)
	types = collectIter(graph.IterUpward(false, nil), nil)
	assert.DeepEqual(
		t,
		allTypes,
		types,
	)
}
