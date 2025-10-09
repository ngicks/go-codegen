package typegraph2

import (
	"go/types"

	"github.com/ngicks/und/option"
)

// Edge is a parent to child connection of nodes.
type Edge struct {
	// route of the edge.
	// For example, the type map[K][]chan V has edge of parent to type V
	// via route map, slice and channel.
	//
	// The key type K of the map[K]V is ignored.
	Route []EdgeRoute
	// non-instantiated parent
	Parent *Node
	// instantiated child
	ChildType *types.Named
	// non-instantiated child node.
	ChildNode *Node
	// child is not in []*packages.Package
	External bool
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
	EdgeKindTypeArg
)

type EdgeRoute struct {
	Kind EdgeKind
	// position in struct field or type arg.
	Pos option.Option[int]
}
