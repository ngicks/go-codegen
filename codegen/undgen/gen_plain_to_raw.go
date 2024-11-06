package undgen

import (
	"fmt"
	"go/types"
	"io"
	"slices"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/und/undtag"
)

func generateMethodToRaw(w io.Writer, data *replaceData, node *typeNode, exprMap map[string]fieldAstExprSet) (err error) {
	ts := data.dec.Dst.Nodes[node.ts].(*dst.TypeSpec)
	plainTyName := ts.Name.Name + printTypeParamVars(ts)
	rawTyName, _ := strings.CutSuffix(ts.Name.Name, "Plain")
	rawTyName += printTypeParamVars(ts)
	printf, flush := bufPrintf(w)
	defer func() {
		if err == nil {
			err = flush()
		}
	}()
	printf(`func (v %s) UndRaw() %s {
`,
		plainTyName, rawTyName,
	)
	defer printf(`}
`)

	named := node.typeInfo
	switch named.Underlying().(type) {
	case *types.Array, *types.Slice, *types.Map:
		generateMethodToRawElemTypes(printf, node, data.importMap, exprMap)
	case *types.Struct:
		generateMethodToRawStructFields(printf, ts, node, rawTyName, plainTyName, data.importMap, exprMap)
	}
	// unreachable: should panic instead?
	return nil
}

func generateMethodToRawElemTypes(
	printf func(format string, args ...any),
	node *typeNode,
	importMap importDecls,
	exprMap map[string]fieldAstExprSet,
) {
	_generateConversionMethodElemTypes(false, printf, node, importMap, exprMap)
}

func generateMethodToRawStructFields(
	printf func(format string, args ...any),
	ts *dst.TypeSpec,
	node *typeNode,
	rawTyName, plainTyName string,
	importMap importDecls,
	exprMap map[string]fieldAstExprSet,
) {
	_generateMethodToRawStructFields(false, printf, ts, node, rawTyName, plainTyName, importMap, exprMap)
}
func _generateMethodToRawStructFields(
	toPlain bool,
	printf func(format string, args ...any),
	ts *dst.TypeSpec,
	node *typeNode,
	rawTyName, plainTyName string,
	importMap importDecls,
	exprMap map[string]fieldAstExprSet,
) {
	printf(`return %s{
`,
		or(toPlain, plainTyName, rawTyName),
	)
	defer printf(`}
`)
	dstutil.Apply(
		ts.Type,
		func(c *dstutil.Cursor) bool {
			dstNode := c.Node()
			switch field := dstNode.(type) {
			default:
				return true
			case *dst.Field:
				if len(field.Names) == 0 { // Is it possible?
					return false
				}

				var fieldConverter func(ident string) string
				defer func() {
					if fieldConverter == nil {
						fieldConverter = func(ident string) string {
							return ident
						}
					}
					for _, n := range field.Names {
						printf("\t%s: %s,\n", n.Name, fieldConverter("v."+n.Name))
					}
				}()

				edge, typeVar, tag, ok := node.byFieldName(field.Names[0].Name)
				if !ok {
					return false
				}

				plainExpr := exprMap[typeVar.Name()]

				var needsArg bool
				undTag, ok := tag.Lookup(undtag.TagName)
				if ok {
					undOpt, err := undtag.ParseOption(undTag)
					if err != nil { // this case should already be filtered out.
						panic(err)
					}

					var plainParam types.Type
					if edge.hasSingleNamedTypeArg(isUndConversionImplementor) {
						plainParam, _ = ConstUnd.ConversionMethod.ConvertedType(edge.typeArgs[0].org.(*types.Named))
					} else {
						plainParam = edge.typeArgs[0].org
					}
					ty := printAstExprPanicking(typeToAst(
						plainParam,
						edge.parentNode.typeInfo.Obj().Pkg().Path(),
						importMap,
					))
					fieldConverter, needsArg = generateConversionMethodDirect(toPlain, edge, undOpt, ty, importMap)
				} else if isUndConversionImplementor(edge.childType) {
					fieldConverter = func(ident string) string {
						return ident + or(toPlain, ".UndPlain()", ".UndRaw()")
					}
					needsArg = true
				}

				rawExpr := typeToAst(
					edge.parentNode.typeInfo.Underlying().(*types.Struct).Field(edge.stack[0].pos.Value()).Type(),
					edge.parentNode.typeInfo.Obj().Pkg().Path(),
					importMap,
				)
				unwrapper := unwrapFieldAlongPath(
					or(toPlain, rawExpr, plainExpr.Wrapped),
					or(toPlain, plainExpr.Wrapped, rawExpr),
					edge,
					1, // skip top struct-kind.
				)
				if unwrapper != nil {
					unwrappedConverter := fieldConverter
					fieldConverter = func(ident string) string {
						return unwrapper(
							func(s string) string {
								expr := unwrappedConverter(s)
								if !needsArg {
									expr += "\n_ = v // just to avoid compilation error"
								}
								return expr
							},
							ident,
						)
					}
				}

				return false
			}
		},
		nil,
	)
}

