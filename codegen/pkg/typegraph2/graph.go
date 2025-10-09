package typegraph2

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

	"github.com/ngicks/go-codegen/codegen/pkg/pkgsutil"
	"github.com/ngicks/go-iterator-helper/hiter"
	"golang.org/x/tools/go/packages"
)

// Graph enumerates types from []*packages.Package and
// draws edges between named types.
//
// Types can be marked using matcher after built.
//
// There are some kinds of edge:
//
//   - as struct fields
//   - as indirect types (boxed by map, slice, array or channel e.g. K and V in map[K][]V)
//   - as type args
//   - as type aliases
type Graph struct {
	// all types in []*packages.Package
	types map[Ident]*Node
	// cached matched
	matched map[Ident]*Node
	// external dependencies
	external map[Ident]*Node
}

func New(
	pkgs []*packages.Package,
	genDeclFilter func(*ast.GenDecl) (bool, error),
	typeSpecFilter func(*ast.TypeSpec, types.Object) (bool, error),
	privParser func(n *Node) (any, error),
) (*Graph, error) {
	graph := &Graph{
		types:    make(map[Ident]*Node),
		matched:  make(map[Ident]*Node),
		external: make(map[Ident]*Node),
	}

	err := graph.listTypes(pkgs, genDeclFilter, typeSpecFilter, privParser)
	if err != nil {
		return graph, err
	}

	err = graph.buildEdge()
	if err != nil {
		return graph, err
	}

	return graph, nil
}

func (g *Graph) listTypes(
	pkgs []*packages.Package,
	genDeclFilter func(*ast.GenDecl) (bool, error),
	typeSpecFilter func(*ast.TypeSpec, types.Object) (bool, error),
	privParser func(n *Node) (any, error),
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
						pos += len(genDecl.Specs)
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
					var err error
					node := addType(g.types, pkg, file, currentPos, ts, named)
					if privParser != nil {
						node.Priv, err = privParser(node)
						if err != nil {
							return fmt.Errorf("parsing priv: %w", err)
						}
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
	ident := NewIdentFromTypesObject(typeInfo.Obj())
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

func (g *Graph) buildEdge() error {
	for _, node := range g.types {
		// Underlying matches what of go spec.
		// It means what follows type idents like below:
		//
		// type Foo struct {Foo string; Bar int}
		//          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^ this part is underlying
		err := visitTypes(
			node,
			node.Type.Underlying(),
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

func visitTypes(
	parentNode *Node,
	ty types.Type,
	allType map[Ident]*Node,
	externalType map[Ident]*Node,
	stack []EdgeRoute,
) error {
	return TraverseToNamed(
		ty,
		func(named *types.Named, stack []EdgeRoute) error {
			node, ok := allType[NewIdentFromTypesObject(named.Obj())]
			if ok {
				parentNode.drawEdge(
					stack,
					named,
					node,
					false,
				)
				return nil
			}
			externalNode := addType(externalType, nil, nil, -1, nil, named)
			parentNode.drawEdge(
				stack,
				named,
				externalNode,
				true,
			)
			return nil
		},
		stack,
	)
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
				node := edge.Parent

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
		hiter.Filter2(
			func(_ Ident, edges []Edge) bool {
				if isStruct {
					for pos, edge := range hiter.Map2(
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
			hiter.Map2(
				func(i Ident, edges []Edge) (Ident, []Edge) {
					return i, slices.Collect(hiter.Filter(edgeFilter, slices.Values(edges)))
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
	return hiter.Map2(
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
