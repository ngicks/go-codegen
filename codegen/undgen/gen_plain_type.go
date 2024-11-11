package undgen

import (
	"go/token"
	"go/types"
	"strconv"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/go-codegen/codegen/astutil"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-codegen/codegen/typegraph"
	"github.com/ngicks/und/undtag"
)

type fieldDstExprSet struct {
	Wrapped   dst.Expr
	Unwrapped dst.Expr
}

func _replaceToPlainTypes(data *replaceData, node *typegraph.TypeNode) (map[string]fieldDstExprSet, bool) {
	ts := data.dec.Dst.Nodes[node.Ts].(*dst.TypeSpec)
	ts.Name.Name += "Plain"
	named := node.Type
	switch named.Underlying().(type) {
	case *types.Array, *types.Slice, *types.Map:
		wrapped, unwrapped := unwrapElemTypes(ts, node, data.importMap)
		return map[string]fieldDstExprSet{"": {wrapped, unwrapped}}, true
	case *types.Struct:
		return unwrapStructFields(ts, node, data.importMap)
	}
	return nil, false
}

func unwrapExprAlongPath(expr *dst.Expr, edge typegraph.TypeDependencyEdge, skip int) *dst.Expr {
	unwrapped := expr
	for _, p := range edge.Stack[skip:] {
		switch p.Kind {
		case typegraph.TypeDependencyEdgeKindArray, typegraph.TypeDependencyEdgeKindSlice:
			next := (*unwrapped).(*dst.ArrayType)
			unwrapped = &next.Elt
		case typegraph.TypeDependencyEdgeKindMap:
			next := (*unwrapped).(*dst.MapType)
			unwrapped = &next.Value
		}
	}
	return unwrapped
}

func unwrapElemTypes(ts *dst.TypeSpec, node *typegraph.TypeNode, importMap imports.ImportMap) (wrapped dst.Expr, unwrapped dst.Expr) {
	var elem *dst.Expr
	switch x := ts.Type.(type) {
	case *dst.ArrayType: // slice or array. difference is Len expr.
		elem = &x.Elt
	case *dst.MapType:
		elem = &x.Value
	}
	// should be only one since we prohibit struct literals.
	_, edge := typegraph.FirstTypeIdent(node.Children)
	if isUndType(edge.ChildType) {
		// matched, wrapped implementor
		unwrapped := unwrapExprAlongPath(elem, edge, 1)
		index := (*unwrapped).(*dst.IndexExpr)
		converted, _ := ConstUnd.ConversionMethod.ConvertedType(edge.TypeArgs[0].Ty)
		index.Index = astutil.TypeToDst(
			converted,
			node.Type.Obj().Pkg().Path(),
			importMap,
		)
		return ts.Type, index
	} else {
		// implementor
		unwrapped := unwrapExprAlongPath(elem, edge, 1)
		converted, _ := ConstUnd.ConversionMethod.ConvertedType(edge.ChildType)
		*unwrapped = astutil.TypeToDst(
			converted,
			node.Type.Obj().Pkg().Path(),
			importMap,
		)
		return ts.Type, *elem
	}
}

func unwrapStructFields(ts *dst.TypeSpec, node *typegraph.TypeNode, importMap imports.ImportMap) (map[string]fieldDstExprSet, bool) {
	edgeMap := node.ChildEdgeMap(isUndPlainAllowedEdge)

	exprMap := make(map[string]fieldDstExprSet)
	var atLeastOne bool
	dstutil.Apply(
		ts.Type,
		func(c *dstutil.Cursor) bool {
			dstNode := c.Node()
			switch field := dstNode.(type) {
			default:
				return true
			case *dst.Field:
				if len(field.Names) == 0 {
					return false // is it even possible?
				}

				edge, _, tag, ok := edgeMap.ByFieldName(field.Names[0].Name)
				if !ok {
					// not found
					return false
				}
				unwrapped := unwrapExprAlongPath(&field.Type, edge, 1)

				undTagValue, hasTag := tag.Lookup(undtag.TagName)
				if hasTag {
					undOpt, err := undtag.ParseOption(undTagValue)
					if err != nil { // This case should be filtered when forming the graph.
						panic(err)
					}
					expr, modified := unwrapUndType((*unwrapped).(*dst.IndexExpr), edge, undOpt, importMap)
					if modified {
						*unwrapped = expr
						atLeastOne = true
						for _, name := range field.Names {
							exprMap[name.Name] = fieldDstExprSet{
								Wrapped:   field.Type,
								Unwrapped: *unwrapped,
							}
						}
					}
					return false
				}

				if named := edge.ChildType; ConstUnd.ConversionMethod.IsImplementor(named) {
					converted, _ := ConstUnd.ConversionMethod.ConvertedType(named)
					var ty types.Type = converted
					if len(edge.Stack) > 0 && edge.Stack[len(edge.Stack)-1].Kind == typegraph.TypeDependencyEdgeKindPointer {
						ty = types.NewPointer(ty)
					}
					*unwrapped = astutil.TypeToDst(
						ty,
						edge.ParentNode.Type.Obj().Pkg().Path(),
						importMap,
					)
					atLeastOne = true
					for _, name := range field.Names {
						exprMap[name.Name] = fieldDstExprSet{
							Wrapped:   field.Type,
							Unwrapped: *unwrapped,
						}
					}
				}

				return false
			}
		},
		nil,
	)
	return exprMap, atLeastOne
}

