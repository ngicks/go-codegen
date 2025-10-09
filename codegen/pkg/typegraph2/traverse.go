package typegraph2

import (
	"go/types"

	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/und/option"
)

func TraverseToNamed(
	ty types.Type,
	cb func(named *types.Named, stack []EdgeRoute) error,
	stack []EdgeRoute,
) error {
	return TraverseTypes(
		ty,
		nil,
		func(ty types.Type, named *types.Named, stack []EdgeRoute) error {
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
	stopper func(ty types.Type, currentStack []EdgeRoute) bool,
	cb func(ty types.Type, named *types.Named, stack []EdgeRoute) error,
	stack []EdgeRoute,
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
		return TraverseTypes(x.Rhs(), stopper, cb, append(stack, EdgeRoute{Kind: EdgeKindAlias}))
	case *types.Array:
		return TraverseTypes(x.Elem(), stopper, cb, append(stack, EdgeRoute{Kind: EdgeKindArray}))
	case *types.Chan:
		return TraverseTypes(x.Elem(), stopper, cb, append(stack, EdgeRoute{Kind: EdgeKindChan}))
	case *types.Map:
		return TraverseTypes(x.Elem(), stopper, cb, append(stack, EdgeRoute{Kind: EdgeKindMap}))
	case *types.Named:
		for i, t := range hiter.Enumerate(x.TypeArgs().Types()) {
			err := TraverseTypes(
				t,
				stopper,
				cb,
				append(stack, EdgeRoute{Kind: EdgeKindTypeArg, Pos: option.Some(i)}),
			)
			if err != nil {
				return err
			}
		}
		return cb(x, x, stack)
	case *types.Pointer:
		return TraverseTypes(x.Elem(), stopper, cb, append(stack, EdgeRoute{Kind: EdgeKindPointer}))
	case *types.Slice:
		return TraverseTypes(x.Elem(), stopper, cb, append(stack, EdgeRoute{Kind: EdgeKindSlice}))
	case *types.Struct:
		// We don't support type-parametrized struct fields.
		// Thus not checking type args.
		for i := range x.NumFields() {
			f := x.Field(i)
			err := TraverseTypes(
				f.Type(),
				stopper,
				cb,
				append(stack, EdgeRoute{Kind: EdgeKindStruct, Pos: option.Some(i)}),
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
