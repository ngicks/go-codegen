// typegraph forms a graph by drawing edges between a named type and other named type.
package typegraph

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"maps"
	"reflect"
	"slices"

	"github.com/ngicks/go-codegen/codegen/codegen"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"github.com/ngicks/und/option"
	"golang.org/x/tools/go/packages"
)

// Graph enumerates type decls in given []*packages.Package and forms a type-dependency graph.
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
type Graph struct {
	types      map[Ident]*Node
	matched    map[Ident]*Node
	external   map[Ident]*Node
	privParser PrivParser
}

type Ident struct {
	PkgPath  string
	TypeName string
}

func (t Ident) TargetType() imports.TargetType {
	return imports.TargetType{ImportPath: t.PkgPath, Name: t.TypeName}
}

func IdentFromTypesObject(obj types.Object) Ident {
	var pkgsPath string
	if obj.Pkg() != nil {
		pkgsPath = obj.Pkg().Path()
	}
	return Ident{
		pkgsPath,
		obj.Name(),
	}
}

type Node struct {
	Parent   map[Ident][]Edge
	Children map[Ident][]Edge

	Matched MatchKind

	Pkg  *packages.Package
	File *ast.File
	// nth type spec in the file.
	Pos  int
	Ts   *ast.TypeSpec
	Type *types.Named

	Priv any
}

type MatchKind uint64

const (
	MatchKindMatched = MatchKind(1 << iota)
	MatchKindDependant
	MatchKindExternal
)

func (k MatchKind) IsMatched() bool {
	return k&MatchKindMatched > 0
}

func (k MatchKind) IsDependant() bool {
	return k&MatchKindDependant > 0
}

func (k MatchKind) IsExternal() bool {
	return k&MatchKindExternal > 0
}

type Edge struct {
	Stack    []EdgeRouteNode
	TypeArgs []TypeArg
	// non-instantiated parent
	ParentNode *Node
	// instantiated child
	ChildType *types.Named
	// non-instantiated child node.
	ChildNode *Node
}

func (e Edge) IsChildMatched() bool {
	return e.ChildNode.Matched&^MatchKindExternal > 0
}

func (e Edge) HasSingleNamedTypeArg(additionalCond func(named *types.Named) bool) (ok bool, pointer bool) {
	if len(e.TypeArgs) != 1 {
		return false, false
	}
	arg := e.TypeArgs[0]
	isPointer := (len(arg.Route) == 1 && arg.Route[0].Kind == EdgeKindPointer)
	if arg.Ty == nil || len(arg.Route) > 1 || (len(arg.Route) != 0 && !isPointer) {
		return false, isPointer
	}
	if additionalCond != nil {
		return additionalCond(arg.Ty), isPointer
	}
	return true, isPointer
}

func (e Edge) IsTypeArgMatched() bool {
	if len(e.TypeArgs) == 0 {
		return false
	}
	node := e.TypeArgs[0].Node
	if node == nil {
		return false
	}
	return node.Matched&^MatchKindExternal > 0
}

func (e Edge) LastPointer() option.Option[EdgeRouteNode] {
	return option.GetSlice(e.Stack, len(e.Stack)-1)
}

func (e Edge) PrintChildType(importMap imports.ImportMap) string {
	return codegen.PrintAstExprPanicking(
		codegen.TypeToAst(
			e.ChildType,
			e.ParentNode.Type.Obj().Pkg().Path(),
			importMap,
		),
	)
}

func (e Edge) PrintChildArg(i int, importMap imports.ImportMap) string {
	return codegen.PrintAstExprPanicking(
		codegen.TypeToAst(
			e.TypeArgs[i].Org,
			e.ParentNode.Type.Obj().Pkg().Path(),
			importMap,
		),
	)
}

func (e Edge) PrintChildArgConverted(converter func(ty *types.Named, isMatched bool) (*types.Named, bool), importMap imports.ImportMap) string {
	isMatched := false
	isConverter := func(named *types.Named) bool {
		if node := e.TypeArgs[0].Node; node != nil {
			isMatched = (node.Matched &^ MatchKindExternal) > 0
		}
		_, ok := converter(named, isMatched)
		return ok
	}

	var plainParam types.Type
	if ok, isPointer := e.HasSingleNamedTypeArg(isConverter); ok {
		converted, _ := converter(e.TypeArgs[0].Ty, isMatched)
		plainParam = converted
		if isPointer {
			plainParam = types.NewPointer(plainParam)
		}
	} else {
		plainParam = e.TypeArgs[0].Org
	}

	return codegen.PrintAstExprPanicking(codegen.TypeToAst(
		plainParam,
		e.ParentNode.Type.Obj().Pkg().Path(),
		importMap,
	))
}

