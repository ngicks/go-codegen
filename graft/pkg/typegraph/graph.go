// Package typegraph analyzes Go modules and forms dependency graph of types.
package typegraph

import (
	"cmp"
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"maps"
	"slices"
	"unsafe"

	"github.com/ngicks/go-codegen/graft/internal/pkgsutil"
	"github.com/ngicks/go-iterator-helper/hiter"
	"golang.org/x/tools/go/packages"
)

type QualifiedType struct {
	PackagePath string
	TypeName    string
}

func NewQualifiedType(pkgPath, typeName string) QualifiedType {
	return QualifiedType{pkgPath, typeName}
}

// QualifiedTypeFromObject creates a QualifiedType from a types.Object.
func QualifiedTypeFromObject(obj types.Object) QualifiedType {
	var pkgPath string
	if obj.Pkg() != nil {
		pkgPath = obj.Pkg().Path()
	}
	return QualifiedType{
		PackagePath: pkgPath,
		TypeName:    obj.Name(),
	}
}

// Edge is bi-directional linkage between 2 nodes.
//
// Multiple linkages may be coalesced into single Edge;
// multiple fields of struct, key-value of map[K]V and/or type parameters may use target type.
type Edge[T Node | MatchedNode] struct {
	Parent *T
	// non-instantiated child node
	Child  *T
	Routes []*EdgeRoute
}

type TypeDetail struct {
	Pkg  *packages.Package
	File *ast.File
	// nth type spec in the file.
	Pos  int
	Ts   *ast.TypeSpec
	Type *types.Named
}

type Node struct {
	Id       QualifiedType
	Detail   *TypeDetail
	Parent   map[QualifiedType]*Edge[Node]
	Children map[QualifiedType]*Edge[Node]
}

func (n *Node) drawEdge(
	child *Node,
	routes ...*EdgeRoute,
) {
	newEdge := func() *Edge[Node] {
		return &Edge[Node]{
			Parent: n,
			Child:  child,
		}
	}
	parentIdent := QualifiedTypeFromObject(n.Detail.Type.Obj())
	e, ok := child.Parent[parentIdent]
	if !ok {
		e = newEdge()
		child.Parent[parentIdent] = e
		childIdent := QualifiedTypeFromObject(child.Detail.Type.Obj())
		n.Children[childIdent] = e
	}

	e.Routes = append(e.Routes, routes...)
}

type Graph struct {
	// all named types, excluding type alias
	all map[QualifiedType]*Node
}

// New creates a new Graph from the given packages.
// It enumerates all type declarations and builds edges between types.
//
// genDeclFilter filters which GenDecl to process (nil means all).
// typeSpecFilter filters which TypeSpec to include (nil means all).
func New(
	pkgs []*packages.Package,
	genDeclFilter func(*ast.GenDecl) (bool, error),
	typeSpecFilter func(*ast.TypeSpec, types.Object) (bool, error),
) (*Graph, error) {
	g := &Graph{
		all: make(map[QualifiedType]*Node),
	}

	if err := listTypes(g, pkgs, genDeclFilter, typeSpecFilter); err != nil {
		return nil, err
	}

	buildEdges(g)

	return g, nil
}

