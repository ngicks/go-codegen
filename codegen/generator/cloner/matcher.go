package cloner

import (
	"go/types"
	"log/slog"
	"slices"

	"github.com/ngicks/go-codegen/codegen/matcher"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/typegraph"
	"github.com/ngicks/und/option"
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

	CustomHandlers CustomHandlers

	logger *slog.Logger
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
	if option.Assert[clonerPriv](e.ParentNode.Priv).IsSomeAnd(func(cp clonerPriv) bool { return cp.disallowed }) {
		return false
	}
	return !slices.ContainsFunc(
		s,
		func(p typegraph.EdgeRouteNode) bool {
			return slices.Contains(disallowedEdge[:], p.Kind)
		},
	)
}

func (c *MatcherConfig) MatchType(node *typegraph.Node, external bool) (ok bool, err error) {
	named := node.Type

	logger := c.logger
	attr := []any{}
	if pkg := named.Obj().Pkg(); pkg != nil {
		attr = append(attr, slog.String("pkgPath", named.Obj().Pkg().Path()))
	}
	attr = append(attr, slog.String("name", named.Obj().Name()))
	logger.Debug("matching", attr...)

	priv := option.Assert[clonerPriv](node.Priv).Or(option.Some(clonerPriv{}))
	defer func() {
		// below priv might be changed
		v := priv.Value()
		if v.disallowed {
			logger.Debug("disallowed")
		}
		node.Priv = v
	}()

	switch x := named.Underlying().(type) {
	default:
		logger.Debug(
			"not matched: unsupported type",
			slog.Any("supported", []string{"struct", "array", "slice", "map"}),
		)
		return false, nil
	case *types.Struct:
		for i, f := range pkgsutil.EnumerateFields(x) {
			conf := *c
			if priv.IsSomeAnd(func(cp clonerPriv) bool { _, ok := cp.lines[i]; return ok }) {
				direction := priv.Value().lines[i]
				logger.Debug("match conf overridden", slog.Any("directive", direction))
				conf = direction.override(conf)
			}
			logger := logger.With(slog.Int("at", i), slog.String("fieldName", f.Name()))
			unwrapped, _, kind, _, fieldOk := conf.matchTy(f.Type(), logger)
			if !fieldOk {
				priv = priv.Map(func(v clonerPriv) clonerPriv { v.disallowed = true; return v })
				logger.Debug("not matched")
				return false, nil
			}
			if kind == handleKindIgnore {
				logger.Debug("field ignored")
				continue
			}
			if !external || (external && (clonerMatcher.IsFuncImplementor(unwrapped) || clonerMatcher.IsImplementor(unwrapped))) {
				logger.Debug(
					"field ok",
					slog.Bool("external", external),
					slog.Bool("implementsCloneFunc", clonerMatcher.IsFuncImplementor(unwrapped)),
					slog.Bool("implementsClone", clonerMatcher.IsFuncImplementor(unwrapped)),
				)
				ok = true
			}
		}
		if ok {
			logger.Debug("matched")
		} else {
			logger.Debug("not matched")
		}
	case *types.Array, *types.Slice, *types.Map:
		ty := x.(interface{ Elem() types.Type }).Elem()
		_, _, kind, _, fieldOk := c.matchTy(ty, logger)
		if kind == handleKindIgnore {
			logger.Debug("not matched: type ignored")
			return false, nil
		}
		if fieldOk {
			logger.Debug("matched: type ok")
		} else {
			priv = priv.Map(func(v clonerPriv) clonerPriv { v.disallowed = true; return v })
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
	handleKindUseCustomHandler
)

func (c *MatcherConfig) matchTy(ty types.Type, logger *slog.Logger) (unwrapped types.Type, stack []typegraph.EdgeRouteNode, k handleKind, customHandlerIndex int, ok bool) {
	k = handleKindIgnore
	ok = true
	customHandlerIndex = -1
	_ = typegraph.TraverseTypes(
		ty,
		func(ty types.Type) bool {
			customHandlerIndex = c.CustomHandlers.Match(ty)
			return customHandlerIndex >= 0
		},
		func(unwrapped_ types.Type, _ *types.Named, stack_ []typegraph.EdgeRouteNode) error {
			unwrapped = unwrapped_
			stack = stack_
			if slices.ContainsFunc(
				stack,
				func(er typegraph.EdgeRouteNode) bool {
					return slices.Contains([]typegraph.EdgeKind{typegraph.EdgeKindStruct, typegraph.EdgeKindInterface}, er.Kind)
				},
			) {
				logger.Debug(
					"disallowed route edge node: struct literal or interface literal",
					slog.Any("stack", stack),
				)
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
					logger.Debug("ignoring field since it contains channel: if it should be copied place " +
						"//" + DirectivePrefix + DirectiveCommentCopyPtr +
						" as field doc comment",
					)
					k = handleKindIgnore
				case NoCopyHandleDisallow:
					logger.Debug(
						"ignoring type since it contains channel: if this is mistake change MatchConfig or place " +
							"//" + DirectivePrefix + DirectiveCommentIgnore +
							" or " +
							"//" + DirectivePrefix + DirectiveCommentCopyPtr +
							" as field doc comment",
					)
					ok = false
				case NoCopyHandleCopyPointer:
					// chan itself is a pointer type.
					k = handleKindAssign
					stack = stack[:i] // reduced to first occurrence of channel
				}
				return nil
			}

			if customHandlerIndex >= 0 {
				k = handleKindUseCustomHandler
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
				switch {
				case matcher.IsNoCopy(unwrapped_):
					switch c.NoCopyHandle {
					case NoCopyHandleIgnore:
						logger.Debug("ignoring field since it contains no copy object: if it should be copied place " +
							"//" + DirectivePrefix + DirectiveCommentCopyPtr +
							" as field doc comment",
						)
						k = handleKindIgnore
						return nil
					case NoCopyHandleDisallow:
						logger.Debug(
							"ignoring type since it contains no copy object: if this is mistake change MatchConfig or place " +
								"//" + DirectivePrefix + DirectiveCommentIgnore +
								" or " +
								"//" + DirectivePrefix + DirectiveCommentCopyPtr +
								" as field doc comment",
						)
					case NoCopyHandleCopyPointer:
						_, isInterface := unwrapped_.(*types.Interface)
						if isInterface || (len(stack) > 0 && stack[len(stack)-1].Kind == typegraph.EdgeKindPointer) {
							k = handleKindAssign
							stack = stack[:len(stack)-1] // ignore last pointer.
							return nil
						}
						logger.Debug("ignoring type: configured to copy pointer of no copy objects but field does not contain it as a pointer or an interface")
					}
					ok = false
					return nil
				case clonerMatcher.IsFuncImplementor(unwrapped_):
					k = handleKindCallCloneFunc
				case clonerMatcher.IsImplementor(unwrapped_):
					k = handleKindCallClone
				case matcher.IsCloneByAssign(unwrapped_):
					k = handleKindAssign
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
) (unwrapped types.Type, stack []typegraph.EdgeRouteNode, k handleKind, customHandlerIndex int) {
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

	unwrapped, stack, k, customHandlerIndex, ok := conf.matchTy(ty, noopLogger)
	if !ok {
		k = handleKindIgnore
		return
	}

	if customHandlerIndex >= 0 {
		return
	}

	if child != nil && child.Matched&^typegraph.MatchKindExternal > 0 {
		if child.Type.TypeParams().Len() == 0 {
			k = handleKindCallClone
		} else {
			k = handleKindCallCloneFunc
		}
	}

	return
}