func generateMethodToRawDirect(
	edge typeDependencyEdge,
	undOpt undtag.UndOpt,
	typeParam string,
	importMap importDecls,
) (convert func(ident string) string, needsArg bool) {
	return generateConversionMethodDirect(false, edge, undOpt, typeParam, importMap)
}

func optionToRaw(undOpt undtag.UndOpt, typeParam string, importMap importDecls) (func(ident string) string, bool) {
	optionIdent, _ := importMap.Ident(UndTargetTypeOption.ImportPath)
	switch s := undOpt.States().Value(); {
	default:
		return nil, false
	case s.Def && (s.Null || s.Und):
		return nil, false
	case s.Def:
		return func(ident string) string {
			return fmt.Sprintf("%s.Some(%s)", optionIdent, ident)
		}, true
	case s.Null || s.Und:
		return func(ident string) string {
			return fmt.Sprintf("%s.None[%s]()", optionIdent, typeParam)
		}, true
	}
}

func undToRaw(isSlice bool, undOpt undtag.UndOpt, typeParam string, importMap importDecls) (func(ident string) string, bool) {
	convertIdent, _ := importMap.Ident(UndPathConversion)
	undIdent, _ := importMap.Ident(UndTargetTypeUnd.ImportPath)
	if isSlice {
		undIdent, _ = importMap.Ident(UndTargetTypeSliceUnd.ImportPath)
	}
	switch s := undOpt.States().Value(); {
	default:
		return nil, false
	case s.Def && s.Null && s.Und:
		return nil, false
	case s.Def && (s.Null || s.Und):
		return func(ident string) string {
			return fmt.Sprintf(
				"%s.OptionUnd%s(%t, %s)",
				convertIdent,
				sliceSuffix(isSlice),
				s.Null,
				ident,
			)
		}, true
	case s.Null && s.Und:
		return func(ident string) string {
			return fmt.Sprintf("%s.NullishUnd%s[%s](%s)", convertIdent, sliceSuffix(isSlice), typeParam, ident)
		}, true
	case s.Def:
		return func(ident string) string {
			return fmt.Sprintf("%s.Defined(%s)", undIdent, ident)
		}, true
	case s.Null || s.Und:
		if s.Null {
			return func(ident string) string {
				return fmt.Sprintf("%s.Null[%s]()", undIdent, typeParam)
			}, false
		} else {
			return func(ident string) string {
				return fmt.Sprintf("%s.Undefined[%s]()", undIdent, typeParam)
			}, false
		}
	}
}