type EdgeRouteNode struct {
	Kind EdgeKind
	Pos  option.Option[int]
}

type TypeArg struct {
	Route []EdgeRouteNode
	Node  *Node
	Ty    *types.Named
	Org   types.Type
}

type EdgeKind uint64

const (
	EdgeKindAlias = EdgeKind(1 << iota)
	EdgeKindArray
	EdgeKindChan
	EdgeKindInterface
	EdgeKindMap
	EdgeKindNamed
	EdgeKindPointer
	EdgeKindSlice
	EdgeKindStruct
)

func FirstTypeIdent(m map[Ident][]Edge) (Ident, Edge) {
	for k, e := range m {
		return k, e[0]
	}
	return Ident{}, Edge{}
}

func New(
	pkgs []*packages.Package,
	matcher func(node *Node, external bool) (bool, error),
	genDeclFilter func(*ast.GenDecl) (bool, error),
	typeSpecFilter func(*ast.TypeSpec, types.Object) (bool, error),
	opts ...Option,
) (*Graph, error) {
	graph := &Graph{
		types:    make(map[Ident]*Node),
		matched:  make(map[Ident]*Node),
		external: make(map[Ident]*Node),
	}

	for _, opt := range opts {
		opt.apply(graph)
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

func (g *Graph) listTypes(
	pkgs []*packages.Package,
	matcher func(node *Node, external bool) (bool, error),
	genDeclFilter func(*ast.GenDecl) (bool, error),
	typeSpecFilter func(*ast.TypeSpec, types.Object) (bool, error),
) error {
	parser := g.privParser
	if parser == nil {
		parser = func(n *Node) (any, error) { return nil, nil }
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
					var err error
					node.Priv, err = parser(node)
					if err != nil {
						return fmt.Errorf("parsing priv: %w", err)
					}
					ok, err = matcher(node, false)
					if err != nil {
						return fmt.Errorf("matching type: %w", err)
					}
					if ok {
						node.Matched |= MatchKindMatched
						g.matched[IdentFromTypesObject(node.Type.Obj())] = node
					}
				}
			}
		}
	}
	return nil
}

func addType(
	to map[Ident]*Node,
	pkg *packages.Package,
	file *ast.File,
	pos int,
	ts *ast.TypeSpec,
	typeInfo *types.Named,
) *Node {
	ident := IdentFromTypesObject(typeInfo.Obj())
	n, ok := to[ident]
	if ok {
		return n
	}
	n = &Node{
		Pkg:  pkg,
		File: file,
		Pos:  pos,
		Ts:   ts,
		Type: typeInfo,
	}
	to[ident] = n
	return n
}

