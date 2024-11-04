package undgen

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"slices"

	"github.com/ngicks/go-codegen/codegen/msg"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/und/undtag"
)

var undFieldAllowedEdges = []typeDependencyEdgeKind{
	typeDependencyEdgeKindAlias,
	typeDependencyEdgeKindArray,
	typeDependencyEdgeKindMap,
	typeDependencyEdgeKindSlice,
	// typeDependencyEdgeKindStruct, // struct literal is not allowed
}

func isUndAllowedEdgeKind(k typeDependencyEdgeKind) bool {
	return slices.Contains(undFieldAllowedEdges, k)
}

func isUndAllowedPointer(p []typeDependencyEdgePointer) bool {
	return len(p) == 0 ||
		// input p should be attached directly named types.
		// struct kind is only allowed as underlying of the named type.
		((p[0].kind == typeDependencyEdgeKindStruct || slices.Contains(undFieldAllowedEdges, p[0].kind)) &&
			hiter.Every(
				func(p typeDependencyEdgePointer) bool {
					return isUndAllowedEdgeKind(p.kind)
				},
				slices.Values(p[1:]),
			))
}

func isUndPlainAllowedEdge(edge typeDependencyEdge) bool {
	return _isUndAllowedEdge(edge, isUndConversionImplementor)
}

func isUndValidatorAllowedEdge(edge typeDependencyEdge) bool {
	return _isUndAllowedEdge(edge, isUndValidatorImplementor)
}

func _isUndAllowedEdge(edge typeDependencyEdge, implementorOf func(named *types.Named) bool) bool {
	if !isUndAllowedPointer(edge.stack) {
		return false
	}
	// struct field
	if len(edge.stack) > 0 && edge.stack[0].kind == typeDependencyEdgeKindStruct && edge.stack[0].pos >= 0 {
		// case 1. tagged und types.
		st := edge.parentNode.typeInfo.Underlying().(*types.Struct)
		_, ok := reflect.StructTag(st.Tag(edge.stack[0].pos)).Lookup(undtag.TagName)
		// we've rejected cases where tag on implementor
		if ok {
			return true
		}
		// case 2. implementor
		if implementorOf(edge.childType) {
			return true
		}
		// case 3. implementor wrapped in und types.
		if isOnlySingleImplementorTypeArg(edge, implementorOf) {
			return true
		}
		return false
	}

	// map, slice, array
	// only allowed element is implementor or implementor wrapped in und types
	if len(edge.stack) > 0 {
		switch edge.stack[0].kind {
		default:
			return false
		case typeDependencyEdgeKindMap, typeDependencyEdgeKindArray, typeDependencyEdgeKindSlice:
		}

		elem := edge.parentNode.typeInfo.Underlying().(interface{ Elem() types.Type }).Elem()
		named, ok := elem.(*types.Named)
		if !ok {
			return false
		}

		if implementorOf(named) {
			return true
		}

		if isOnlySingleImplementorTypeArg(edge, implementorOf) {
			return true
		}
	}

	return false
}

func isOnlySingleImplementorTypeArg(edge typeDependencyEdge, implementorOf func(named *types.Named) bool) bool {
	if len(edge.typeArgs) == 1 {
		arg := edge.typeArgs[0]
		if arg.ty != nil && len(arg.stack) == 0 && implementorOf(arg.ty) {
			return true
		}
	}
	return false
}

func isUndPlainTarget(named *types.Named, external bool) (bool, error) {
	return _isUndTarget(named, external, isUndConversionImplementor)
}

func isUndValidatorTarget(named *types.Named, external bool) (bool, error) {
	return _isUndTarget(named, external, isUndValidatorImplementor)
}

func _isUndTarget(named *types.Named, external bool, implementorOf func(named *types.Named) bool) (bool, error) {
	if external {
		return matchUndType(
			namedTypeToTargetType(named),
			false,
			func() bool { return true }, nil, nil,
		) || implementorOf(named), nil
	}
	switch x := named.Underlying().(type) {
	// case 1: map, array, slice that contain implementor
	// validator and plain/raw converter rely on struct tag.
	// therefore types where struct tags can not be placed are not target,
	// but still it is possible to call implemented methods anyway.
	case *types.Map, *types.Array, *types.Slice:
		elem := named.Underlying().(interface{ Elem() types.Type }).Elem()
		var found bool
		// deeply nested map, array, slice is allowed.
		_ = visitToNamed(
			elem,
			func(named *types.Named, stack []typeDependencyEdgePointer) error {
				if !isUndAllowedPointer(stack) {
					return nil
				}

				if isUndType(named) {
					inner := named.TypeArgs().At(0)
					named, ok := inner.(*types.Named)
					if ok && !isUndType(named) && implementorOf(named) {
						found = true
					}
				} else if !isUndType(named) && implementorOf(named) {
					found = true
				}

				return nil
			},
			nil,
		)
		if found {
			return true, nil
		}
	// case 2: struct type which includes
	//  - untagged implementor field or
	//  - tagged target type field (und types, or even implementor wrapped with und types)
	case *types.Struct:
		for i, f := range pkgsutil.EnumerateFields(x) {
			undTagValue, ok := reflect.StructTag(x.Tag(i)).Lookup(undtag.TagName)
			if ok {
				undOpt, err := undtag.ParseOption(undTagValue)
				if err != nil {
					return false, fmt.Errorf(
						"parsing und tag failed: %s: %w",
						msg.PrintFieldDesc(named.Obj().Name(), i, f), err,
					)
				}
				// tagged. If type is other than und types, it's an error.
				var (
					found      bool
					targetType TargetType
				)
				_ = visitToNamed(
					f.Type(),
					func(named *types.Named, stack []typeDependencyEdgePointer) error {
						if isUndAllowedPointer(stack) && isUndType(named) {
							found = true
							targetType = namedTypeToTargetType(named)
						}
						return nil
					},
					nil,
				)
				if !found {
					return false, fmt.Errorf(
						"und tag is set for non und field: %s",
						msg.PrintFieldDesc(named.Obj().Name(), i, f),
					)
				}
				if err := matchUndType(
					targetType,
					true,
					func() error {
						return errWrongUndTagForNonElastic(undOpt)
					},
					func(s bool) error {
						return errWrongUndTagForNonElastic(undOpt)
					},
					func(s bool) error {
						return nil
					},
				); err != nil {
					return false, fmt.Errorf(
						"%w: %s",
						err, msg.PrintFieldDesc(named.Obj().Name(), i, f),
					)
				}
				return true, nil
			}
			var found bool
			_ = visitToNamed(
				f.Type(),
				func(named *types.Named, stack []typeDependencyEdgePointer) error {
					if isUndAllowedPointer(stack) && implementorOf(named) {
						found = true
					}
					return nil
				},
				nil,
			)
			return found, nil
			// untagged und fields are allowed. they'll be simply just ignored.
		}
	}
	return false, nil
}

