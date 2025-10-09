package typegraph2

import (
	"go/ast"
	"go/types"
	"slices"

	"golang.org/x/tools/go/packages"
)

type Node struct {
	Parent map[Ident][]Edge
	// One or many field may refer to same field via same edge.
	Children map[Ident][]Edge

	Matched MatchKind

	Pkg  *packages.Package
	File *ast.File
	// nth type spec in the file.
	Pos  int
	Ts   *ast.TypeSpec
	Type *types.Named

	// Private data parsed through parser
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

func (n *Node) drawEdge(
	stack []EdgeRoute,
	childTy *types.Named,
	child *Node,
	external bool,
) {
	if n.Children == nil {
		n.Children = make(map[Ident][]Edge)
	}
	if child.Parent == nil {
		child.Parent = make(map[Ident][]Edge)
	}

	edge := Edge{
		Route:     slices.Clone(stack),
		Parent:    n,
		ChildType: childTy,
		ChildNode: child,
		External:  external,
	}

	parentIdent := NewIdentFromTypesObject(n.Type.Obj())
	child.Parent[parentIdent] = append(child.Parent[parentIdent], edge)

	childIdent := NewIdentFromTypesObject(child.Type.Obj())
	n.Children[childIdent] = append(n.Children[childIdent], edge)
}
