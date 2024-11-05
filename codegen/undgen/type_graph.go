package undgen

import (
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"reflect"
	"slices"

	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/und/option"
	"golang.org/x/tools/go/packages"
)

// typeGraph enumerates type decls in given []*packages.Package and forms a type-dependency graph.
// It lists types which matches input matcher.
// Callers can traverse graph from a node upwards and downwards.
//
// .         ...                               +──────+
// .         ┠─...                             │ node │────...
// +──────+  ┠─(as struct member)──────────────+──────+
// │ node │──╂─(as struct member map elem)─────│ node │────...
// +──────+  ┠─(as struct member elt of [][]T)─+──────+
// .         ┠─...                             │ node │────...
// .         ...                               +──────+
//
// Nodes are connected by [typeDependencyEdge].
type typeGraph struct {
	types    map[typeIdent]*typeNode
	matched  map[typeIdent]*typeNode
	external map[typeIdent]*typeNode
}

type typeIdent struct {
	pkgPath  string
	typeName string
}

func (t typeIdent) targetType() TargetType {
	return TargetType{t.pkgPath, t.typeName}
}

func typeIdentFromTypesObject(obj types.Object) typeIdent {
	return typeIdent{
		obj.Pkg().Path(),
		obj.Name(),
	}
}

type typeNode struct {
	parent   map[typeIdent][]typeDependencyEdge
	children map[typeIdent][]typeDependencyEdge

	matched typeNodeMatchKind

	pkg  *packages.Package
	file *ast.File
	// nth type spec in the file.
	pos      int
	ts       *ast.TypeSpec
	typeInfo *types.Named
}

type typeNodeMatchKind uint64

const (
	typeNodeMatchKindMatched = typeNodeMatchKind(1 << iota)
	typeNodeMatchKindTransitive
	typeNodeMatchKindExternal
)

func (k typeNodeMatchKind) IsMatched() bool {
	return k&typeNodeMatchKindMatched > 0
}

func (k typeNodeMatchKind) IsTransitive() bool {
	return k&typeNodeMatchKindTransitive > 0
}

func (k typeNodeMatchKind) IsExternal() bool {
	return k&typeNodeMatchKindExternal > 0
}

type typeDependencyEdge struct {
	stack    []typeDependencyEdgePointer
	typeArgs []typeArg
	// non-instantiated parent
	parentNode *typeNode
	// instantiated child
	childType *types.Named
	// non-instantiated child node.
	childNode *typeNode
}

func (e typeDependencyEdge) hasSingleNamedTypeArg(additionalCond func(named *types.Named) bool) bool {
	if len(e.typeArgs) != 1 {
		return false
	}
	arg := e.typeArgs[0]
	if arg.org != arg.ty { // current implementation restriction: type params where []T or map[string]T is not allowed.
		return false
	}
	named, ok := arg.org.(*types.Named)
	if !ok {
		return false
	}
	if additionalCond != nil {
		return additionalCond(named)
	}
	return true
}

type typeDependencyEdgePointer struct {
	kind typeDependencyEdgeKind
	pos  option.Option[int]
}

type typeArg struct {
	stack []typeDependencyEdgePointer
	node  *typeNode
	ty    *types.Named
	org   types.Type
}

type typeDependencyEdgeKind uint64

const (
	typeDependencyEdgeKindAlias = typeDependencyEdgeKind(1 << iota)
	typeDependencyEdgeKindArray
	typeDependencyEdgeKindChan
	typeDependencyEdgeKindInterface
	typeDependencyEdgeKindMap
	typeDependencyEdgeKindNamed
	typeDependencyEdgeKindPointer
	typeDependencyEdgeKindSlice
	typeDependencyEdgeKindStruct
)

func firstTypeIdent(m map[typeIdent][]typeDependencyEdge) (typeIdent, typeDependencyEdge) {
	for k, e := range m {
		return k, e[0]
	}
	return typeIdent{}, typeDependencyEdge{}
}

