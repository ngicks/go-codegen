package undgen

import (
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"reflect"
	"slices"

	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-iterator-helper/hiter"
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
	pkgPath  string
	typeName string
}

func (t TypeIdent) TargetType() imports.TargetType {
	return imports.TargetType{ImportPath: t.pkgPath, Name: t.typeName}
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

	Matched typeNodeMatchKind

	Pkg  *packages.Package
	File *ast.File
	// nth type spec in the file.
	Pos  int
	Ts   *ast.TypeSpec
	Type *types.Named
}

type typeNodeMatchKind uint64

const (
	TypeNodeMatchKindMatched = typeNodeMatchKind(1 << iota)
	TypeNodeMatchKindTransitive
	TypeNodeMatchKindExternal
)

func (k typeNodeMatchKind) IsMatched() bool {
	return k&TypeNodeMatchKindMatched > 0
}

func (k typeNodeMatchKind) IsTransitive() bool {
	return k&TypeNodeMatchKindTransitive > 0
}

func (k typeNodeMatchKind) IsExternal() bool {
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

func (e TypeDependencyEdge) HasSingleNamedTypeArg(additionalCond func(named *types.Named) bool) (ok bool, pointer bool) {
	if len(e.TypeArgs) != 1 {
		return false, false
	}
	arg := e.TypeArgs[0]
	isPointer := (len(arg.Stack) == 1 && arg.Stack[0].kind == TypeDependencyEdgeKindPointer)
	if arg.Ty == nil || len(arg.Stack) > 1 || (len(arg.Stack) != 0 && !isPointer) {
		return false, isPointer
	}
	if additionalCond != nil {
		return additionalCond(arg.Ty), isPointer
	}
	return true, isPointer
}

func (e TypeDependencyEdge) LastPointer() option.Option[TypeDependencyEdgePointer] {
	return option.GetSlice(e.Stack, len(e.Stack)-1)
}

func (e TypeDependencyEdge) PrintChildType(importMap imports.ImportMap) string {
	return printAstExprPanicking(
		typeToAst(
			e.ChildType,
			e.ParentNode.Type.Obj().Pkg().Path(),
			importMap,
		),
	)
}

func (e TypeDependencyEdge) PrintChildArg(i int, importMap imports.ImportMap) string {
	return printAstExprPanicking(
		typeToAst(
			e.TypeArgs[i].Org,
			e.ParentNode.Type.Obj().Pkg().Path(),
			importMap,
		),
	)
}

func (e TypeDependencyEdge) PrintChildArgConverted(converter func(ty *types.Named) (*types.Named, bool), importMap imports.ImportMap) string {
	isConverter := func(named *types.Named) bool {
		_, ok := converter(named)
		return ok
	}

	var plainParam types.Type
	if ok, isPointer := e.HasSingleNamedTypeArg(isConverter); ok {
		converted, _ := converter(e.TypeArgs[0].Ty)
		plainParam = converted
		if isPointer {
			plainParam = types.NewPointer(plainParam)
		}
	} else {
		plainParam = e.TypeArgs[0].Org
	}

	return printAstExprPanicking(typeToAst(
		plainParam,
		e.ParentNode.Type.Obj().Pkg().Path(),
		importMap,
	))
}

type TypeDependencyEdgePointer struct {
	kind typeDependencyEdgeKind
	pos  option.Option[int]
}

type TypeArg struct {
	Stack []TypeDependencyEdgePointer
	Node  *TypeNode
	Ty    *types.Named
	Org   types.Type
}

type typeDependencyEdgeKind uint64

const (
	TypeDependencyEdgeKindAlias = typeDependencyEdgeKind(1 << iota)
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
	matcher func(typeInfo *types.Named, within bool) (bool, error),
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
		return VisitToNamed(x.Rhs(), cb, append(stack, TypeDependencyEdgePointer{kind: TypeDependencyEdgeKindAlias}))
	case *types.Array:
		return VisitToNamed(x.Elem(), cb, append(stack, TypeDependencyEdgePointer{kind: TypeDependencyEdgeKindArray}))
	case *types.Basic:
		return nil
	case *types.Chan:
		return VisitToNamed(x.Elem(), cb, append(stack, TypeDependencyEdgePointer{kind: TypeDependencyEdgeKindChan}))
	case *types.Interface:
		return nil
	case *types.Map:
		return VisitToNamed(x.Elem(), cb, append(stack, TypeDependencyEdgePointer{kind: TypeDependencyEdgeKindMap}))
	case *types.Pointer:
		return VisitToNamed(x.Elem(), cb, append(stack, TypeDependencyEdgePointer{kind: TypeDependencyEdgeKindPointer}))
	case *types.Slice:
		return VisitToNamed(x.Elem(), cb, append(stack, TypeDependencyEdgePointer{kind: TypeDependencyEdgeKindSlice}))
	case *types.Struct:
		// We don't support type-parametrized struct fields.
		// Thus not checking type args.
		for i := range x.NumFields() {
			f := x.Field(i)
			err := VisitToNamed(f.Type(), cb, append(stack, TypeDependencyEdgePointer{kind: TypeDependencyEdgeKindStruct, pos: option.Some(i)}))
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

func (g *TypeGraph) markTransitive(edgeFilter func(edge TypeDependencyEdge) bool) {
	for _, node := range g.types {
		node.Matched = node.Matched &^ TypeNodeMatchKindTransitive
	}
	for _, node := range g.IterUpward(false, edgeFilter) {
		if node.Matched.IsExternal() || node.Matched.IsMatched() {
			continue
		}
		node.Matched |= TypeNodeMatchKindTransitive
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
			if includeMatched && !visited[n] {
				visited[n] = true
				if !yield(i, n) {
					return
				}
			}
			for ii, nn := range visitUpward(n, edgeFilter, visited) {
				if !yield(ii, nn) {
					return
				}
			}
		}
	}
}

func visitUpward(
	from *TypeNode,
	edgeFilter func(edge TypeDependencyEdge) bool,
	visited map[*TypeNode]bool,
) iter.Seq2[TypeIdent, *TypeNode] {
	return visitNodes(from, true, edgeFilter, visited)
}

func visitNodes(
	n *TypeNode,
	up bool,
	edgeFilter func(edge TypeDependencyEdge) bool,
	visited map[*TypeNode]bool,
) iter.Seq2[TypeIdent, *TypeNode] {
	return func(yield func(TypeIdent, *TypeNode) bool) {
		direction := n.Parent
		if !up {
			direction = n.Children
		}
		for i, v := range direction {
			for _, edge := range v {
				node := edge.ParentNode
				if !up {
					node = edge.ChildNode
				}

				if edgeFilter == nil || edgeFilter(edge) {
					if !visited[edge.ParentNode] &&
						!yield(i, edge.ParentNode) {
						return
					}
					visited[edge.ParentNode] = true
				}

				for i, n := range visitNodes(node, up, edgeFilter, visited) {
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
				if !yield(e.Stack[0].pos.Value(), e) {
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
				if !yield(structTy.Field(e.Stack[0].pos.Value()).Name(), e) {
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
			if e.Stack[0].pos.IsNone() {
				return TypeDependencyEdge{}, nil, "", false
			}
			pos := e.Stack[0].pos.Value()
			v := structObj.Field(pos)
			if v.Name() == name {
				return e, v, reflect.StructTag(structObj.Tag(pos)), true
			}
		}
	}
	return TypeDependencyEdge{}, nil, "", false
}

func (g *TypeGraph) EnumerateTypesKeys(keys iter.Seq[TypeIdent]) iter.Seq2[TypeIdent, *TypeNode] {
	return hiter.MapKeys(g.types, keys)
}
