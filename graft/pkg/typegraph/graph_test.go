package typegraph

import (
	"cmp"
	"go/ast"
	"go/types"
	"iter"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ngicks/go-codegen/graft/internal/loader"
	"github.com/ngicks/und/option"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

var cmpOptEquateOption = cmpopts.EquateComparable(option.Option[int]{})

func compareQualifiedType(a, b QualifiedType) int {
	if c := cmp.Compare(a.PackagePath, b.PackagePath); c != 0 {
		return c
	}
	if c := cmp.Compare(a.TypeName, b.TypeName); c != 0 {
		return c
	}
	return 0
}

func keysSorted[V any](m map[QualifiedType]V) []QualifiedType {
	keys := make([]QualifiedType, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, compareQualifiedType)
	return keys
}

func collectKeysSorted[V any](seq iter.Seq2[QualifiedType, V]) []QualifiedType {
	var keys []QualifiedType
	for k := range seq {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, compareQualifiedType)
	return keys
}

func compareEdgeRoutesByFirstPos(a, b *EdgeRoute) int {
	if len(a.Nodes) == 0 || len(b.Nodes) == 0 {
		return 0
	}
	return cmp.Compare(a.Nodes[0].Pos.Value(), b.Nodes[0].Pos.Value())
}

func compareEdgeRoutes(a, b *EdgeRoute) int {
	if c := cmp.Compare(len(a.Nodes), len(b.Nodes)); c != 0 {
		return c
	}
	for i := range a.Nodes {
		if i >= len(b.Nodes) {
			return 1
		}
		if c := cmp.Compare(a.Nodes[i].Kind, b.Nodes[i].Kind); c != 0 {
			return c
		}
	}
	return 0
}

func TestNew_Basic(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/basic")

	g, err := New(pkgs, nil, nil)
	assert.NilError(t, err)

	var basicPkgPath string
	for _, pkg := range pkgs {
		if pkg.Name == "basic" {
			basicPkgPath = pkg.PkgPath
			break
		}
	}

	// All 5 types should exist
	assert.DeepEqual(t,
		[]QualifiedType{
			NewQualifiedType(basicPkgPath, "TypeA"),
			NewQualifiedType(basicPkgPath, "TypeB"),
			NewQualifiedType(basicPkgPath, "TypeC"),
			NewQualifiedType(basicPkgPath, "TypeD"),
		},
		collectKeysSorted(g.All()),
	)

	// TypeA children: TypeB, TypeC
	typeA, ok := g.Get(NewQualifiedType(basicPkgPath, "TypeA"))
	assert.Assert(t, ok)
	assert.DeepEqual(t,
		[]QualifiedType{
			NewQualifiedType(basicPkgPath, "TypeB"),
			NewQualifiedType(basicPkgPath, "TypeC"),
		},
		keysSorted(typeA.Children),
	)

	// TypeA has no parents
	assert.DeepEqual(t, []QualifiedType{}, keysSorted(typeA.Parent))

	// TypeC children: TypeD
	typeC, ok := g.Get(NewQualifiedType(basicPkgPath, "TypeC"))
	assert.Assert(t, ok)
	assert.DeepEqual(t,
		[]QualifiedType{NewQualifiedType(basicPkgPath, "TypeD")},
		keysSorted(typeC.Children),
	)

	// TypeC parent: TypeA
	assert.DeepEqual(t,
		[]QualifiedType{NewQualifiedType(basicPkgPath, "TypeA")},
		keysSorted(typeC.Parent),
	)
}

func TestNew_Edges(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/edges")

	g, err := New(pkgs, nil, nil)
	assert.NilError(t, err)

	var edgesPkgPath string
	for _, pkg := range pkgs {
		if pkg.Name == "edges" {
			edgesPkgPath = pkg.PkgPath
			break
		}
	}

	targetId := NewQualifiedType(edgesPkgPath, "Target")
	genericId := NewQualifiedType(edgesPkgPath, "GenericType")

	// SliceType -> Target
	sliceType, _ := g.Get(NewQualifiedType(edgesPkgPath, "SliceType"))
	assert.DeepEqual(t, []QualifiedType{targetId}, keysSorted(sliceType.Children))

	// ArrayType -> Target
	arrayType, _ := g.Get(NewQualifiedType(edgesPkgPath, "ArrayType"))
	assert.DeepEqual(t, []QualifiedType{targetId}, keysSorted(arrayType.Children))

	// MapType -> Target
	mapType, _ := g.Get(NewQualifiedType(edgesPkgPath, "MapType"))
	assert.DeepEqual(t, []QualifiedType{targetId}, keysSorted(mapType.Children))

	// ChanType -> Target
	chanType, _ := g.Get(NewQualifiedType(edgesPkgPath, "ChanType"))
	assert.DeepEqual(t, []QualifiedType{targetId}, keysSorted(chanType.Children))

	// PointerType -> Target
	pointerType, _ := g.Get(NewQualifiedType(edgesPkgPath, "PointerType"))
	assert.DeepEqual(t, []QualifiedType{targetId}, keysSorted(pointerType.Children))

	// StructType -> Target
	structType, _ := g.Get(NewQualifiedType(edgesPkgPath, "StructType"))
	assert.DeepEqual(t, []QualifiedType{targetId}, keysSorted(structType.Children))

	// NestedType -> Target (through *map[string][]Target)
	nestedType, _ := g.Get(NewQualifiedType(edgesPkgPath, "NestedType"))
	assert.DeepEqual(t, []QualifiedType{targetId}, keysSorted(nestedType.Children))

	// InstantiatedType -> GenericType, Target (through type arg)
	instantiatedType, _ := g.Get(NewQualifiedType(edgesPkgPath, "InstantiatedType"))
	assert.DeepEqual(t,
		[]QualifiedType{genericId, targetId},
		keysSorted(instantiatedType.Children),
	)

	// Target should have many parents
	target, _ := g.Get(targetId)
	assert.DeepEqual(t,
		[]QualifiedType{
			NewQualifiedType(edgesPkgPath, "ArrayType"),
			NewQualifiedType(edgesPkgPath, "ChanType"),
			NewQualifiedType(edgesPkgPath, "InstantiatedType"),
			NewQualifiedType(edgesPkgPath, "MapType"),
			NewQualifiedType(edgesPkgPath, "NestedType"),
			NewQualifiedType(edgesPkgPath, "PointerType"),
			NewQualifiedType(edgesPkgPath, "SliceType"),
			NewQualifiedType(edgesPkgPath, "StructType"),
		},
		keysSorted(target.Parent),
	)
}

func TestNew_WithFilters(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/basic")

	// Filter to only include TypeA and TypeB
	typeSpecFilter := func(ts *ast.TypeSpec, obj types.Object) (bool, error) {
		return ts.Name.Name == "TypeA" || ts.Name.Name == "TypeB", nil
	}

	g, err := New(pkgs, nil, typeSpecFilter)
	assert.NilError(t, err)

	var basicPkgPath string
	for _, pkg := range pkgs {
		if pkg.Name == "basic" {
			basicPkgPath = pkg.PkgPath
			break
		}
	}

	// Only TypeA and TypeB should exist
	assert.DeepEqual(t,
		[]QualifiedType{
			NewQualifiedType(basicPkgPath, "TypeA"),
			NewQualifiedType(basicPkgPath, "TypeB"),
		},
		collectKeysSorted(g.All()),
	)

	// TypeA -> TypeB (TypeC filtered out)
	typeA, _ := g.Get(NewQualifiedType(basicPkgPath, "TypeA"))
	assert.DeepEqual(t,
		[]QualifiedType{NewQualifiedType(basicPkgPath, "TypeB")},
		keysSorted(typeA.Children),
	)
}

func TestTrim_MatchAll(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/basic")

	g, err := New(pkgs, nil, nil)
	assert.NilError(t, err)

	var basicPkgPath string
	for _, pkg := range pkgs {
		if pkg.Name == "basic" {
			basicPkgPath = pkg.PkgPath
			break
		}
	}

	// Match all nodes
	mg := g.Trim(
		func(n *Node) bool { return true },
		func(e *Edge[Node]) bool { return true },
		nil,
	)

	// All nodes should be matched
	assert.DeepEqual(t,
		[]QualifiedType{
			NewQualifiedType(basicPkgPath, "TypeA"),
			NewQualifiedType(basicPkgPath, "TypeB"),
			NewQualifiedType(basicPkgPath, "TypeC"),
			NewQualifiedType(basicPkgPath, "TypeD"),
		},
		collectKeysSorted(mg.Matched()),
	)

	// TypeA and TypeC are dependants because they are parents of matched nodes
	// (markDependants walks upward from matched nodes)
	assert.DeepEqual(t,
		[]QualifiedType{
			NewQualifiedType(basicPkgPath, "TypeA"),
			NewQualifiedType(basicPkgPath, "TypeC"),
		},
		collectKeysSorted(mg.Dependant()),
	)
}

func TestTrim_MatchLeaf(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/basic")

	g, err := New(pkgs, nil, nil)
	assert.NilError(t, err)

	var basicPkgPath string
	for _, pkg := range pkgs {
		if pkg.Name == "basic" {
			basicPkgPath = pkg.PkgPath
			break
		}
	}

	typeDId := NewQualifiedType(basicPkgPath, "TypeD")

	// Match only TypeD (leaf node)
	mg := g.Trim(
		func(n *Node) bool { return n.Id == typeDId },
		func(e *Edge[Node]) bool { return true },
		nil,
	)

	// Only TypeD should be matched
	assert.DeepEqual(t,
		[]QualifiedType{typeDId},
		collectKeysSorted(mg.Matched()),
	)

	// TypeA and TypeC should be dependants (path: TypeA -> TypeC -> TypeD)
	assert.DeepEqual(t,
		[]QualifiedType{
			NewQualifiedType(basicPkgPath, "TypeA"),
			NewQualifiedType(basicPkgPath, "TypeC"),
		},
		collectKeysSorted(mg.Dependant()),
	)

	// Check MatchKind
	typeD, _ := mg.Get(typeDId)
	assert.Equal(t, MatchKindMatched, typeD.MatchKind)
	typeA, _ := mg.Get(NewQualifiedType(basicPkgPath, "TypeA"))
	assert.Equal(t, MatchKindDependant, typeA.MatchKind)
	typeC, _ := mg.Get(NewQualifiedType(basicPkgPath, "TypeC"))
	assert.Equal(t, MatchKindDependant, typeC.MatchKind)
}

func TestTrim_EdgeMatcher(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/basic")

	g, err := New(pkgs, nil, nil)
	assert.NilError(t, err)

	var basicPkgPath string
	for _, pkg := range pkgs {
		if pkg.Name == "basic" {
			basicPkgPath = pkg.PkgPath
			break
		}
	}

	typeDId := NewQualifiedType(basicPkgPath, "TypeD")
	typeCId := NewQualifiedType(basicPkgPath, "TypeC")

	// Match TypeD, but block edge from TypeC (which blocks TypeA from being a dependant)
	mg := g.Trim(
		func(n *Node) bool {
			return n.Id == typeDId
		},
		func(e *Edge[Node]) bool {
			// Block edge from TypeA to TypeC
			return e.Parent.Id != NewQualifiedType(basicPkgPath, "TypeA")
		},
		nil,
	)

	// TypeD should be matched
	assert.DeepEqual(t, []QualifiedType{typeDId}, collectKeysSorted(mg.Matched()))

	// Only TypeC is dependant (TypeA is blocked because edge TypeA->TypeC is filtered)
	assert.DeepEqual(t,
		[]QualifiedType{typeCId},
		collectKeysSorted(mg.Dependant()),
	)
}

func TestMatchKind(t *testing.T) {
	t.Run("Matched", func(t *testing.T) {
		assert.Equal(t, true, MatchKindMatched.IsMatched())
		assert.Equal(t, false, MatchKindMatched.IsDependant())
	})

	t.Run("Dependant", func(t *testing.T) {
		assert.Equal(t, false, MatchKindDependant.IsMatched())
		assert.Equal(t, true, MatchKindDependant.IsDependant())
	})

	t.Run("Combined", func(t *testing.T) {
		combined := MatchKindMatched | MatchKindDependant
		assert.Equal(t, true, combined.IsMatched())
		assert.Equal(t, true, combined.IsDependant())
	})
}

func TestTraverseTypes(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/edges")

	var edgesPkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.Name == "edges" {
			edgesPkg = pkg
			break
		}
	}

	// Find NestedType: A *map[string][]Target
	var nestedType *types.Named
	for _, obj := range edgesPkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}
		if named, ok := obj.Type().(*types.Named); ok && obj.Name() == "NestedType" {
			nestedType = named
			break
		}
	}
	assert.Assert(t, nestedType != nil)

	// Traverse and collect all named types
	var namedTypes []string
	err := WalkTypesToNamed(nil, nestedType.Underlying(), nil, func(named *types.Named, stack []EdgeRouteNode) error {
		namedTypes = append(namedTypes, named.Obj().Name())
		return nil
	})
	assert.NilError(t, err)

	slices.Sort(namedTypes)
	assert.DeepEqual(t, []string{"Target"}, namedTypes)
}

