package undgen

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"slices"

	"github.com/ngicks/go-codegen/codegen/astmeta"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-codegen/codegen/msg"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/typegraph"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/und/undtag"
)

var undFieldAllowedEdges = []typegraph.TypeDependencyEdgeKind{
	typegraph.TypeDependencyEdgeKindAlias,
	typegraph.TypeDependencyEdgeKindArray,
	typegraph.TypeDependencyEdgeKindMap,
	typegraph.TypeDependencyEdgeKindSlice,
	// typeDependencyEdgeKindStruct, // struct literal is not allowed
}

func isUndAllowedEdgeKind(k typegraph.TypeDependencyEdgeKind) bool {
	return slices.Contains(undFieldAllowedEdges, k)
}

func isUndAllowedPointer(ty *types.Named, p []typegraph.TypeDependencyEdgePointer) bool {
	if len(p) == 0 {
		// directly attached named type.
		// Basically for type param.
		return true
	}
	if p[0].Kind == typegraph.TypeDependencyEdgeKindStruct {
		p = p[1:]
	}
	if len(p) == 0 {
		return true
	}
	rest, last := p[:len(p)-1], p[len(p)-1]
	if !slices.Contains(undFieldAllowedEdges, last.Kind) &&
		// non und type is allowed to be pointer.
		// but only for the last element
		last.Kind != typegraph.TypeDependencyEdgeKindPointer && isUndType(ty) {
		return false
	}
	return hiter.Every(
		func(p typegraph.TypeDependencyEdgePointer) bool {
			return isUndAllowedEdgeKind(p.Kind)
		},
		slices.Values(rest),
	)
}

func isUndPlainAllowedEdge(edge typegraph.TypeDependencyEdge) bool {
	return _isUndAllowedEdge(edge, isUndConversionImplementor)
}

func isUndValidatorAllowedEdge(edge typegraph.TypeDependencyEdge) bool {
	return _isUndAllowedEdge(edge, isUndValidatorImplementor)
}

func _isUndAllowedEdge(edge typegraph.TypeDependencyEdge, implementorOf func(named *types.Named) bool) bool {
	if !isUndAllowedPointer(edge.ChildType, edge.Stack) {
		return false
	}
	// struct field
	if len(edge.Stack) > 0 && edge.Stack[0].Kind == typegraph.TypeDependencyEdgeKindStruct && edge.Stack[0].Pos.IsSome() {
		// case 1. tagged und types.
		st := edge.ParentNode.Type.Underlying().(*types.Struct)
		_, ok := reflect.StructTag(st.Tag(edge.Stack[0].Pos.Value())).Lookup(undtag.TagName)
		// we've rejected cases where tag on implementor
		if ok {
			return true
		}
		// case 2. implementor
		if implementorOf(edge.ChildType) {
			return true
		}
		// case 3. dependant match
		if edge.IsChildMatched() {
			return true
		}
		// case 4. implementor wrapped in und types.
		if ok, _ := edge.HasSingleNamedTypeArg(implementorOf); ok {
			return true
		}
		// case 5. dependant wrapped in und types.
		if isUndType(edge.ChildType) && edge.IsTypeArgMatched() {
			return true
		}
		return false
	}

	// map, slice, array
	// only allowed element is implementor or implementor wrapped in und types
	if len(edge.Stack) > 0 {
		switch edge.Stack[0].Kind {
		default:
			return false
		case typegraph.TypeDependencyEdgeKindMap, typegraph.TypeDependencyEdgeKindArray, typegraph.TypeDependencyEdgeKindSlice:
		}
		if implementorOf(edge.ChildType) {
			return true
		}
		if edge.IsChildMatched() {
			return true
		}
		if ok, _ := edge.HasSingleNamedTypeArg(implementorOf); ok {
			return true
		}
		if isUndType(edge.ChildType) && edge.IsTypeArgMatched() {
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
		return matchUndTypeBool(
			namedTypeToTargetType(named),
			false,
			func() {}, nil, nil,
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
		_ = typegraph.VisitToNamed(
			elem,
			func(named *types.Named, stack []typegraph.TypeDependencyEdgePointer) error {
				if !isUndAllowedPointer(named, stack) {
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
		var atLeastOne bool
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
					targetType imports.TargetType
				)
				_ = typegraph.VisitToNamed(
					f.Type(),
					func(named *types.Named, stack []typegraph.TypeDependencyEdgePointer) error {
						if isUndAllowedPointer(named, stack) && isUndType(named) {
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
				atLeastOne = true
				continue
			}
			var found bool
			_ = typegraph.VisitToNamed(
				f.Type(),
				func(named *types.Named, stack []typegraph.TypeDependencyEdgePointer) error {
					if isUndAllowedPointer(named, stack) && implementorOf(named) {
						found = true
					}
					return nil
				},
				nil,
			)
			if found {
				atLeastOne = true
				continue
			}
			// untagged und fields are allowed. they'll be simply just ignored.
		}
		return atLeastOne, nil
	}
	return false, nil
}

func namedTypeToTargetType(named *types.Named) imports.TargetType {
	obj := named.Obj()
	var pkgPath string
	if pkg := obj.Pkg(); pkg != nil {
		pkgPath = pkg.Path()
	}
	return imports.TargetType{
		ImportPath: pkgPath,
		Name:       obj.Name(),
	}
}

// isUndType returns true if named is one of "github.com/ngicks/und/option".Option[T], "github.com/ngicks/und".Und[T],
// "github.com/ngicks/und/elastic".Elastic[T], "github.com/ngicks/und/sliceund".Und[T] or "github.com/ngicks/und/sliceund/elastic".Elastic[T].
func isUndType(named *types.Named) bool {
	return slices.Contains(
		[]imports.TargetType{
			UndTargetTypeOption,
			UndTargetTypeUnd, UndTargetTypeSliceUnd,
			UndTargetTypeElastic, UndTargetTypeSliceElastic,
		},
		namedTypeToTargetType(named),
	)
}

func matchUndType[T any](
	tt imports.TargetType,
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
	tt imports.TargetType,
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
	direction, _, err := astmeta.ParseComment(genDecl.Doc)
	if err != nil {
		return false, err
	}
	return !direction.MustIgnore(), nil
}

func excludeUndIgnoredCommentedTypeSpec(ts *ast.TypeSpec, _ types.Object) (bool, error) {
	direction, _, err := astmeta.ParseComment(ts.Doc)
	if err != nil {
		return false, err
	}
	return !direction.MustIgnore(), nil
}
