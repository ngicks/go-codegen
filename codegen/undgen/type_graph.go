package undgen

import (
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"slices"

	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-iterator-helper/hiter"
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
type typeGraph struct {
	types    map[typeIdent]*typeNode
	matched  map[typeIdent]*typeNode
	external map[typeIdent]*typeNode
}

type typeIdent struct {
	pkgPath  string
	typeName string
}

func typeIdentFromTypesObject(obj types.Object) typeIdent {
	return typeIdent{
		obj.Pkg().Path(),
		obj.Name(),
	}
}

type typeNode struct {
	parent   map[typeIdent]typeDependencyEdge
	children map[typeIdent]typeDependencyEdge

	matched typeNodeMatchKind

	pkg  *packages.Package
	file *ast.File
	// nth type spec in the file.
	pos      int
	ts       *ast.TypeSpec
	typeInfo types.Object
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
	stack []typeDependencyEdgePointer
	node  *typeNode
}

type typeDependencyEdgePointer struct {
	kind typeDependencyEdgeKind
	pos  int
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
	typeDependencyEdgeKindTypeParam
)

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
	for pkg, fileSeq := range pkgsutil.EnumeratePackages(pkgs) {
		if err := pkgsutil.LoadError(pkg); err != nil {
			return err
		}
		for file := range fileSeq {
			var pos int
			for _, dec := range file.Decls {
				genDecl, ok := dec.(*ast.GenDecl)
				if !ok {
					continue
				}
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

					node := addType(g.types, pkg, file, currentPos, ts, obj)
					ok, err := matcher(named, false)
					if err != nil {
						return err
					}
					if ok {
						node.matched |= typeNodeMatchKindMatched
						g.matched[typeIdentFromTypesObject(node.typeInfo)] = node
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
	typeInfo types.Object,
) *typeNode {
	ident := typeIdentFromTypesObject(typeInfo)
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
			node.typeInfo.Type().Underlying(),
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
				parentNode.drawEdge(stack, node)
				if node.matched.IsMatched() {
					for i, arg := range hiter.AtterAll(named.TypeArgs()) {
						err := visitTypes(
							parentNode,
							arg,
							matcher,
							allType,
							externalType,
							append(
								stack,
								typeDependencyEdgePointer{
									kind: typeDependencyEdgeKindTypeParam,
									pos:  i,
								},
							),
						)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}
			ok, err := matcher(named, true)
			if !ok || err != nil {
				return err
			}
			externalNode := addType(externalType, nil, nil, -1, nil, named.Obj())
			externalNode.matched |= typeNodeMatchKindExternal
			parentNode.drawEdge(stack, externalNode)
			return nil
		},
		stack,
	)
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
			err := visitToNamed(f.Type(), cb, append(stack, typeDependencyEdgePointer{kind: typeDependencyEdgeKindStruct}))
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (parent *typeNode) drawEdge(stack []typeDependencyEdgePointer, children *typeNode) {
	if parent.children == nil {
		parent.children = make(map[typeIdent]typeDependencyEdge)
	}
	if children.parent == nil {
		children.parent = make(map[typeIdent]typeDependencyEdge)
	}
	stack = slices.Clone(stack)
	parent.children[typeIdentFromTypesObject(children.typeInfo)] = typeDependencyEdge{
		stack: stack,
		node:  children,
	}
	children.parent[typeIdentFromTypesObject(parent.typeInfo)] = typeDependencyEdge{
		stack: stack,
		node:  parent,
	}
}

func (g *typeGraph) markTransitive(edgeFilter func(p []typeDependencyEdgePointer) bool) {
	for _, node := range g.types {
		node.matched = node.matched &^ typeNodeMatchKindTransitive
	}
	for node := range g.iterUpward(false, edgeFilter) {
		if node.matched.IsExternal() || node.matched.IsMatched() {
			continue
		}
		node.matched |= typeNodeMatchKindTransitive
	}
}

func (g *typeGraph) iterUpward(includeMatched bool, edgeFilter func(p []typeDependencyEdgePointer) bool) iter.Seq[*typeNode] {
	return func(yield func(*typeNode) bool) {
		visited := make(map[*typeNode]bool)
		for _, n := range g.external {
			for nn := range visitUpward(n, edgeFilter, visited) {
				if !yield(nn) {
					return
				}
			}
		}
		for _, n := range g.matched {
			if includeMatched {
				if !yield(n) {
					return
				}
			}
			for nn := range visitUpward(n, edgeFilter, visited) {
				if !yield(nn) {
					return
				}
			}
		}
	}
}

func visitUpward(
	from *typeNode,
	edgeFilter func(p []typeDependencyEdgePointer) bool,
	visited map[*typeNode]bool,
) iter.Seq[*typeNode] {
	return visitNodes(from, true, edgeFilter, visited)
}

func visitNodes(
	n *typeNode,
	up bool,
	edgeFilter func(p []typeDependencyEdgePointer) bool,
	visited map[*typeNode]bool,
) iter.Seq[*typeNode] {
	return func(yield func(*typeNode) bool) {
		direction := n.parent
		if !up {
			direction = n.children
		}
		for _, v := range direction {
			if visited[v.node] {
				continue
			}
			visited[v.node] = true
			if (edgeFilter == nil || edgeFilter(v.stack)) && !yield(v.node) {
				return
			}
			for n := range visitNodes(v.node, up, edgeFilter, visited) {
				if !yield(n) {
					return
				}
			}
		}
	}
}
