package cloner

import (
	"go/types"
	"log/slog"
	"slices"
	"strconv"

	"github.com/ngicks/go-codegen/codegen/matcher"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/typegraph"
	"github.com/ngicks/go-iterator-helper/hiter"
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
	FuncHandle    NoCopyHandle

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
	logger := c.logger

	pkgPath, name := matcher.Name(node.Type)
	if pkgPath != "" {
		pkgPath = strconv.Quote(pkgPath) + "."
	}

	logger.Debug("matching", slog.String("name", pkgPath+name))

	priv := option.Assert[clonerPriv](node.Priv).Or(option.Some(clonerPriv{}))
	defer func() {
		// below priv might be changed
		v := priv.Value()
		if v.disallowed {
			logger.Debug("disallowed")
		}
		node.Priv = v
	}()

	switch x := node.Type.Underlying().(type) {
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
				case asSignature(unwrapped_) != nil:
					k = handleSig(c, logger).Or(option.Some(k)).Value()
					// TODO: getting larger, split this function into smaller pieces.
				}
			case *types.Signature:
				k = handleSig(c, logger).Or(option.Some(k)).Value()
			}

			if k != handleKindIgnore && k != handleKindCallCloneFunc {
				return nil
			}

			param, assertOk := unwrapped_.(matcher.HasTypeParam)
			if assertOk && param.TypeArgs().Len() > 0 {
				for i, arg := range hiter.AtterAll(param.TypeArgs()) {
					pkgPath, name := matcher.Name(unwrapped_)
					if pkgPath != "" {
						pkgPath = strconv.Quote(pkgPath) + "."
					}
					_, _, kind, _, argOk := c.matchTy(arg, logger.With("typeArgFor", pkgPath+name))
					if !argOk {
						logger.Debug("ignoring type: type arg at " + strconv.FormatInt(int64(i), 10) + " is disallowed")
						ok = false
						return nil
					} else if kind == handleKindIgnore {
						logger.Debug("ignoring field: type arg at " + strconv.FormatInt(int64(i), 10) + " is ignored")
						k = handleKindIgnore
						return nil
					}
				}
			}
			return nil
		},
		nil,
	)
	return
}

func handleSig(c *MatcherConfig, logger *slog.Logger) option.Option[handleKind] {
	switch c.FuncHandle {
	case NoCopyHandleIgnore:
		logger.Debug("ignoring field since it contains function: if it should be copied place " +
			"//" + DirectivePrefix + DirectiveCommentCopyPtr +
			" as field doc comment",
		)
		return option.Some(handleKindIgnore)
	case NoCopyHandleDisallow:
		logger.Debug(
			"ignoring type since it contains function: if this is mistake change MatchConfig or place " +
				"//" + DirectivePrefix + DirectiveCommentIgnore +
				" or " +
				"//" + DirectivePrefix + DirectiveCommentCopyPtr +
				" as field doc comment",
		)
		return option.None[handleKind]()
	case NoCopyHandleCopyPointer:
		// func itself is a pointer type.
		return option.Some(handleKindAssign)
	}
	logger.Debug("unknown kind", slog.Int("kind", int(c.FuncHandle)))
	return option.Some(handleKindIgnore)
}

func asSignature(ty types.Type) *types.Signature {
	if sig, ok := ty.(*types.Signature); ok {
		return sig
	}
	if named, ok := ty.(*types.Named); ok {
		if sig, ok := named.Underlying().(*types.Signature); ok {
			return sig
		}
	}
	return nil
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