func (g *Graph) buildEdge(
	matcher func(node *Node, external bool) (bool, error),
) error {
	for _, node := range g.types {
		// Underlying matches what of go spec.
		// It means what follows type idents like below:
		//
		// type Foo struct {Foo string; Bar int}
		//          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^ this part is underlying
		err := visitTypes(
			node,
			node.Type.Underlying(),
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
	parentNode *Node,
	ty types.Type,
	matcher func(node *Node, external bool) (bool, error),
	allType map[Ident]*Node,
	externalType map[Ident]*Node,
	stack []EdgeRouteNode,
) error {
	return TraverseToNamed(
		ty,
		func(named *types.Named, stack []EdgeRouteNode) error {
			node, ok := allType[IdentFromTypesObject(named.Obj())]
			if ok {
				parentNode.drawEdge(
					stack,
					visitOnTypeArgs(
						named.TypeArgs(),
						matcher,
						allType,
						externalType,
					),
					named,
					node,
				)
				return nil
			}
			ok, err := matcher(
				&Node{
					Pos:  -1,
					Type: named,
				},
				true,
			)
			if !ok || err != nil {
				return err
			}
			externalNode := addType(externalType, nil, nil, -1, nil, named)
			externalNode.Matched |= MatchKindExternal
			parentNode.drawEdge(
				stack,
				visitOnTypeArgs(
					named.TypeArgs(),
					matcher,
					allType,
					externalType,
				),
				named,
				externalNode,
			)
			return nil
		},
		stack,
	)
}

func visitOnTypeArgs(
	typeList *types.TypeList,
	matcher func(node *Node, external bool) (bool, error),
	allType map[Ident]*Node,
	externalType map[Ident]*Node,
) []TypeArg {
	var typeArgs []TypeArg
	for _, arg := range hiter.AtterAll(typeList) {
		var found bool
		_ = TraverseToNamed(
			arg,
			func(named *types.Named, stack []EdgeRouteNode) error {
				found = true
				// TODO: split this `check if internal types, if not, then add as external type` sequence as a function or a method.
				node, ok := allType[IdentFromTypesObject(named.Obj())]
				if !ok {
					ok, err := matcher(&Node{Pos: -1, Type: named}, true)
					if ok && err == nil {
						externalNode := addType(externalType, nil, nil, -1, nil, named)
						externalNode.Matched |= MatchKindExternal
						node = externalNode
					}
				}
				typeArgs = append(typeArgs, TypeArg{
					Route: slices.Clone(stack),
					Node:  node, // might still be nil.
					Ty:    named,
					Org:   arg,
				})
				return nil
			},
			nil,
		)
		if !found {
			typeArgs = append(typeArgs, TypeArg{
				Org: arg,
			})
		}
	}
	return typeArgs
}

func TraverseToNamed(
	ty types.Type,
	cb func(named *types.Named, stack []EdgeRouteNode) error,
	stack []EdgeRouteNode,
) error {
	return TraverseTypes(
		ty,
		nil,
		func(ty types.Type, named *types.Named, stack []EdgeRouteNode) error {
			if named == nil {
				return nil
			}
			return cb(named, stack)
		},
		stack,
	)
}

func TraverseTypes(
	ty types.Type,
	stopper func(ty types.Type, currentStack []EdgeRouteNode) bool,
	cb func(ty types.Type, named *types.Named, stack []EdgeRouteNode) error,
	stack []EdgeRouteNode,
) error {
	if stopper != nil && stopper(ty, stack) {
		named, _ := ty.(*types.Named)
		return cb(ty, named, stack)
	}
	// types may recurse.
	// but should be impossible without naming type,
	// which breaks visitToNamed from infinite loop.
	switch x := ty.(type) {
	default:
		return cb(x, nil, stack)
	case *types.Alias:
		// TODO: check for type param after go1.24
		// see https://github.com/golang/go/issues/46477
		return TraverseTypes(x.Rhs(), stopper, cb, append(stack, EdgeRouteNode{Kind: EdgeKindAlias}))
	case *types.Array:
		return TraverseTypes(x.Elem(), stopper, cb, append(stack, EdgeRouteNode{Kind: EdgeKindArray}))
	case *types.Chan:
		return TraverseTypes(x.Elem(), stopper, cb, append(stack, EdgeRouteNode{Kind: EdgeKindChan}))
	case *types.Map:
		return TraverseTypes(x.Elem(), stopper, cb, append(stack, EdgeRouteNode{Kind: EdgeKindMap}))
	case *types.Named:
		return cb(x, x, stack)
	case *types.Pointer:
		return TraverseTypes(x.Elem(), stopper, cb, append(stack, EdgeRouteNode{Kind: EdgeKindPointer}))
	case *types.Slice:
		return TraverseTypes(x.Elem(), stopper, cb, append(stack, EdgeRouteNode{Kind: EdgeKindSlice}))
	case *types.Struct:
		// We don't support type-parametrized struct fields.
		// Thus not checking type args.
		for i := range x.NumFields() {
			f := x.Field(i)
			err := TraverseTypes(
				f.Type(),
				stopper,
				cb,
				append(stack, EdgeRouteNode{Kind: EdgeKindStruct, Pos: option.Some(i)}),
			)
			if err != nil {
				return err
			}
		}
		return nil
	case *types.TypeParam:
		return cb(x, nil, stack)
	}
}

func (parent *Node) drawEdge(
	stack []EdgeRouteNode,
	typeArgs []TypeArg,
	childTy *types.Named,
	child *Node,
) {
	if parent.Children == nil {
		parent.Children = make(map[Ident][]Edge)
	}
	if child.Parent == nil {
		child.Parent = make(map[Ident][]Edge)
	}

	edge := Edge{
		Stack:      slices.Clone(stack),
		TypeArgs:   typeArgs,
		ParentNode: parent,
		ChildType:  childTy,
		ChildNode:  child,
	}

	parentIdent := IdentFromTypesObject(parent.Type.Obj())
	child.Parent[parentIdent] = append(child.Parent[parentIdent], edge)

	childIdent := IdentFromTypesObject(child.Type.Obj())
	parent.Children[childIdent] = append(parent.Children[childIdent], edge)
}

func (g *Graph) MarkDependant(edgeFilter func(edge Edge) bool) {
	for _, node := range g.types {
		node.Matched = node.Matched &^ MatchKindDependant
	}
	for _, node := range g.IterUpward(false, edgeFilter) {
		if node.Matched.IsExternal() || node.Matched.IsMatched() {
			continue
		}
		node.Matched |= MatchKindDependant
	}
}

func (g *Graph) IterUpward(includeMatched bool, edgeFilter func(edge Edge) bool) iter.Seq2[Ident, *Node] {
	return func(yield func(Ident, *Node) bool) {
		// record visited nodes to break cyclic link.
		visited := make(map[*Node]bool)
		for _, n := range g.external {
			for ii, nn := range visitUpward(n, edgeFilter, visited) {
				if !yield(ii, nn) {
					return
				}
			}
		}
		for i, n := range g.matched {
			if visited[n] {
				continue
			}

			if includeMatched && !visited[n] {
				if !yield(i, n) {
					return
				}
			}

			visited[n] = true
			for ii, nn := range visitUpward(n, edgeFilter, visited) {
				if !yield(ii, nn) {
					return
				}
			}
		}
	}
}

func visitUpward(
	n *Node,
	edgeFilter func(edge Edge) bool,
	visited map[*Node]bool,
) iter.Seq2[Ident, *Node] {
	return func(yield func(Ident, *Node) bool) {
		for i, v := range n.Parent {
			for _, edge := range v {
				node := edge.ParentNode

				if visited[node] {
					continue
				}
				if edgeFilter == nil || edgeFilter(edge) {
					if !yield(i, node) {
						return
					}
				} else {
					continue
				}

				visited[node] = true
				for i, n := range visitUpward(node, edgeFilter, visited) {
					if !yield(i, n) {
						return
					}
				}
			}
		}
	}
}

// Fields enumerates its children edges as iter.Seq2[int, typeDependencyEdge] assuming n's underlying type is struct.
// The key of the iterator is position of field in source code order.
func (n *Node) Fields() iter.Seq2[int, Edge] {
	_ = n.Type.Underlying().(*types.Struct)
	return func(yield func(int, Edge) bool) {
		for _, edges := range n.Children {
			for _, e := range edges {
				if !yield(e.Stack[0].Pos.Value(), e) {
					return
				}
			}
		}
	}
}

func (n *Node) FieldsName() iter.Seq2[string, Edge] {
	structTy := n.Type.Underlying().(*types.Struct)
	return func(yield func(string, Edge) bool) {
		for _, edges := range n.Children {
			for _, e := range edges {
				if !yield(structTy.Field(e.Stack[0].Pos.Value()).Name(), e) {
					return
				}
			}
		}
	}
}

func (n *Node) ByFieldName(name string) (Edge, *types.Var, reflect.StructTag, bool) {
	structObj := n.Type.Underlying().(*types.Struct)
	for _, edges := range n.Children {
		for _, e := range edges {
			if e.Stack[0].Pos.IsNone() {
				return Edge{}, nil, "", false
			}
			pos := e.Stack[0].Pos.Value()
			v := structObj.Field(pos)
			if v.Name() == name {
				return e, v, reflect.StructTag(structObj.Tag(pos)), true
			}
		}
	}
	return Edge{}, nil, "", false
}

func (g *Graph) EnumerateTypes() iter.Seq2[Ident, *Node] {
	keys := slices.SortedFunc(maps.Keys(g.types), func(i, j Ident) int {
		if c := cmp.Compare(i.PkgPath, j.PkgPath); c != 0 {
			return c
		}
		return cmp.Compare(g.types[i].Pos, g.types[j].Pos)
	})
	return hiter.MapsKeys(g.types, slices.Values(keys))
}

func (g *Graph) Get(i Ident) (*Node, bool) {
	n, ok := g.types[i]
	return n, ok
}

func (g *Graph) GetByType(ty types.Type) (*Node, bool) {
	named, ok := ty.(*types.Named)
	if !ok {
		return nil, false
	}
	if named.Obj().Pkg() == nil {
		// error built-in interface
		return nil, false
	}
	return g.Get(IdentFromTypesObject(named.Obj()))
}

func (g *Graph) EnumerateTypesKeys(keys iter.Seq[Ident]) iter.Seq2[Ident, *Node] {
	return hiter.MapsKeys(g.types, keys)
}

type EdgeMap struct {
	node    *Node
	edgeMap map[Ident][]Edge
	posMap  map[int]Edge
	nameMap map[string]Edge
}

func (n *Node) ChildEdgeMap(edgeFilter func(edge Edge) bool) EdgeMap {
	if edgeFilter == nil {
		edgeFilter = func(edge Edge) bool { return true }
	}

	st, isStruct := n.Type.Underlying().(*types.Struct)
	var (
		posMap  map[int]Edge
		nameMap map[string]Edge
	)
	if isStruct {
		posMap = make(map[int]Edge)
		nameMap = make(map[string]Edge)
	}

	edgeMap := maps.Collect(
		xiter.Filter2(
			func(_ Ident, edges []Edge) bool {
				if isStruct {
					for pos, edge := range xiter.Map2(
						func(_ int, edge Edge) (int, Edge) {
							return edge.Stack[0].Pos.Value(), edge
						},
						slices.All(edges),
					) {
						posMap[pos] = edge
						nameMap[st.Field(pos).Name()] = edge
					}
				}
				return len(edges) > 0
			},
			xiter.Map2(
				func(i Ident, edges []Edge) (Ident, []Edge) {
					return i, slices.Collect(xiter.Filter(edgeFilter, slices.Values(edges)))
				},
				maps.All(n.Children),
			),
		),
	)

	return EdgeMap{
		node:    n,
		edgeMap: edgeMap,
		posMap:  posMap,
		nameMap: nameMap,
	}
}

func (em EdgeMap) First() (Ident, Edge, bool) {
	for k, v := range em.edgeMap {
		return k, v[0], true
	}
	return Ident{}, Edge{}, false
}

// Fields enumerates its children edges as iter.Seq2[int, typeDependencyEdge] assuming node's underlying type is struct.
// The key of the iterator is position of field in source code order.
func (em EdgeMap) Fields() iter.Seq2[int, Edge] {
	_ = em.node.Type.Underlying().(*types.Struct) // panic if not a struct.
	return func(yield func(int, Edge) bool) {
		for _, edges := range em.edgeMap {
			for _, e := range edges {
				if !yield(e.Stack[0].Pos.Value(), e) {
					return
				}
			}
		}
	}
}

// FieldsName is like [EdgeMap.Fields] but the key of the pair is field name.
func (em EdgeMap) FieldsName() iter.Seq2[string, Edge] {
	structTy := em.node.Type.Underlying().(*types.Struct) // panic if not
	return xiter.Map2(
		func(i int, edge Edge) (string, Edge) {
			return structTy.Field(i).Name(), edge
		},
		em.Fields(),
	)
}

// ByFieldPos returns the edge, the field var and the struct tag for the field positioned at pos in source code order,
// It assumes node's underlying is struct type, otherwise panics.
func (em EdgeMap) ByFieldPos(pos int) (Edge, *types.Var, reflect.StructTag, bool) {
	st := em.node.Type.Underlying().(*types.Struct) // panic if not
	edge, ok := em.posMap[pos]
	if !ok {
		return Edge{}, nil, "", false
	}
	return edge, st.Field(pos), reflect.StructTag(st.Tag(pos)), true
}

// ByFieldName is like [EdgeMap.ByFieldPos] but queries for fieldName.
func (em EdgeMap) ByFieldName(fieldName string) (Edge, *types.Var, reflect.StructTag, bool) {
	st := em.node.Type.Underlying().(*types.Struct) // panic if not
	edge, ok := em.nameMap[fieldName]
	if !ok {
		return Edge{}, nil, "", false
	}
	pos := edge.Stack[0].Pos.Value()
	return edge, st.Field(pos), reflect.StructTag(st.Tag(pos)), true
}