func TestTraverseTypes_InstantiatedGeneric(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/edges")

	var edgesPkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.Name == "edges" {
			edgesPkg = pkg
			break
		}
	}

	// Find InstantiatedType: G GenericType[Target, map[[2]Target][]Target]
	var instantiatedType *types.Named
	for _, obj := range edgesPkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}
		if named, ok := obj.Type().(*types.Named); ok && obj.Name() == "InstantiatedType" {
			instantiatedType = named
			break
		}
	}
	assert.Assert(t, instantiatedType != nil)

	// Traverse and collect all named types
	var namedTypes []string
	err := WalkTypesToNamed(nil, instantiatedType.Underlying(), nil, func(named *types.Named, stack []EdgeRouteNode) error {
		namedTypes = append(namedTypes, named.Obj().Name())
		return nil
	})
	assert.NilError(t, err)

	slices.Sort(namedTypes)
	// Should find GenericType once, and Target 3 times (from type arguments:
	// first type arg Target, second type arg map[[2]Target][]Target has Target in key and value)
	assert.DeepEqual(t, []string{"GenericType", "Target", "Target", "Target"}, namedTypes)
}

func TestBidirectionalEdges(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/basic")

	g, err := New(pkgs, nil, nil)
	assert.NilError(t, err)

	var basicPkgPath string
	for _, pkg := range pkgs {
		if pkg.Name == "basic" {
			basicPkgPath = pkg.PkgPath
			break
		}
	}

	typeAId := NewQualifiedType(basicPkgPath, "TypeA")
	typeBId := NewQualifiedType(basicPkgPath, "TypeB")

	typeA, _ := g.Get(typeAId)
	typeB, _ := g.Get(typeBId)

	// Get edges from both sides
	edgeFromA := typeA.Children[typeBId]
	edgeFromB := typeB.Parent[typeAId]

	// Both should reference the same edge object
	assert.Assert(t, edgeFromA == edgeFromB)

	// Verify edge structure
	assert.Equal(t, typeAId, edgeFromA.Parent.Id)
	assert.Equal(t, typeBId, edgeFromA.Child.Id)
}

