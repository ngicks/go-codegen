package cloner

import (
	"fmt"
	"go/types"
	"slices"

	"github.com/ngicks/go-codegen/codegen/matcher"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/typegraph"
)

var (
	clonerMatcher = matcher.ClonerMethod{
		Name: "Clone",
	}
)

type NoCopyHandle int

const (
	// ignore nocopy object
	NoCopyHandleIgnore = NoCopyHandle(1<<iota - 1)
	NoCopyHandleDisallow
	NoCopyHandleCopyPointer
)

type MatcherConfig struct {
	NoCopyHandle  NoCopyHandle
	ChannelHandle NoCopyHandle
}

func (c *MatcherConfig) MatchEdge(e typegraph.Edge) bool {
	return !slices.ContainsFunc(
		e.Stack,
		func(p typegraph.EdgeRouteNode) bool {
			return p.Kind == typegraph.EdgeKindChan
		},
	)
}

func (c *MatcherConfig) MatchType(named *types.Named, external bool) (ok bool, err error) {
	if matcher.IsCloneByAssign(named) {
		return true, nil
	}
	switch x := named.Underlying().(type) {
	default:
		return false, nil
	case *types.Struct:
		for _, f := range pkgsutil.EnumerateFields(x) {
			fieldOk, err := c.matchTy(f.Type())
			if err != nil {
				return false, nil
			}
			if fieldOk {
				ok = true
			}
		}
	case *types.Array, *types.Slice, *types.Map:
		ty := x.(interface{ Elem() types.Type }).Elem()
		ok, err := c.matchTy(ty)
		return err == nil && ok, nil
	}
	return
}

func (c *MatcherConfig) matchTy(ty types.Type) (ok bool, err error) {
	err = typegraph.TraverseTypes(
		ty,
		func(ty types.Type, _ *types.Named, stack []typegraph.EdgeRouteNode) error {
			if slices.ContainsFunc(
				stack,
				func(p typegraph.EdgeRouteNode) bool {
					return p.Kind == typegraph.EdgeKindChan
				},
			) {
				switch c.ChannelHandle {
				case NoCopyHandleIgnore:
					return nil
				case NoCopyHandleDisallow:
					return fmt.Errorf("channel")
				case NoCopyHandleCopyPointer:
				}
			}
			switch ty.(type) {
			case *types.Basic:
				ok = true
			case *types.Interface, *types.Named:
				if matcher.IsNoCopy(ty) {
					switch c.NoCopyHandle {
					case NoCopyHandleIgnore:
						return nil
					case NoCopyHandleDisallow:
					case NoCopyHandleCopyPointer:
						_, isInterface := ty.(*types.Interface)
						if isInterface || (len(stack) > 0 && stack[len(stack)-1].Kind == typegraph.EdgeKindPointer) {
							ok = true
							return nil
						}
					}
					return fmt.Errorf("no copy")
				}
				if clonerMatcher.IsFuncImplementor(ty) ||
					clonerMatcher.IsImplementor(ty) {
					ok = true
				}
			}
			return nil
		},
		nil,
	)
	return ok, err
}