func unwrapUndType(fieldTy *dst.IndexExpr, edge typegraph.TypeDependencyEdge, undOpt undtag.UndOpt, importMap imports.ImportMap) (expr dst.Expr, modified bool) {
	modified = true

	// default: unchanged.
	// maybe below lines writes expr entirely.
	expr = fieldTy

	// fieldTy -> X.Sel[Index]
	sel := fieldTy.X.(*dst.SelectorExpr) // X.Sel

	if ok, isPointer := edge.HasSingleNamedTypeArg(isUndConversionImplementor); ok {
		arg := edge.TypeArgs[0].Ty
		named, _ := ConstUnd.ConversionMethod.ConvertedType(arg)
		var ty types.Type = named
		if isPointer {
			ty = types.NewPointer(ty)
		}
		fieldTy.Index = astutil.TypeToDst(
			ty,
			edge.ParentNode.Type.Obj().Pkg().Path(),
			importMap,
		)
	}

	_ = matchUndTypeBool(
		namedTypeToTargetType(edge.ChildType),
		false,
		func() {
			switch s := undOpt.States().Value(); {
			default:
				modified = false
			case s.Def && (s.Null || s.Und):
				modified = false
			case s.Def:
				expr = fieldTy.Index // unwrap, simply T.
			case s.Null || s.Und:
				expr = conversionEmptyExpr(importMap)
			}
		},
		func(isSlice bool) {
			switch s := undOpt.States().Value(); {
			case s.Def && s.Null && s.Und:
				modified = false
			case s.Def && (s.Null || s.Und):
				*sel = *importMap.DstExpr(UndTargetTypeOption)
			case s.Null && s.Und:
				fieldTy.Index = conversionEmptyExpr(importMap)
				*sel = *importMap.DstExpr(UndTargetTypeOption)
			case s.Def:
				// unwrap
				expr = fieldTy.Index
			case s.Null || s.Und:
				expr = conversionEmptyExpr(importMap)
			}
		},
		func(isSlice bool) {
			// early return if nothing to change
			if (undOpt.States().IsSomeAnd(func(s undtag.StateValidator) bool {
				return s.Def && s.Null && s.Und
			})) && (undOpt.Len().IsNone() || undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool {
				// when opt is eq, we'll narrow its type to [n]T. but otherwise it remains []T
				return lv.Op != undtag.LenOpEqEq
			})) && (undOpt.Values().IsNone()) {
				modified = false
				return
			}

			// Generally for other cases, replace types
			// und.Und[[]option.Option[T]]
			if isSlice {
				fieldTy.X = importMap.DstExpr(UndTargetTypeSliceUnd)
			} else {
				fieldTy.X = importMap.DstExpr(UndTargetTypeUnd)
			}
			fieldTy.Index = &dst.ArrayType{ // []option.Option[T]
				Elt: &dst.IndexExpr{
					X:     importMap.DstExpr(UndTargetTypeOption),
					Index: fieldTy.Index,
				},
			}

			if undOpt.Len().IsSome() {
				lv := undOpt.Len().Value()
				if lv.Op == undtag.LenOpEqEq {
					if lv.Len == 1 {
						// und.Und[[]option.Option[T]] -> und.Und[option.Option[T]]
						fieldTy.Index = fieldTy.Index.(*dst.ArrayType).Elt
					} else {
						// und.Und[[]option.Option[T]] -> und.Und[[n]option.Option[T]]
						fieldTy.Index.(*dst.ArrayType).Len = &dst.BasicLit{
							Kind:  token.INT,
							Value: strconv.FormatInt(int64(undOpt.Len().Value().Len), 10),
						}
					}
				}
			}

			if undOpt.Values().IsSome() {
				switch x := undOpt.Values().Value(); {
				case x.Nonnull:
					switch x := fieldTy.Index.(type) {
					case *dst.ArrayType:
						// und.Und[[n]option.Option[T]] -> und.Und[[n]T]
						x.Elt = x.Elt.(*dst.IndexExpr).Index
					case *dst.IndexExpr:
						// und.Und[option.Option[T]] -> und.Und[T]
						fieldTy.Index = x.Index
					default:
						panic("implementation error")
					}
				}
			}

			states := undOpt.States().Value()

			switch s := states; {
			default:
			case s.Def && s.Null && s.Und:
				// no conversion
			case s.Def && (s.Null || s.Und):
				// und.Und[[]option.Option[T]] -> option.Option[[]option.Option[T]]
				fieldTy.X = importMap.DstExpr(UndTargetTypeOption)
			case s.Null && s.Und:
				// option.Option[*struct{}]
				fieldTy.Index = conversionEmptyExpr(importMap)
				fieldTy.X = importMap.DstExpr(UndTargetTypeOption)
			case s.Def:
				// und.Und[[]option.Option[T]] -> []option.Option[T]
				expr = fieldTy.Index
			case s.Null || s.Und:
				expr = conversionEmptyExpr(importMap)
			}
		},
	)
	return expr, modified
}
