package typegraph

import (
	"go/types"

	"github.com/ngicks/und/option"
)

type EdgeRoute struct {
	Nodes []EdgeRouteNode
}

type EdgeRouteNode struct {
	Kind EdgeKind
	// position in struct field, or type param.
	Pos option.Option[int]
}

type EdgeKind uint64

const (
	EdgeKindAlias = EdgeKind(1 << iota)
	EdgeKindArray
	EdgeKindChan
	EdgeKindInterface
	EdgeKindMapKey
	EdgeKindMapValue
	EdgeKindNamed
	EdgeKindPointer
	EdgeKindSlice
	EdgeKindStruct
	EdgeKindTypeParam
	EdgeKindTypeArg
)

func WalkTypesToNamed(
	stack []EdgeRouteNode,
	ty types.Type,
	stop func(ty types.Type, currentStack []EdgeRouteNode) bool,
	cb func(named *types.Named, stack []EdgeRouteNode) error,
) error {
	return WalkTypes(
		stack,
		ty,
		nil,
		func(ty types.Type, stack []EdgeRouteNode) error {
			named, ok := ty.(*types.Named)
			if !ok {
				return nil
			}
			return cb(named, stack)
		},
	)
}

func WalkTypes(
	stack []EdgeRouteNode,
	ty types.Type,
	stop func(ty types.Type, currentStack []EdgeRouteNode) bool,
	cb func(ty types.Type, stack []EdgeRouteNode) error,
) error {
	if stop != nil && stop(ty, stack) {
		return cb(ty, stack)
	}
	// types may recurse.
	// but should be impossible without naming type,
	// which breaks visitToNamed from infinite loop.
	switch x := ty.(type) {
	default:
		return cb(x, stack)
	case *types.Alias:
		if err := walkTypeParam(stack, x, stop, cb); err != nil {
			return err
		}
		return WalkTypes(append(stack, EdgeRouteNode{Kind: EdgeKindAlias}), x.Rhs(), stop, cb)
	case *types.Array:
		return WalkTypes(append(stack, EdgeRouteNode{Kind: EdgeKindArray}), x.Elem(), stop, cb)
	case *types.Chan:
		return WalkTypes(append(stack, EdgeRouteNode{Kind: EdgeKindChan}), x.Elem(), stop, cb)
	case *types.Map:
		err := WalkTypes(append(stack, EdgeRouteNode{Kind: EdgeKindMapKey}), x.Key(), stop, cb)
		if err != nil {
			return err
		}
		return WalkTypes(append(stack, EdgeRouteNode{Kind: EdgeKindMapValue}), x.Elem(), stop, cb)
	case *types.Named:
		if err := walkTypeParam(stack, x, stop, cb); err != nil {
			return err
		}
		return cb(x, stack)
	case *types.Pointer:
		return WalkTypes(append(stack, EdgeRouteNode{Kind: EdgeKindPointer}), x.Elem(), stop, cb)
	case *types.Slice:
		return WalkTypes(append(stack, EdgeRouteNode{Kind: EdgeKindSlice}), x.Elem(), stop, cb)
	case *types.Struct:
		for i := range x.NumFields() {
			f := x.Field(i)
			err := WalkTypes(
				append(stack, EdgeRouteNode{Kind: EdgeKindStruct, Pos: option.Some(i)}),
				f.Type(),
				stop,
				cb,
			)
			if err != nil {
				return err
			}
		}
		return nil
	case *types.TypeParam:
		return cb(x, stack)
	}
}

func walkTypeParam[Type interface {
	TypeArgs() *types.TypeList
	TypeParams() *types.TypeParamList
}](
	stack []EdgeRouteNode,
	ty Type,
	stopper func(ty types.Type, currentStack []EdgeRouteNode) bool,
	cb func(ty types.Type, stack []EdgeRouteNode) error,
) error {
	if ty.TypeArgs() != nil {
		for i := range ty.TypeArgs().Len() {
			arg := ty.TypeArgs().At(i)
			err := WalkTypes(
				append(stack, EdgeRouteNode{Kind: EdgeKindTypeArg, Pos: option.Some(i)}),
				arg,
				stopper,
				cb,
			)
			if err != nil {
				return err
			}
		}
	} else if ty.TypeParams() == nil {
		for i := range ty.TypeParams().Len() {
			param := ty.TypeParams().At(i)
			err := WalkTypes(
				append(stack, EdgeRouteNode{Kind: EdgeKindTypeParam, Pos: option.Some(i)}),
				param,
				stopper,
				cb,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