func elasticToRaw(isSlice bool, undOpt undtag.UndOpt, typeParam string, importMap importDecls) (func(ident string) string, bool) {
	optionIdent, _ := importMap.Ident(UndTargetTypeOption.ImportPath)
	convertIdent, _ := importMap.Ident(UndPathConversion)

	elasticIdent, _ := importMap.Ident(UndTargetTypeElastic.ImportPath)
	if isSlice {
		elasticIdent, _ = importMap.Ident(UndTargetTypeSliceElastic.ImportPath)
	}

	undIdent, _ := importMap.Ident(UndTargetTypeUnd.ImportPath)
	if isSlice {
		undIdent, _ = importMap.Ident(UndTargetTypeSliceUnd.ImportPath)
	}

	// very really simple case.
	if undOpt.States().IsSome() && undOpt.Len().IsNone() && undOpt.Values().IsNone() {
		switch s := undOpt.States().Value(); {
		default:
			return nil, false
		case s.Def && s.Null && s.Und:
			return nil, false
		case s.Def && (s.Null || s.Und):
			return func(ident string) string {
				return fmt.Sprintf("%s.OptionOptionElastic(%t, %s)", convertIdent, s.Null, ident)
			}, true
		case s.Null && s.Und:
			return func(ident string) string {
				return fmt.Sprintf("%s.NullishElastic[%s](%s)", convertIdent, typeParam, ident)
			}, true
		case s.Def:
			return func(ident string) string {
				return fmt.Sprintf("%s.FromOptions(%s...)", elasticIdent, ident)
			}, true
		case s.Null || s.Und:
			if s.Null {
				return func(ident string) string {
					return fmt.Sprintf("%s.Null[%s]()", elasticIdent, typeParam)
				}, false
			} else {
				return func(ident string) string {
					return fmt.Sprintf("%s.Undefined[%s]()", elasticIdent, typeParam)
				}, false
			}
		}
	}

	states := undOpt.States().Value()
	if !states.Def {
		// return early.
		switch s := states; {
		case s.Null && s.Und:
			return func(ident string) string {
				return fmt.Sprintf("%s.NullishUnd%s[%s](%s)", convertIdent, sliceSuffix(isSlice), typeParam, ident)
			}, true
		case s.Null || s.Und:
			if s.Null {
				return func(ident string) string {
					return fmt.Sprintf("%s.Null[%s]()", elasticIdent, typeParam)
				}, false
			} else {
				return func(ident string) string {
					return fmt.Sprintf("%s.Undefined[%s]()", elasticIdent, typeParam)
				}, false
			}
		}
	}

	c := func(ident string) string {
		return fmt.Sprintf("%s.FromUnd(%s)", elasticIdent, ident)
	}
	wrappers := []func(val string) string{}

	// Below converts much like UndPlain but reversed order.
	// At last, Und[[]option.Option[T]] -> converts Elastic[T]
	if undOpt.Len().IsSome() {
		// if len is EqEq, map Und[[n]option.Option[T]] -> Und[[]option.Option[T]]
		lv := undOpt.Len().Value()
		switch lv.Op {
		case undtag.LenOpEqEq:
			wrappers = append(wrappers, func(val string) string {
				return fmt.Sprintf(
					`%[1]s.Map(
					%[2]s,
					func(s [%[3]d]%[4]s.Option[%[5]s]) []%[4]s.Option[%[5]s] {
						return s[:]
					},
				)`,
					undIdent, val, lv.Len, optionIdent, typeParam)
			})
		}
	}
	if undOpt.Values().IsSome() {
		v := undOpt.Values().Value()
		// Und[[n]T] -> Und[[n]option.Option[T]]
		switch {
		case v.Nonnull:
			if undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool { return lv.Op == undtag.LenOpEqEq }) {
				wrappers = append(wrappers, func(val string) string {
					return fmt.Sprintf(
						`%[1]s.Map(
							%[2]s,
							func(s [%[3]d]%[4]s) (out [%[3]d]%[5]s.Option[%[4]s]) {
								for i := 0; i < %[3]d; i++ {
									out[i] = %[5]s.Some(s[i])
								}
								return
							},
)`,
						/*1*/ undIdent,
						/*2*/ val,
						/*3*/ undOpt.Len().Value().Len,
						/*4*/ typeParam,
						/*5*/ optionIdent,
					)
				})
			} else {
				wrappers = append(wrappers, func(val string) string {
					return fmt.Sprintf(
						"%s.Nullify%s(%s)",
						convertIdent, sliceSuffix(isSlice), val,
					)
				})
			}
			// no other cases at the moment. I don't expect it to expand tho.
		}
	}

	// When len == 1, convert und.Und[[1]option.Option[T]] or und.Und[[1]T] to und.Und[option.Option[T]], und.Und[T] respectively
	if undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool { return lv.Op == undtag.LenOpEqEq && lv.Len == 1 }) {
		wrappers = append(wrappers, func(val string) string {
			return fmt.Sprintf(
				"%s.WrapLen1%s(%s)",
				convertIdent, sliceSuffix(isSlice), val,
			)
		})
	}

	// Finally wrap value based on req,null,und
	if wrapper, _ := undToRaw(isSlice, undOpt, typeParam, importMap); wrapper != nil {
		wrappers = append(wrappers, wrapper)
	}

	return func(ident string) string {
		exp := ident
		for _, wrapper := range slices.Backward(wrappers) {
			exp = wrapper(exp)
		}
		return c(exp)
	}, true
}