func newTypeGraph(
	pkgs []*packages.Package,
	matcher func(typeInfo *types.Named, within bool) (bool, error),
	genDeclFilter func(*ast.GenDecl) (bool, error),
	typeSpecFilter func(*ast.TypeSpec, types.Object) (bool, error),
) (*typeGraph, error) {
	graph := &typeGraph{
		types:    make(map[typeIdent]*typeNode),
		matched:  make(map[typeIdent]*typeNode),
		external: make(map[typeIdent]*typeNode),
	}
	err := graph.listTypes(pkgs, matcher, genDeclFilter, typeSpecFilter)
	if err != nil {
		return graph, err
	}
	err = graph.buildEdge(matcher)
	if err != nil {
		return graph, err
	}
	return graph, nil
}

func (g *typeGraph) listTypes(
	pkgs []*packages.Package,
	matcher func(named *types.Named, external bool) (bool, error),
	genDeclFilter func(*ast.GenDecl) (bool, error),
	typeSpecFilter func(*ast.TypeSpec, types.Object) (bool, error),
) error {
	for pkg, fileSeq := range pkgsutil.EnumerateGenDecls(pkgs) {
		if err := pkgsutil.LoadError(pkg); err != nil {
			return err
		}
		for file, seq := range fileSeq {
			var pos int
			for genDecl := range seq {
				if genDecl.Tok != token.TYPE {
					continue
				}
				if genDeclFilter != nil {
					ok, err := genDeclFilter(genDecl)
					if err != nil {
						return err
					}
					if !ok {
						continue
					}
				}
				for _, s := range genDecl.Specs {
					currentPos := pos
					pos++
					ts := s.(*ast.TypeSpec)
					obj := pkg.TypesInfo.Defs[ts.Name]
					if typeSpecFilter != nil {
						ok, err := typeSpecFilter(ts, obj)
						if err != nil {
							return err
						}
						if !ok {
							continue
						}
					}

					named, ok := obj.Type().(*types.Named)
					if !ok {
						continue
					}

					node := addType(g.types, pkg, file, currentPos, ts, named)
					ok, err := matcher(named, false)
					if err != nil {
						return err
					}
					if ok {
						node.matched |= typeNodeMatchKindMatched
						g.matched[typeIdentFromTypesObject(node.typeInfo.Obj())] = node
					}
				}
			}
		}
	}
	return nil
}

func addType(
	to map[typeIdent]*typeNode,
	pkg *packages.Package,
	file *ast.File,
	pos int,
	ts *ast.TypeSpec,
	typeInfo *types.Named,
) *typeNode {
	ident := typeIdentFromTypesObject(typeInfo.Obj())
	n, ok := to[ident]
	if ok {
		return n
	}
	n = &typeNode{
		pkg:      pkg,
		file:     file,
		pos:      pos,
		ts:       ts,
		typeInfo: typeInfo,
	}
	to[ident] = n
	return n
}