func listTypes(
	g *Graph,
	pkgs []*packages.Package,
	genDeclFilter func(*ast.GenDecl) (bool, error),
	typeSpecFilter func(*ast.TypeSpec, types.Object) (bool, error),
) error {
	if genDeclFilter == nil {
		genDeclFilter = func(gd *ast.GenDecl) (bool, error) { return true, nil }
	}
	if typeSpecFilter == nil {
		typeSpecFilter = func(ts *ast.TypeSpec, o types.Object) (bool, error) { return true, nil }
	}

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
				ok, err := genDeclFilter(genDecl)
				if err != nil {
					return err
				}
				if !ok {
					continue
				}
				for _, s := range genDecl.Specs {
					currentPos := pos
					pos++
					ts := s.(*ast.TypeSpec)
					obj := pkg.TypesInfo.Defs[ts.Name]

					ok, err = typeSpecFilter(ts, obj)
					if err != nil {
						return err
					}
					if !ok {
						continue
					}

					named, ok := obj.Type().(*types.Named)
					if !ok {
						continue
					}

					_ = addNode(
						g.all,
						named,
						func() *Node {
							return &Node{
								Id:       QualifiedTypeFromObject(named.Obj()),
								Detail:   &TypeDetail{pkg, file, currentPos, ts, named},
								Parent:   map[QualifiedType]*Edge[Node]{},
								Children: map[QualifiedType]*Edge[Node]{},
							}
						},
					)
				}
			}
		}
	}
	return nil
}

func addNode[T Node | MatchedNode](
	to map[QualifiedType]*T,
	named *types.Named,
	createNode func() *T,
) *T {
	ident := QualifiedTypeFromObject(named.Obj())
	n, ok := to[ident]
	if ok {
		return n
	}
	n = createNode()
	to[ident] = n
	return n
}