func TestNew_Edges_DetailedRoutes(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/edges")

	g, err := New(pkgs, nil, nil)
	assert.NilError(t, err)

	var edgesPkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.Name == "edges" {
			edgesPkg = pkg
			break
		}
	}
	edgesPkgPath := edgesPkg.PkgPath

	targetId := NewQualifiedType(edgesPkgPath, "Target")
	genericId := NewQualifiedType(edgesPkgPath, "GenericType")

	t.Run("SliceType routes", func(t *testing.T) {
		// SliceType = []Target
		// Route: underlying is []Target, so route is just [Slice]
		sliceType, _ := g.Get(NewQualifiedType(edgesPkgPath, "SliceType"))
		edge := sliceType.Children[targetId]
		assert.Assert(t, edge != nil)
		assert.DeepEqual(t, &EdgeRoute{Nodes: []EdgeRouteNode{
			{Kind: EdgeKindSlice},
		}}, edge.Routes[0], cmpOptEquateOption)
	})

	t.Run("ArrayType routes", func(t *testing.T) {
		// ArrayType = [5]Target
		// Route: underlying is [5]Target, so route is just [Array]
		arrayType, _ := g.Get(NewQualifiedType(edgesPkgPath, "ArrayType"))
		edge := arrayType.Children[targetId]
		assert.Assert(t, edge != nil)
		assert.DeepEqual(t, &EdgeRoute{Nodes: []EdgeRouteNode{
			{Kind: EdgeKindArray},
		}}, edge.Routes[0], cmpOptEquateOption)
	})

	t.Run("MapType routes", func(t *testing.T) {
		// MapType = map[string]Target
		// Route: underlying is map[string]Target, so route is just [MapValue]
		mapType, _ := g.Get(NewQualifiedType(edgesPkgPath, "MapType"))
		edge := mapType.Children[targetId]
		assert.Assert(t, edge != nil)
		assert.DeepEqual(t, &EdgeRoute{Nodes: []EdgeRouteNode{
			{Kind: EdgeKindMapValue},
		}}, edge.Routes[0], cmpOptEquateOption)
	})

	t.Run("ChanType routes", func(t *testing.T) {
		// ChanType = chan Target
		// Route: underlying is chan Target, so route is just [Chan]
		chanType, _ := g.Get(NewQualifiedType(edgesPkgPath, "ChanType"))
		edge := chanType.Children[targetId]
		assert.Assert(t, edge != nil)
		assert.DeepEqual(t, &EdgeRoute{Nodes: []EdgeRouteNode{
			{Kind: EdgeKindChan},
		}}, edge.Routes[0], cmpOptEquateOption)
	})

	t.Run("PointerType routes", func(t *testing.T) {
		// PointerType = *Target
		// Route: underlying is *Target, so route is just [Pointer]
		pointerType, _ := g.Get(NewQualifiedType(edgesPkgPath, "PointerType"))
		edge := pointerType.Children[targetId]
		assert.Assert(t, edge != nil)
		assert.DeepEqual(t, &EdgeRoute{Nodes: []EdgeRouteNode{
			{Kind: EdgeKindPointer},
		}}, edge.Routes[0], cmpOptEquateOption)
	})

	t.Run("StructType routes", func(t *testing.T) {
		// StructType struct { A Target; B *Target; C []Target }
		// Multiple routes to Target
		structType, _ := g.Get(NewQualifiedType(edgesPkgPath, "StructType"))
		edge := structType.Children[targetId]
		assert.Assert(t, edge != nil)
		// Sort routes by first node's pos for stable comparison
		routes := slices.Clone(edge.Routes)
		slices.SortFunc(routes, compareEdgeRoutesByFirstPos)
		// A Target (field 0)
		assert.DeepEqual(t, &EdgeRoute{Nodes: []EdgeRouteNode{
			{Kind: EdgeKindStruct, Pos: option.Some(0)},
		}}, routes[0], cmpOptEquateOption)
		// B *Target (field 1)
		assert.DeepEqual(t, &EdgeRoute{Nodes: []EdgeRouteNode{
			{Kind: EdgeKindStruct, Pos: option.Some(1)},
			{Kind: EdgeKindPointer},
		}}, routes[1], cmpOptEquateOption)
		// C []Target (field 2)
		assert.DeepEqual(t, &EdgeRoute{Nodes: []EdgeRouteNode{
			{Kind: EdgeKindStruct, Pos: option.Some(2)},
			{Kind: EdgeKindSlice},
		}}, routes[2], cmpOptEquateOption)
	})

	t.Run("NestedType routes", func(t *testing.T) {
		// NestedType struct { A *map[string][]Target }
		nestedType, _ := g.Get(NewQualifiedType(edgesPkgPath, "NestedType"))
		edge := nestedType.Children[targetId]
		assert.Assert(t, edge != nil)
		assert.DeepEqual(t, &EdgeRoute{Nodes: []EdgeRouteNode{
			{Kind: EdgeKindStruct, Pos: option.Some(0)},
			{Kind: EdgeKindPointer},
			{Kind: EdgeKindMapValue},
			{Kind: EdgeKindSlice},
		}}, edge.Routes[0], cmpOptEquateOption)
	})

	t.Run("InstantiatedType routes", func(t *testing.T) {
		instantiatedType, _ := g.Get(NewQualifiedType(edgesPkgPath, "InstantiatedType"))

		edgeToGeneric := instantiatedType.Children[genericId]

		assert.Assert(t, edgeToGeneric != nil)
		assert.DeepEqual(
			t,
			&EdgeRoute{
				Nodes: []EdgeRouteNode{
					{Kind: EdgeKindStruct, Pos: option.Some(0)},
				},
			},
			edgeToGeneric.Routes[0],
			cmpOptEquateOption,
		)

		edgeToTarget := instantiatedType.Children[targetId]

		assert.Assert(t, edgeToTarget != nil)

		routes := slices.Clone(edgeToTarget.Routes)
		slices.SortFunc(routes, compareEdgeRoutes)

		assert.DeepEqual(
			t,
			&EdgeRoute{
				Nodes: []EdgeRouteNode{
					{Kind: EdgeKindStruct, Pos: option.Some(0)},
					{Kind: EdgeKindTypeArg, Pos: option.Some(0)},
				},
			},
			routes[0],
			cmpOptEquateOption,
		)
		assert.DeepEqual(
			t,
			&EdgeRoute{
				Nodes: []EdgeRouteNode{
					{Kind: EdgeKindStruct, Pos: option.Some(0)},
					{Kind: EdgeKindTypeArg, Pos: option.Some(1)},
					{Kind: EdgeKindMapKey},
					{Kind: EdgeKindArray},
				},
			},
			routes[1],
			cmpOptEquateOption,
		)
		assert.DeepEqual(
			t,
			&EdgeRoute{
				Nodes: []EdgeRouteNode{
					{Kind: EdgeKindStruct, Pos: option.Some(0)},
					{Kind: EdgeKindTypeArg, Pos: option.Some(1)},
					{Kind: EdgeKindMapValue},
					{Kind: EdgeKindSlice},
				},
			},
			routes[2],
			cmpOptEquateOption,
		)
	})
}