func (g *typeGraph) buildEdge(matcher func(named *types.Named, external bool) (bool, error)) error {
	for _, node := range g.types {
		// Underlying matches what of go spec.
		// It means what follows type idents like below:
		//
		// type Foo struct {Foo string; Bar int}
		//          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^ this part is underlying
		//
		err := visitTypes(
			node,
			node.typeInfo.Underlying(),
			matcher,
			g.types,
			g.external,
			nil,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// visitTypes visits
func visitTypes(
	parentNode *typeNode,
	ty types.Type,
	matcher func(named *types.Named, external bool) (bool, error),
	allType map[typeIdent]*typeNode,
	externalType map[typeIdent]*typeNode,
	stack []typeDependencyEdgePointer,
) error {
	return visitToNamed(
		ty,
		func(named *types.Named, stack []typeDependencyEdgePointer) error {
			node, ok := allType[typeIdentFromTypesObject(named.Obj())]
			if ok {
				parentNode.drawEdge(stack, visitOnTypeArgs(named.TypeArgs(), matcher, allType, externalType), named, node)
				return nil
			}
			ok, err := matcher(named, true)
			if !ok || err != nil {
				return err
			}
			externalNode := addType(externalType, nil, nil, -1, nil, named)
			externalNode.matched |= typeNodeMatchKindExternal
			parentNode.drawEdge(stack, visitOnTypeArgs(named.TypeArgs(), matcher, allType, externalType), named, externalNode)
			return nil
		},
		stack,
	)
}

func visitOnTypeArgs(
	typeList *types.TypeList,
	matcher func(named *types.Named, external bool) (bool, error),
	allType map[typeIdent]*typeNode,
	externalType map[typeIdent]*typeNode,
) []typeArg {
	var typeArgs []typeArg
	for _, arg := range hiter.AtterAll(typeList) {
		var found bool
		_ = visitToNamed(
			arg,
			func(named *types.Named, stack []typeDependencyEdgePointer) error {
				found = true
				// TODO: split this `check if internal types, if not, then add as external type` sequence as a function or a method.
				node, ok := allType[typeIdentFromTypesObject(named.Obj())]
				if !ok {
					ok, err := matcher(named, true)
					if ok && err == nil {
						externalNode := addType(externalType, nil, nil, -1, nil, named)
						externalNode.matched |= typeNodeMatchKindExternal
						node = externalNode
					}
				}
				typeArgs = append(typeArgs, typeArg{
					stack: slices.Clone(stack),
					node:  node, // might still be nil.
					ty:    named,
					org:   arg,
				})
				return nil
			},
			nil,
		)
		if !found {
			typeArgs = append(typeArgs, typeArg{
				org: arg,
			})
		}
	}
	return typeArgs
}

func visitToNamed(
	ty types.Type,
	cb func(named *types.Named, stack []typeDependencyEdgePointer) error,
	stack []typeDependencyEdgePointer,
) error {
	// types may recurse.
	// but should be impossible without naming type,
	// which breaks visitToNamed from infinite loop.
	switch x := ty.(type) {
	default:
		return nil
	case *types.Named:
		return cb(x, stack)
	case *types.Alias:
		// TODO: check for type param after go1.24
		// see https://github.com/golang/go/issues/46477
		return visitToNamed(x.Rhs(), cb, append(stack, typeDependencyEdgePointer{kind: typeDependencyEdgeKindAlias}))
	case *types.Array:
		return visitToNamed(x.Elem(), cb, append(stack, typeDependencyEdgePointer{kind: typeDependencyEdgeKindArray}))
	case *types.Basic:
		return nil
	case *types.Chan:
		return visitToNamed(x.Elem(), cb, append(stack, typeDependencyEdgePointer{kind: typeDependencyEdgeKindChan}))
	case *types.Interface:
		return nil
	case *types.Map:
		return visitToNamed(x.Elem(), cb, append(stack, typeDependencyEdgePointer{kind: typeDependencyEdgeKindMap}))
	case *types.Pointer:
		return visitToNamed(x.Elem(), cb, append(stack, typeDependencyEdgePointer{kind: typeDependencyEdgeKindPointer}))
	case *types.Slice:
		return visitToNamed(x.Elem(), cb, append(stack, typeDependencyEdgePointer{kind: typeDependencyEdgeKindSlice}))
	case *types.Struct:
		// We don't support type-parametrized struct fields.
		// Thus not checking type args.
		for i := range x.NumFields() {
			f := x.Field(i)
			err := visitToNamed(f.Type(), cb, append(stack, typeDependencyEdgePointer{kind: typeDependencyEdgeKindStruct, pos: option.Some(i)}))
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (parent *typeNode) drawEdge(
	stack []typeDependencyEdgePointer,
	typeArgs []typeArg,
	childTy *types.Named,
	child *typeNode,
) {
	if parent.children == nil {
		parent.children = make(map[typeIdent][]typeDependencyEdge)
	}
	if child.parent == nil {
		child.parent = make(map[typeIdent][]typeDependencyEdge)
	}

	edge := typeDependencyEdge{
		stack:      slices.Clone(stack),
		typeArgs:   typeArgs,
		parentNode: parent,
		childType:  childTy,
		childNode:  child,
	}

	parentIdent := typeIdentFromTypesObject(parent.typeInfo.Obj())
	child.parent[parentIdent] = append(child.parent[parentIdent], edge)

	childIdent := typeIdentFromTypesObject(child.typeInfo.Obj())
	parent.children[childIdent] = append(parent.children[childIdent], edge)
}

func (g *typeGraph) markTransitive(edgeFilter func(edge typeDependencyEdge) bool) {
	for _, node := range g.types {
		node.matched = node.matched &^ typeNodeMatchKindTransitive
	}
	for _, node := range g.iterUpward(false, edgeFilter) {
		if node.matched.IsExternal() || node.matched.IsMatched() {
			continue
		}
		node.matched |= typeNodeMatchKindTransitive
	}
}

func (g *typeGraph) iterUpward(includeMatched bool, edgeFilter func(edge typeDependencyEdge) bool) iter.Seq2[typeIdent, *typeNode] {
	return func(yield func(typeIdent, *typeNode) bool) {
		// record visited nodes to break cyclic link.
		visited := make(map[*typeNode]bool)
		for _, n := range g.external {
			for ii, nn := range visitUpward(n, edgeFilter, visited) {
				if visited[n] {
					continue
				}
				visited[n] = true
				if !yield(ii, nn) {
					return
				}
			}
		}
		for i, n := range g.matched {
			if includeMatched && !visited[n] {
				visited[n] = true
				if !yield(i, n) {
					return
				}
			}
			for ii, nn := range visitUpward(n, edgeFilter, visited) {
				if visited[n] {
					continue
				}
				visited[n] = true
				if !yield(ii, nn) {
					return
				}
			}
		}
	}
}

func visitUpward(
	from *typeNode,
	edgeFilter func(edge typeDependencyEdge) bool,
	visited map[*typeNode]bool,
) iter.Seq2[typeIdent, *typeNode] {
	return visitNodes(from, true, edgeFilter, visited)
}

func visitNodes(
	n *typeNode,
	up bool,
	edgeFilter func(edge typeDependencyEdge) bool,
	visited map[*typeNode]bool,
) iter.Seq2[typeIdent, *typeNode] {
	return func(yield func(typeIdent, *typeNode) bool) {
		direction := n.parent
		if !up {
			direction = n.children
		}
		for i, v := range direction {
			for _, edge := range v {
				node := edge.parentNode
				if !up {
					node = edge.childNode
				}

				if (edgeFilter == nil || edgeFilter(edge)) && !yield(i, node) {
					return
				}

				for i, n := range visitNodes(node, up, edgeFilter, visited) {
					if visited[n] {
						continue
					}
					if !yield(i, n) {
						return
					}
				}
			}
		}
	}
}

// fields enumerates its children edges as iter.Seq2[int, typeDependencyEdge] assuming n's underlying type is struct.
// The key of the iterator is position of field in source code order.
func (n *typeNode) fields() iter.Seq2[int, typeDependencyEdge] {
	_ = n.typeInfo.Underlying().(*types.Struct)
	return func(yield func(int, typeDependencyEdge) bool) {
		for _, edges := range n.children {
			for _, e := range edges {
				if !yield(e.stack[0].pos.Value(), e) {
					return
				}
			}
		}
	}
}

func (n *typeNode) fieldsName() iter.Seq2[string, typeDependencyEdge] {
	structTy := n.typeInfo.Underlying().(*types.Struct)
	return func(yield func(string, typeDependencyEdge) bool) {
		for _, edges := range n.children {
			for _, e := range edges {
				if !yield(structTy.Field(e.stack[0].pos.Value()).Name(), e) {
					return
				}
			}
		}
	}
}

func (n *typeNode) byFieldName(name string) (typeDependencyEdge, *types.Var, reflect.StructTag, bool) {
	structObj := n.typeInfo.Underlying().(*types.Struct)
	for _, edges := range n.children {
		for _, e := range edges {
			if e.stack[0].pos.IsNone() {
				return typeDependencyEdge{}, nil, "", false
			}
			pos := e.stack[0].pos.Value()
			v := structObj.Field(pos)
			if v.Name() == name {
				return e, v, reflect.StructTag(structObj.Tag(pos)), true
			}
		}
	}
	return typeDependencyEdge{}, nil, "", false
}

func (g *typeGraph) enumerateTypesKeys(keys iter.Seq[typeIdent]) iter.Seq2[typeIdent, *typeNode] {
	return hiter.MapKeys(g.types, keys)
}
