package undgen

import (
	"fmt"
	"go/types"
	"reflect"
	"slices"

	"github.com/ngicks/go-codegen/codegen/msg"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/und/undtag"
)

var undFieldAllowedEdges = []typeDependencyEdgeKind{
	typeDependencyEdgeKindAlias,
	typeDependencyEdgeKindArray,
	typeDependencyEdgeKindMap,
	typeDependencyEdgeKindSlice,
	typeDependencyEdgeKindStruct,
}

func isUndAllowedEdgeKind(k typeDependencyEdgeKind) bool {
	return slices.Contains(undFieldAllowedEdges, k)
}

func isUndAllowedEdge(p []typeDependencyEdgePointer) bool {
	return slices.ContainsFunc(
		p,
		func(p typeDependencyEdgePointer) bool {
			return isUndAllowedEdgeKind(p.kind)
		},
	)
}

func isUndPlainTarget(named *types.Named, external bool) (bool, error) {
	if external {
		return isUndConversionImplementor(named), nil
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
				if (len(stack) == 0 || isUndAllowedEdge(stack)) && isUndConversionImplementor(named) {
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
	//  - tagged target type field
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
						if (len(stack) == 0 || isUndAllowedEdge(stack)) && isUndType(named) {
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
					func() error {
						return errUndTag(undOpt)
					},
					func(s bool) error {
						return errUndTag(undOpt)
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
					if (len(stack) == 0 || isUndAllowedEdge(stack)) && isUndConversionImplementor(named) {
						found = true
					}
					return nil
				},
				nil,
			)
			return found, nil
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
	onOpt func() T,
	onUnd func(isSlice bool) T,
	onElastic func(isSlice bool) T,
) T {
	switch tt {
	case UndTargetTypeOption:
		return onOpt()
	case UndTargetTypeUnd:
		return onUnd(false)
	case UndTargetTypeSliceUnd:
		return onUnd(true)
	case UndTargetTypeElastic:
		return onElastic(false)
	case UndTargetTypeSliceElastic:
		return onElastic(true)
	}
	panic(fmt.Errorf("not a und type: %#v", tt))
}

func errUndTag(undOpt undtag.UndOpt) error {
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

func isUndConversionImplementor(typeInfo *types.Named) bool {
	return ConstUnd.ConversionMethod.IsImplementor(typeInfo)
}