func buildEdges(g *Graph) error {
	for _, node := range g.all {
		// Underlying matches what of go spec.
		// It means what follows type idents like below:
		//
		// type Foo struct {Foo string; Bar int}
		//          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^ this part is underlying
		err := visitTypes(
			g,
			node,
			node.Detail.Type.Underlying(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func visitTypes(
	g *Graph,
	parentNode *Node,
	ty types.Type,
) error {
	return WalkTypesToNamed(
		nil,
		ty,
		nil,
		func(named *types.Named, stack []EdgeRouteNode) error {
			node, ok := g.all[QualifiedTypeFromObject(named.Obj())]
			if ok {
				parentNode.drawEdge(node, &EdgeRoute{slices.Clone(stack)})
			}
			return nil
		},
	)
}

// Get returns the node for the given id.
func (g *Graph) Get(id QualifiedType) (*Node, bool) {
	n, ok := g.all[id]
	return n, ok
}

// All returns all nodes in the graph.
func (g *Graph) All() iter.Seq2[QualifiedType, *Node] {
	return maps.All(g.all)
}

// Trim creates a MatchedGraph by applying matchers to filter nodes and edges.
// nodeMatcher determines which nodes are "matched" (primary targets).
// edgeMatcher determines which edges to follow when marking dependants.
func (g *Graph) Trim(
	nodeMatcher func(n *Node) bool,
	edgeMatcher func(e *Edge[Node]) bool,
	routeFilter func(r *EdgeRoute) bool,
) *MatchedGraph {
	mg := &MatchedGraph{
		origin:    g,
		matched:   make(map[QualifiedType]*MatchedNode),
		dependant: make(map[QualifiedType]*MatchedNode),
	}

	// First pass: identify matched nodes
	for id, node := range g.all {
		if !nodeMatcher(node) {
			continue
		}

		mn := &MatchedNode{
			MatchKind: MatchKindMatched,
			Parent:    make(map[QualifiedType]*Edge[MatchedNode]),
			Children:  make(map[QualifiedType]*Edge[MatchedNode]),
			Node:      node,
		}

		mg.matched[id] = mn
	}

	markDependants(mg, edgeMatcher, routeFilter)

	return mg
}

func markDependants(mg *MatchedGraph, edgeMatcher func(e *Edge[Node]) bool, routeFilter func(r *EdgeRoute) bool) {
	var visited map[QualifiedType]bool

	var visitUpward func(matchedNode *MatchedNode)
	visitUpward = func(matchedNode *MatchedNode) {
		node := matchedNode.Node
		for parentId, edge := range node.Parent {
			if visited[parentId] {
				continue
			}
			if !edgeMatcher(edge) {
				continue
			}

			filtered := edge.Routes
			if routeFilter != nil {
				filtered = slices.Collect(
					hiter.Filter(
						routeFilter,
						slices.Values(edge.Routes),
					),
				)
			}

			visited[parentId] = true

			n := addNode(
				mg.dependant,
				edge.Parent.Detail.Type,
				func() *MatchedNode {
					return &MatchedNode{
						Node:     edge.Parent,
						Parent:   map[QualifiedType]*Edge[MatchedNode]{},
						Children: map[QualifiedType]*Edge[MatchedNode]{},
					}
				},
			)
			n.MatchKind |= MatchKindDependant

			n.drawEdge(matchedNode, filtered...)

			visitUpward(n)
		}
	}

	for _, mn := range mg.matched {
		visited = make(map[QualifiedType]bool)
		visitUpward(mn)
	}

	edgeVisited := make(map[unsafe.Pointer]bool)
	for _, n := range mg.All() {
		n.cleanRoute(edgeVisited)
	}
}

type MatchKind uint64

const (
	MatchKindMatched = MatchKind(1 << iota)
	MatchKindDependant
)

func (k MatchKind) IsMatched() bool {
	return k&MatchKindMatched > 0
}

func (k MatchKind) IsDependant() bool {
	return k&MatchKindDependant > 0
}

type MatchedNode struct {
	MatchKind MatchKind
	Parent    map[QualifiedType]*Edge[MatchedNode]
	Children  map[QualifiedType]*Edge[MatchedNode]
	Node      *Node
}

func (n *MatchedNode) drawEdge(
	child *MatchedNode,
	routes ...*EdgeRoute,
) {
	newEdge := func() *Edge[MatchedNode] {
		return &Edge[MatchedNode]{
			Parent: n,
			Child:  child,
		}
	}

	parentIdent := QualifiedTypeFromObject(n.Node.Detail.Type.Obj())
	e, ok := child.Parent[parentIdent]
	if !ok {
		e = newEdge()
		child.Parent[parentIdent] = e
		childIdent := QualifiedTypeFromObject(child.Node.Detail.Type.Obj())
		n.Children[childIdent] = e
	}
	e.Routes = append(e.Routes, routes...)
}

func (n *MatchedNode) cleanRoute(visited map[unsafe.Pointer]bool) {
	for _, edge := range hiter.Concat2(maps.All(n.Children), maps.All(n.Parent)) {
		ptr := unsafe.Pointer(edge)
		if visited[ptr] {
			continue
		}
		visited[ptr] = true

		slices.SortFunc(
			edge.Routes,
			func(i, j *EdgeRoute) int {
				return cmp.Compare(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(j)))
			},
		)
		edge.Routes = slices.Compact(edge.Routes)
	}
}

type MatchedGraph struct {
	origin *Graph
	// matched types.
	matched map[QualifiedType]*MatchedNode
	// dependant to types listed in matched or external.
	dependant map[QualifiedType]*MatchedNode
}

func (g *MatchedGraph) Get(id QualifiedType) (*MatchedNode, bool) {
	var (
		n  *MatchedNode
		ok bool
	)
	n, ok = g.matched[id]
	if ok {
		return n, true
	}
	n, ok = g.dependant[id]
	if ok {
		return n, true
	}
	return nil, false
}

func (g *MatchedGraph) All() iter.Seq2[QualifiedType, *MatchedNode] {
	return func(yield func(QualifiedType, *MatchedNode) bool) {
		for k, v := range g.matched {
			if !yield(k, v) {
				return
			}
		}
		for k, v := range g.dependant {
			if _, visited := g.matched[k]; visited {
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

func (g *MatchedGraph) Matched() iter.Seq2[QualifiedType, *MatchedNode] {
	return maps.All(g.matched)
}

func (g *MatchedGraph) Dependant() iter.Seq2[QualifiedType, *MatchedNode] {
	return maps.All(g.dependant)
}

func (g *MatchedGraph) Origin() *Graph {
	return g.origin
}