func TestNew_Edges_TypeDetail(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/edges")

	g, err := New(pkgs, nil, nil)
	assert.NilError(t, err)

	var edgesPkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.Name == "edges" {
			edgesPkg = pkg
			break
		}
	}
	edgesPkgPath := edgesPkg.PkgPath

	t.Run("Target TypeDetail", func(t *testing.T) {
		target, ok := g.Get(NewQualifiedType(edgesPkgPath, "Target"))
		assert.Assert(t, ok)
		assert.Assert(t, target.Detail != nil)
		assert.Equal(t, edgesPkg, target.Detail.Pkg)
		assert.Assert(t, target.Detail.File != nil)
		assert.Assert(t, target.Detail.Ts != nil)
		assert.Equal(t, "Target", target.Detail.Ts.Name.Name)
		assert.Assert(t, target.Detail.Type != nil)
		assert.Equal(t, "Target", target.Detail.Type.Obj().Name())
	})

	t.Run("GenericType TypeDetail", func(t *testing.T) {
		genericType, ok := g.Get(NewQualifiedType(edgesPkgPath, "GenericType"))
		assert.Assert(t, ok)
		assert.Assert(t, genericType.Detail != nil)
		assert.Equal(t, edgesPkg, genericType.Detail.Pkg)
		assert.Assert(t, genericType.Detail.Ts != nil)
		assert.Equal(t, "GenericType", genericType.Detail.Ts.Name.Name)
		// Verify it's a generic type with 2 type params
		assert.Assert(t, genericType.Detail.Type != nil)
		assert.Equal(t, 2, genericType.Detail.Type.TypeParams().Len())
	})

	t.Run("Pos ordering for grouped type declarations", func(t *testing.T) {
		// SliceType, ArrayType, MapType, ChanType, PointerType are in the same type block
		sliceType, _ := g.Get(NewQualifiedType(edgesPkgPath, "SliceType"))
		arrayType, _ := g.Get(NewQualifiedType(edgesPkgPath, "ArrayType"))
		mapType, _ := g.Get(NewQualifiedType(edgesPkgPath, "MapType"))
		chanType, _ := g.Get(NewQualifiedType(edgesPkgPath, "ChanType"))
		pointerType, _ := g.Get(NewQualifiedType(edgesPkgPath, "PointerType"))

		// All in same file
		assert.Equal(t, sliceType.Detail.File, arrayType.Detail.File)
		assert.Equal(t, sliceType.Detail.File, mapType.Detail.File)
		assert.Equal(t, sliceType.Detail.File, chanType.Detail.File)
		assert.Equal(t, sliceType.Detail.File, pointerType.Detail.File)

		// Pos should be increasing (they're declared in order)
		assert.Assert(t, sliceType.Detail.Pos < arrayType.Detail.Pos)
		assert.Assert(t, arrayType.Detail.Pos < mapType.Detail.Pos)
		assert.Assert(t, mapType.Detail.Pos < chanType.Detail.Pos)
		assert.Assert(t, chanType.Detail.Pos < pointerType.Detail.Pos)
	})
}
