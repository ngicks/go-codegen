package cloner

import (
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
	NoCopyHandleIgnore   NoCopyHandle = 0
	NoCopyHandleDisallow NoCopyHandle = 1 << iota
	NoCopyHandleCopyPointer
)

type MatcherConfig struct {
	NoCopyHandle  NoCopyHandle
	ChannelHandle NoCopyHandle
}

var disallowedEdge = [...]typegraph.EdgeKind{
	typegraph.EdgeKindChan,
	typegraph.EdgeKindStruct,
	typegraph.EdgeKindInterface,
}

func (c *MatcherConfig) MatchEdge(e typegraph.Edge) bool {
	s := e.Stack
	if len(s) > 0 && s[0].Kind == typegraph.EdgeKindStruct {
		s = s[1:]
	}
	return !slices.ContainsFunc(
		s,
		func(p typegraph.EdgeRouteNode) bool {
			return slices.Contains(disallowedEdge[:], p.Kind)
		},
	)
}

func (c *MatcherConfig) MatchType(named *types.Named, external bool) (ok bool, err error) {
	if !matcher.IsNoCopy(named) && matcher.IsCloneByAssign(named) {
		return true, nil
	}
	switch x := named.Underlying().(type) {
	default:
		return false, nil
	case *types.Struct:
		for _, f := range pkgsutil.EnumerateFields(x) {
			unwrapped, _, kind, fieldOk := c.matchTy(f.Type())
			if !fieldOk {
				return false, nil
			}
			if kind == handleKindIgnore {
				continue
			}
			if !external || (external && (clonerMatcher.IsFuncImplementor(unwrapped) || clonerMatcher.IsImplementor(unwrapped))) {
				ok = true
			}
		}
	case *types.Array, *types.Slice, *types.Map:
		ty := x.(interface{ Elem() types.Type }).Elem()
		_, _, kind, fieldOk := c.matchTy(ty)
		if kind == handleKindIgnore {
			return false, nil
		}
		return fieldOk, nil
	}
	return
}

type handleKind int

const (
	handleKindIgnore handleKind = iota + 1
	handleKindAssign
	handleKindCallCb
	handleKindCallClone
	handleKindCallCloneFunc
)

func (c *MatcherConfig) matchTy(ty types.Type) (unwrapped types.Type, stack []typegraph.EdgeRouteNode, k handleKind, ok bool) {
	k = handleKindIgnore
	ok = true
	_ = typegraph.TraverseTypes(
		ty,
		func(unwrapped_ types.Type, _ *types.Named, stack_ []typegraph.EdgeRouteNode) error {
			unwrapped = unwrapped_
			stack = stack_
			if slices.ContainsFunc(
				stack,
				func(er typegraph.EdgeRouteNode) bool {
					return slices.Contains([]typegraph.EdgeKind{typegraph.EdgeKindStruct, typegraph.EdgeKindInterface}, er.Kind)
				},
			) {
				// disallow struct, interface literal
				ok = false
				return nil
			}
			if i := slices.IndexFunc(
				stack,
				func(p typegraph.EdgeRouteNode) bool {
					return p.Kind == typegraph.EdgeKindChan
				},
			); i >= 0 {
				switch c.ChannelHandle {
				case NoCopyHandleIgnore:
					k = handleKindIgnore
				case NoCopyHandleDisallow:
					ok = false
				case NoCopyHandleCopyPointer:
					// chan itself is a pointer type.
					k = handleKindAssign
					stack = stack[:i] // reduced to first occurrence of channel
				}
				return nil
			}

			switch unwrapped_.(type) {
			case *types.Basic:
				k = handleKindAssign
				return nil
			case *types.TypeParam:
				k = handleKindCallCb
				return nil
			case *types.Named:
				if matcher.IsNoCopy(unwrapped_) {
					switch c.NoCopyHandle {
					case NoCopyHandleIgnore:
						k = handleKindIgnore
						return nil
					case NoCopyHandleDisallow:
					case NoCopyHandleCopyPointer:
						_, isInterface := unwrapped_.(*types.Interface)
						if isInterface || (len(stack) > 0 && stack[len(stack)-1].Kind == typegraph.EdgeKindPointer) {
							k = handleKindAssign
							stack = stack[:len(stack)-1] // ignore last pointer.
							return nil
						}
					}
					ok = false
					return nil
				}
				if clonerMatcher.IsFuncImplementor(unwrapped_) {
					k = handleKindCallCloneFunc
				} else if clonerMatcher.IsImplementor(unwrapped_) {
					k = handleKindCallClone
				}
			}
			return nil
		},
		nil,
	)
	return
}

func (c *MatcherConfig) handleField(
	pos int,
	parent *typegraph.Node,
	child *typegraph.Node,
	ty types.Type,
) (unwrapped types.Type, stack []typegraph.EdgeRouteNode, k handleKind) {
	conf := *c
	if parent != nil && pos >= 0 {
		priv, ok := parent.Priv.(clonerPriv)
		if ok {
			direction, ok := priv.lines[pos]
			if ok {
				conf = direction.override(conf)
			}
		}
	}

	unwrapped, stack, k, ok := conf.matchTy(ty)
	if !ok {
		k = handleKindIgnore
		return
	}

	if child != nil && child.Matched != 0 {
		if child.Type.TypeParams().Len() == 0 {
			k = handleKindCallClone
		} else {
			k = handleKindCallCloneFunc
		}
	}

	return
}