func namedTypeToTargetType(named *types.Named) TargetType {
	obj := named.Obj()
	var pkgPath string
	if pkg := obj.Pkg(); pkg != nil {
		pkgPath = pkg.Path()
	}
	return TargetType{
		ImportPath: pkgPath,
		Name:       obj.Name(),
	}
}

// isUndType returns true if named is one of "github.com/ngicks/und/option".Option[T], "github.com/ngicks/und".Und[T],
// "github.com/ngicks/und/elastic".Elastic[T], "github.com/ngicks/und/sliceund".Und[T] or "github.com/ngicks/und/sliceund/elastic".Elastic[T].
func isUndType(named *types.Named) bool {
	return slices.Contains(
		[]TargetType{
			UndTargetTypeOption,
			UndTargetTypeUnd, UndTargetTypeSliceUnd,
			UndTargetTypeElastic, UndTargetTypeSliceElastic,
		},
		namedTypeToTargetType(named),
	)
}

func matchUndType[T any](
	tt TargetType,
	panicOnMismatch bool,
	onOpt func() T,
	onUnd func(isSlice bool) T,
	onElastic func(isSlice bool) T,
) T {
	switch tt {
	case UndTargetTypeOption:
		return onOpt()
	case UndTargetTypeUnd:
		if onUnd != nil {
			return onUnd(false)
		}
		return onOpt()
	case UndTargetTypeSliceUnd:
		if onUnd != nil {
			return onUnd(true)
		}
		return onOpt()
	case UndTargetTypeElastic:
		if onElastic != nil {
			return onElastic(false)
		}
		if onUnd != nil {
			return onUnd(false)
		}
		return onOpt()
	case UndTargetTypeSliceElastic:
		if onElastic != nil {
			return onElastic(true)
		}
		if onUnd != nil {
			return onUnd(true)
		}
		return onOpt()
	}
	if panicOnMismatch {
		panic(fmt.Errorf("not a und type: %#v", tt))
	}
	return *new(T)
}

func matchUndTypeBool(
	tt TargetType,
	panicOnMismatch bool,
	onOpt func(),
	onUnd func(isSlice bool),
	onElastic func(isSlice bool),
) bool {
	var (
		_onOpt             func() bool
		_onUnd, _onElastic func(isSlice bool) bool
	)
	if onOpt != nil {
		_onOpt = func() bool {
			onOpt()
			return true
		}
	}
	if onUnd != nil {
		_onUnd = func(isSlice bool) bool {
			onUnd(isSlice)
			return true
		}
	}
	if onElastic != nil {
		_onElastic = func(isSlice bool) bool {
			onElastic(isSlice)
			return true
		}
	}
	return matchUndType(
		tt,
		panicOnMismatch,
		_onOpt,
		_onUnd,
		_onElastic,
	)
}

func errWrongUndTagForNonElastic(undOpt undtag.UndOpt) error {
	var v string
	if undOpt.Len().IsSome() {
		v = undtag.UndTagValueLen
	}
	if undOpt.Values().IsSome() {
		v += "|" + undtag.UndTagValueValues
	}
	if v != "" {
		return fmt.Errorf("%s specified for non elastic type", v)
	}
	return nil
}

func isUndConversionImplementor(named *types.Named) bool {
	return ConstUnd.ConversionMethod.IsImplementor(named)
}

func isUndValidatorImplementor(named *types.Named) bool {
	// und types are already implementors.
	// exclude them first.
	return !isUndType(named) && ConstUnd.ValidatorMethod.IsImplementor(named)
}

func excludeUndIgnoredCommentedGenDecl(genDecl *ast.GenDecl) (bool, error) {
	direction, _, err := ParseUndComment(genDecl.Doc)
	if err != nil {
		return false, err
	}
	return !direction.MustIgnore(), nil
}

func excludeUndIgnoredCommentedTypeSpec(ts *ast.TypeSpec, _ types.Object) (bool, error) {
	direction, _, err := ParseUndComment(ts.Doc)
	if err != nil {
		return false, err
	}
	return !direction.MustIgnore(), nil
}
