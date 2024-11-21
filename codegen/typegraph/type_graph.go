// typegraph forms a graph by drawing edges between a named type and other named type.
package typegraph

import (
	"cmp"
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

// TypeGraph enumerates type decls in given []*packages.Package and forms a type-dependency graph.
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
type TypeGraph struct {
	types    map[TypeIdent]*TypeNode
	matched  map[TypeIdent]*TypeNode
	external map[TypeIdent]*TypeNode
}

type TypeIdent struct {
	PkgPath  string
	TypeName string
}

func (t TypeIdent) TargetType() imports.TargetType {
	return imports.TargetType{ImportPath: t.PkgPath, Name: t.TypeName}
}

func IdentFromTypesObject(obj types.Object) TypeIdent {
	return TypeIdent{
		obj.Pkg().Path(),
		obj.Name(),
	}
}

type TypeNode struct {
	Parent   map[TypeIdent][]TypeDependencyEdge
	Children map[TypeIdent][]TypeDependencyEdge

	Matched TypeNodeMatchKind

	Pkg  *packages.Package
	File *ast.File
	// nth type spec in the file.
	Pos  int
	Ts   *ast.TypeSpec
	Type *types.Named
}

type TypeNodeMatchKind uint64

const (
	TypeNodeMatchKindMatched = TypeNodeMatchKind(1 << iota)
	TypeNodeMatchKindDependant
	TypeNodeMatchKindExternal
)

func (k TypeNodeMatchKind) IsMatched() bool {
	return k&TypeNodeMatchKindMatched > 0
}

func (k TypeNodeMatchKind) IsDependant() bool {
	return k&TypeNodeMatchKindDependant > 0
}

func (k TypeNodeMatchKind) IsExternal() bool {
	return k&TypeNodeMatchKindExternal > 0
}

type TypeDependencyEdge struct {
	Stack    []TypeDependencyEdgePointer
	TypeArgs []TypeArg
	// non-instantiated parent
	ParentNode *TypeNode
	// instantiated child
	ChildType *types.Named
	// non-instantiated child node.
	ChildNode *TypeNode
}

func (e TypeDependencyEdge) IsChildMatched() bool {
	return e.ChildNode.Matched&^TypeNodeMatchKindExternal > 0
}

func (e TypeDependencyEdge) HasSingleNamedTypeArg(additionalCond func(named *types.Named) bool) (ok bool, pointer bool) {
	if len(e.TypeArgs) != 1 {
		return false, false
	}
	arg := e.TypeArgs[0]
	isPointer := (len(arg.Stack) == 1 && arg.Stack[0].Kind == TypeDependencyEdgeKindPointer)
	if arg.Ty == nil || len(arg.Stack) > 1 || (len(arg.Stack) != 0 && !isPointer) {
		return false, isPointer
	}
	if additionalCond != nil {
		return additionalCond(arg.Ty), isPointer
	}
	return true, isPointer
}

func (e TypeDependencyEdge) IsTypeArgMatched() bool {
	if len(e.TypeArgs) == 0 {
		return false
	}
	node := e.TypeArgs[0].Node
	if node == nil {
		return false
	}
	return node.Matched&^TypeNodeMatchKindExternal > 0
}

func (e TypeDependencyEdge) LastPointer() option.Option[TypeDependencyEdgePointer] {
	return option.GetSlice(e.Stack, len(e.Stack)-1)
}

func (e TypeDependencyEdge) PrintChildType(importMap imports.ImportMap) string {
	return codegen.PrintAstExprPanicking(
		codegen.TypeToAst(
			e.ChildType,
			e.ParentNode.Type.Obj().Pkg().Path(),
			importMap,
		),
	)
}

func (e TypeDependencyEdge) PrintChildArg(i int, importMap imports.ImportMap) string {
	return codegen.PrintAstExprPanicking(
		codegen.TypeToAst(
			e.TypeArgs[i].Org,
			e.ParentNode.Type.Obj().Pkg().Path(),
			importMap,
		),
	)
}

func (e TypeDependencyEdge) PrintChildArgConverted(converter func(ty *types.Named, isMatched bool) (*types.Named, bool), importMap imports.ImportMap) string {
	isMatched := false
	isConverter := func(named *types.Named) bool {
		if node := e.TypeArgs[0].Node; node != nil {
			isMatched = (node.Matched &^ TypeNodeMatchKindExternal) > 0
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

type TypeDependencyEdgePointer struct {
	Kind TypeDependencyEdgeKind
	Pos  option.Option[int]
}

type TypeArg struct {
	Stack []TypeDependencyEdgePointer
	Node  *TypeNode
	Ty    *types.Named
	Org   types.Type
}

type TypeDependencyEdgeKind uint64

const (
	TypeDependencyEdgeKindAlias = TypeDependencyEdgeKind(1 << iota)
	TypeDependencyEdgeKindArray
	TypeDependencyEdgeKindChan
	TypeDependencyEdgeKindInterface
	TypeDependencyEdgeKindMap
	TypeDependencyEdgeKindNamed
	TypeDependencyEdgeKindPointer
	TypeDependencyEdgeKindSlice
	TypeDependencyEdgeKindStruct
)

func FirstTypeIdent(m map[TypeIdent][]TypeDependencyEdge) (TypeIdent, TypeDependencyEdge) {
	for k, e := range m {
		return k, e[0]
	}
	return TypeIdent{}, TypeDependencyEdge{}
}

func NewTypeGraph(
	pkgs []*packages.Package,
	matcher func(typeInfo *types.Named, external bool) (bool, error),
	genDeclFilter func(*ast.GenDecl) (bool, error),
	typeSpecFilter func(*ast.TypeSpec, types.Object) (bool, error),
) (*TypeGraph, error) {
	graph := &TypeGraph{
		types:    make(map[TypeIdent]*TypeNode),
		matched:  make(map[TypeIdent]*TypeNode),
		external: make(map[TypeIdent]*TypeNode),
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

func (g *TypeGraph) listTypes(
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
						node.Matched |= TypeNodeMatchKindMatched
						g.matched[IdentFromTypesObject(node.Type.Obj())] = node
					}
				}
			}
		}
	}
	return nil
}

func addType(
	to map[TypeIdent]*TypeNode,
	pkg *packages.Package,
	file *ast.File,
	pos int,
	ts *ast.TypeSpec,
	typeInfo *types.Named,
) *TypeNode {
	ident := IdentFromTypesObject(typeInfo.Obj())
	n, ok := to[ident]
	if ok {
		return n
	}
	n = &TypeNode{
		Pkg:  pkg,
		File: file,
		Pos:  pos,
		Ts:   ts,
		Type: typeInfo,
	}
	to[ident] = n
	return n
}

func (g *TypeGraph) buildEdge(matcher func(named *types.Named, external bool) (bool, error)) error {
	for _, node := range g.types {
		// Underlying matches what of go spec.
		// It means what follows type idents like below:
		//
		// type Foo struct {Foo string; Bar int}
		//          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^ this part is underlying
		//
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
	parentNode *TypeNode,
	ty types.Type,
	matcher func(named *types.Named, external bool) (bool, error),
	allType map[TypeIdent]*TypeNode,
	externalType map[TypeIdent]*TypeNode,
	stack []TypeDependencyEdgePointer,
) error {
	return VisitToNamed(
		ty,
		func(named *types.Named, stack []TypeDependencyEdgePointer) error {
			node, ok := allType[IdentFromTypesObject(named.Obj())]
			if ok {
				parentNode.drawEdge(stack, visitOnTypeArgs(named.TypeArgs(), matcher, allType, externalType), named, node)
				return nil
			}
			ok, err := matcher(named, true)
			if !ok || err != nil {
				return err
			}
			externalNode := addType(externalType, nil, nil, -1, nil, named)
			externalNode.Matched |= TypeNodeMatchKindExternal
			parentNode.drawEdge(stack, visitOnTypeArgs(named.TypeArgs(), matcher, allType, externalType), named, externalNode)
			return nil
		},
		stack,
	)
}

func visitOnTypeArgs(
	typeList *types.TypeList,
	matcher func(named *types.Named, external bool) (bool, error),
	allType map[TypeIdent]*TypeNode,
	externalType map[TypeIdent]*TypeNode,
) []TypeArg {
	var typeArgs []TypeArg
	for _, arg := range hiter.AtterAll(typeList) {
		var found bool
		_ = VisitToNamed(
			arg,
			func(named *types.Named, stack []TypeDependencyEdgePointer) error {
				found = true
				// TODO: split this `check if internal types, if not, then add as external type` sequence as a function or a method.
				node, ok := allType[IdentFromTypesObject(named.Obj())]
				if !ok {
					ok, err := matcher(named, true)
					if ok && err == nil {
						externalNode := addType(externalType, nil, nil, -1, nil, named)
						externalNode.Matched |= TypeNodeMatchKindExternal
						node = externalNode
					}
				}
				typeArgs = append(typeArgs, TypeArg{
					Stack: slices.Clone(stack),
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

func VisitToNamed(
	ty types.Type,
	cb func(named *types.Named, stack []TypeDependencyEdgePointer) error,
	stack []TypeDependencyEdgePointer,
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
		return VisitToNamed(x.Rhs(), cb, append(stack, TypeDependencyEdgePointer{Kind: TypeDependencyEdgeKindAlias}))
	case *types.Array:
		return VisitToNamed(x.Elem(), cb, append(stack, TypeDependencyEdgePointer{Kind: TypeDependencyEdgeKindArray}))
	case *types.Basic:
		return nil
	case *types.Chan:
		return VisitToNamed(x.Elem(), cb, append(stack, TypeDependencyEdgePointer{Kind: TypeDependencyEdgeKindChan}))
	case *types.Interface:
		return nil
	case *types.Map:
		return VisitToNamed(x.Elem(), cb, append(stack, TypeDependencyEdgePointer{Kind: TypeDependencyEdgeKindMap}))
	case *types.Pointer:
		return VisitToNamed(x.Elem(), cb, append(stack, TypeDependencyEdgePointer{Kind: TypeDependencyEdgeKindPointer}))
	case *types.Slice:
		return VisitToNamed(x.Elem(), cb, append(stack, TypeDependencyEdgePointer{Kind: TypeDependencyEdgeKindSlice}))
	case *types.Struct:
		// We don't support type-parametrized struct fields.
		// Thus not checking type args.
		for i := range x.NumFields() {
			f := x.Field(i)
			err := VisitToNamed(f.Type(), cb, append(stack, TypeDependencyEdgePointer{Kind: TypeDependencyEdgeKindStruct, Pos: option.Some(i)}))
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (parent *TypeNode) drawEdge(
	stack []TypeDependencyEdgePointer,
	typeArgs []TypeArg,
	childTy *types.Named,
	child *TypeNode,
) {
	if parent.Children == nil {
		parent.Children = make(map[TypeIdent][]TypeDependencyEdge)
	}
	if child.Parent == nil {
		child.Parent = make(map[TypeIdent][]TypeDependencyEdge)
	}

	edge := TypeDependencyEdge{
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

func (g *TypeGraph) MarkDependant(edgeFilter func(edge TypeDependencyEdge) bool) {
	for _, node := range g.types {
		node.Matched = node.Matched &^ TypeNodeMatchKindDependant
	}
	for _, node := range g.IterUpward(false, edgeFilter) {
		if node.Matched.IsExternal() || node.Matched.IsMatched() {
			continue
		}
		node.Matched |= TypeNodeMatchKindDependant
	}
}

func (g *TypeGraph) IterUpward(includeMatched bool, edgeFilter func(edge TypeDependencyEdge) bool) iter.Seq2[TypeIdent, *TypeNode] {
	return func(yield func(TypeIdent, *TypeNode) bool) {
		// record visited nodes to break cyclic link.
		visited := make(map[*TypeNode]bool)
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
	n *TypeNode,
	edgeFilter func(edge TypeDependencyEdge) bool,
	visited map[*TypeNode]bool,
) iter.Seq2[TypeIdent, *TypeNode] {
	return func(yield func(TypeIdent, *TypeNode) bool) {
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
func (n *TypeNode) Fields() iter.Seq2[int, TypeDependencyEdge] {
	_ = n.Type.Underlying().(*types.Struct)
	return func(yield func(int, TypeDependencyEdge) bool) {
		for _, edges := range n.Children {
			for _, e := range edges {
				if !yield(e.Stack[0].Pos.Value(), e) {
					return
				}
			}
		}
	}
}

func (n *TypeNode) FieldsName() iter.Seq2[string, TypeDependencyEdge] {
	structTy := n.Type.Underlying().(*types.Struct)
	return func(yield func(string, TypeDependencyEdge) bool) {
		for _, edges := range n.Children {
			for _, e := range edges {
				if !yield(structTy.Field(e.Stack[0].Pos.Value()).Name(), e) {
					return
				}
			}
		}
	}
}

func (n *TypeNode) ByFieldName(name string) (TypeDependencyEdge, *types.Var, reflect.StructTag, bool) {
	structObj := n.Type.Underlying().(*types.Struct)
	for _, edges := range n.Children {
		for _, e := range edges {
			if e.Stack[0].Pos.IsNone() {
				return TypeDependencyEdge{}, nil, "", false
			}
			pos := e.Stack[0].Pos.Value()
			v := structObj.Field(pos)
			if v.Name() == name {
				return e, v, reflect.StructTag(structObj.Tag(pos)), true
			}
		}
	}
	return TypeDependencyEdge{}, nil, "", false
}

func (g *TypeGraph) EnumerateTypes() iter.Seq2[TypeIdent, *TypeNode] {
	keys := slices.SortedFunc(maps.Keys(g.types), func(i, j TypeIdent) int {
		if c := cmp.Compare(i.PkgPath, j.PkgPath); c != 0 {
			return c
		}
		return cmp.Compare(g.types[i].Pos, g.types[j].Pos)
	})
	return hiter.MapKeys(g.types, slices.Values(keys))
}

func (g *TypeGraph) EnumerateTypesKeys(keys iter.Seq[TypeIdent]) iter.Seq2[TypeIdent, *TypeNode] {
	return hiter.MapKeys(g.types, keys)
}

type TypeDependencyEdgeMap struct {
	node    *TypeNode
	edgeMap map[TypeIdent][]TypeDependencyEdge
	posMap  map[int]TypeDependencyEdge
	nameMap map[string]TypeDependencyEdge
}

func (n *TypeNode) ChildEdgeMap(edgeFilter func(edge TypeDependencyEdge) bool) TypeDependencyEdgeMap {
	if edgeFilter == nil {
		edgeFilter = func(edge TypeDependencyEdge) bool { return true }
	}

	st, isStruct := n.Type.Underlying().(*types.Struct)
	var (
		posMap  map[int]TypeDependencyEdge
		nameMap map[string]TypeDependencyEdge
	)
	if isStruct {
		posMap = make(map[int]TypeDependencyEdge)
		nameMap = make(map[string]TypeDependencyEdge)
	}

	edgeMap := maps.Collect(
		xiter.Filter2(
			func(_ TypeIdent, edges []TypeDependencyEdge) bool {
				if isStruct {
					for pos, edge := range xiter.Map2(
						func(_ int, edge TypeDependencyEdge) (int, TypeDependencyEdge) {
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
				func(i TypeIdent, edges []TypeDependencyEdge) (TypeIdent, []TypeDependencyEdge) {
					return i, slices.Collect(xiter.Filter(edgeFilter, slices.Values(edges)))
				},
				maps.All(n.Children),
			),
		),
	)

	return TypeDependencyEdgeMap{
		node:    n,
		edgeMap: edgeMap,
		posMap:  posMap,
		nameMap: nameMap,
	}
}

func (em TypeDependencyEdgeMap) First() (TypeIdent, TypeDependencyEdge, bool) {
	for k, v := range em.edgeMap {
		return k, v[0], true
	}
	return TypeIdent{}, TypeDependencyEdge{}, false
}

// Fields enumerates its children edges as iter.Seq2[int, typeDependencyEdge] assuming node's underlying type is struct.
// The key of the iterator is position of field in source code order.
func (em TypeDependencyEdgeMap) Fields() iter.Seq2[int, TypeDependencyEdge] {
	_ = em.node.Type.Underlying().(*types.Struct) // panic if not a struct.
	return func(yield func(int, TypeDependencyEdge) bool) {
		for _, edges := range em.edgeMap {
			for _, e := range edges {
				if !yield(e.Stack[0].Pos.Value(), e) {
					return
				}
			}
		}
	}
}

// FieldsName is like [TypeDependencyEdgeMap.Fields] but the key of the pair is field name.
func (em TypeDependencyEdgeMap) FieldsName() iter.Seq2[string, TypeDependencyEdge] {
	structTy := em.node.Type.Underlying().(*types.Struct) // panic if not
	return xiter.Map2(
		func(i int, edge TypeDependencyEdge) (string, TypeDependencyEdge) {
			return structTy.Field(i).Name(), edge
		},
		em.Fields(),
	)
}

// ByFieldPos returns the edge, the field var and the struct tag for the field positioned at pos in source code order,
// It assumes node's underlying is struct type, otherwise panics.
func (em TypeDependencyEdgeMap) ByFieldPos(pos int) (TypeDependencyEdge, *types.Var, reflect.StructTag, bool) {
	st := em.node.Type.Underlying().(*types.Struct) // panic if not
	edge, ok := em.posMap[pos]
	if !ok {
		return TypeDependencyEdge{}, nil, "", false
	}
	return edge, st.Field(pos), reflect.StructTag(st.Tag(pos)), true
}

// ByFieldName is like [TypeDependencyEdgeMap.ByFieldPos] but queries for fieldName.
func (em TypeDependencyEdgeMap) ByFieldName(fieldName string) (TypeDependencyEdge, *types.Var, reflect.StructTag, bool) {
	st := em.node.Type.Underlying().(*types.Struct) // panic if not
	edge, ok := em.nameMap[fieldName]
	if !ok {
		return TypeDependencyEdge{}, nil, "", false
	}
	pos := edge.Stack[0].Pos.Value()
	return edge, st.Field(pos), reflect.StructTag(st.Tag(pos)), true
}
