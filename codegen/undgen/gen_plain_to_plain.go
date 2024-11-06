package undgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"slices"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/und/undtag"
)

type fieldAstExprSet struct {
	Wrapped   ast.Expr
	Unwrapped ast.Expr
}

func generateMethodToPlain(w io.Writer, data *replaceData, node *typeNode, exprMap map[string]fieldAstExprSet) (err error) {
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
	printf(`func (v %s) UndPlain() %s {
`,
		rawTyName, plainTyName,
	)
	defer printf(`}
`)

	named := node.typeInfo
	switch named.Underlying().(type) {
	case *types.Array, *types.Slice, *types.Map:
		generateMethodToPlainElemTypes(printf, node, data.importMap, exprMap)
	case *types.Struct:
		generateMethodToPlainStructFields(printf, ts, node, rawTyName, plainTyName, data.importMap, exprMap)
	}
	// unreachable: should panic instead?
	return nil
}

func printAstExprPanicking(expr ast.Expr) string {
	buf := new(bytes.Buffer)
	err := printer.Fprint(buf, token.NewFileSet(), expr)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func unwrapExprOne(expr ast.Expr, kind typeDependencyEdgeKind) ast.Expr {
	switch kind {
	case typeDependencyEdgeKindArray, typeDependencyEdgeKindSlice:
		return expr.(*ast.ArrayType).Elt
	default:
		return expr.(*ast.MapType).Value
	}
}

func unwrapFieldAlongPath(
	fromExpr, toExpr ast.Expr,
	edge typeDependencyEdge,
	skip int,
) func(wrappee func(string) string, fieldExpr string) string {
	if fromExpr == nil || toExpr == nil {
		return nil
	}
	input := printAstExprPanicking(fromExpr)
	output := printAstExprPanicking(toExpr)

	s := edge.stack[skip:]
	if len(s) == 0 {
		return nil
	}

	initializer := func(expr ast.Expr, kind typeDependencyEdgeKind) string {
		switch kind {
		case typeDependencyEdgeKindArray:
			return fmt.Sprintf("%s{}", printAstExprPanicking(expr))
		default:
			return fmt.Sprintf("make(%s, len(v))", printAstExprPanicking(expr))
		}
	}

	var wrappers []func(string) string
	unwrapped := toExpr
	for p := range hiter.Window(s, 2) {
		unwrapped = unwrapExprOne(unwrapped, p[0].kind)
		initializerExpr := initializer(unwrapped, p[1].kind)
		wrappers = append(wrappers, func(s string) string {
			return fmt.Sprintf(
				`for k, v := range v {
					outer := inner
					mid := %s
					inner := &mid
					%s
					(*outer)[k] = *inner
				}`,
				initializerExpr, s,
			)
		})

	}
	wrappers = append(wrappers, func(s string) string {
		return fmt.Sprintf(
			`for k, v := range v {
				(*inner)[k] = %s
			}`,
			s,
		)
	})
	return func(wrappee func(string) string, fieldExpr string) string {
		expr := wrappee("v")
		for _, wrapper := range slices.Backward(wrappers) {
			expr = wrapper(expr)
		}
		return fmt.Sprintf(`(func (v %s) %s {
	out := %s

	inner := &out
	%s

	return out
})(%s)`,
			input, output, initializer(toExpr, s[0].kind), expr, fieldExpr)
	}

}

func generateMethodToPlainElemTypes(
	printf func(format string, args ...any),
	node *typeNode,
	importMap importDecls,
	exprMap map[string]fieldAstExprSet,
) {
	_generateConversionMethodElemTypes(true, printf, node, importMap, exprMap)
}

func generateMethodToPlainStructFields(
	printf func(format string, args ...any),
	ts *dst.TypeSpec,
	node *typeNode,
	rawTyName, plainTyName string,
	importMap importDecls,
	exprMap map[string]fieldAstExprSet,
) {
	printf(`return %s{
`,
		plainTyName,
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
					fieldConverter, needsArg = generateMethodToPlainDirect(edge, undOpt, ty, importMap)
				} else if isUndConversionImplementor(edge.childType) {
					fieldConverter = func(ident string) string {
						return ident + ".UndPlain()"
					}
				}

				unwrapper := unwrapFieldAlongPath(
					typeToAst(
						edge.parentNode.typeInfo.Underlying().(*types.Struct).Field(edge.stack[0].pos.Value()).Type(),
						edge.parentNode.typeInfo.Obj().Pkg().Path(),
						importMap,
					),
					plainExpr.Wrapped,
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

func generateMethodToPlainDirect(
	edge typeDependencyEdge,
	undOpt undtag.UndOpt,
	typeParam string,
	importMap importDecls,
) (convert func(ident string) string, needsArg bool) {
	return generateConversionMethodDirect(true, edge, undOpt, typeParam, importMap)
}

func optionToPlain(undOpt undtag.UndOpt) (func(ident string) string, bool) {
	switch s := undOpt.States().Value(); {
	default:
		return nil, false
	case s.Def && (s.Null || s.Und):
		return nil, false
	case s.Def:
		return func(fieldName string) string {
			return fmt.Sprintf("%s.Value()", fieldName)
		}, true
	case s.Null || s.Und:
		return func(fieldName string) string { return "nil" }, false
	}
}

func undToPlain(undOpt undtag.UndOpt, importMap importDecls) (func(ident string) string, bool) {
	convertIdent, _ := importMap.Ident(UndPathConversion)
	switch s := undOpt.States().Value(); {
	default:
		return nil, false
	case s.Def && s.Null && s.Und:
		return nil, false
	case s.Def && (s.Null || s.Und):
		return func(ident string) string {
			return fmt.Sprintf("%s.Unwrap().Value()", ident)
		}, true
	case s.Null && s.Und:
		return func(ident string) string {
			return fmt.Sprintf("%s.UndNullish(%s)", convertIdent, ident)
		}, true
	case s.Def:
		return func(ident string) string {
			return fmt.Sprintf("%s.Value()", ident)
		}, true
	case s.Null || s.Und:
		return func(ident string) string { return "nil" }, false
	}
}

func elasticToPlain(isSlice bool, undOpt undtag.UndOpt, typeParam string, importMap importDecls) (func(ident string) string, bool) {
	optionIdent, _ := importMap.Ident(UndTargetTypeOption.ImportPath)
	convertIdent, _ := importMap.Ident(UndPathConversion)
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
				return fmt.Sprintf(`%s.UnwrapElastic%s(%s).Unwrap().Value()`,
					convertIdent, sliceSuffix(isSlice), ident,
				)
			}, true
		case s.Null && s.Und:
			return func(ident string) string {
				return fmt.Sprintf("%s.UndNullish(%s)", convertIdent, ident)
			}, true
		case s.Def:
			return func(ident string) string {
				return fmt.Sprintf("%s.Unwrap().Value()", ident)
			}, true
		case s.Null || s.Und:
			return func(ident string) string {
				return "nil"
			}, false
		}
	}

	states := undOpt.States().Value()
	if !states.Def {
		// return early.
		switch s := states; {
		case s.Null && s.Und:
			return func(ident string) string {
				return fmt.Sprintf("%s.Unwrap().Value()", ident)
			}, true
		case s.Null || s.Und:
			return func(ident string) string {
				return "nil"
			}, false
		}
	}
	// fist, converts Elastic[T] -> Und[[]option.Option[T]]
	c := func(ident string) string {
		return fmt.Sprintf(`%s.UnwrapElastic%s(%s)`,
			convertIdent, sliceSuffix(isSlice), ident,
		)
	}
	wrappers := []func(val string) string{}

	if undOpt.Len().IsSome() {
		lv := undOpt.Len().Value()

		switch lv.Op {
		default:
			panic("unknown len op")
		case undtag.LenOpEqEq:
			// to [n]option.Option[T]
			wrappers = append(wrappers, func(val string) string {
				return fmt.Sprintf(
					`%[1]s.Map(
						%[2]s,
						func(o []%[3]s.Option[%[4]s]) (out [%[5]d]option.Option[%[4]s]) {
					 		copy(out[:], o)
					 		return out
						},
					)`,
					/*1*/ undIdent,
					/*2*/ val,
					/*3*/ optionIdent,
					/*4*/ typeParam,
					/*5*/ lv.Len,
				)
			})
		case undtag.LenOpGr, undtag.LenOpGrEq, undtag.LenOpLe, undtag.LenOpLeEq:
			// other then trim down or append it to the size at most or at least.
			var methodName string
			len := lv.Len
			switch lv.Op {
			case undtag.LenOpGr:
				methodName = "LenNAtLeast"
				len += 1
			case undtag.LenOpGrEq:
				methodName = "LenNAtLeast"
			case undtag.LenOpLe:
				methodName = "LenNAtMost"
				len -= 1
			case undtag.LenOpLeEq:
				methodName = "LenNAtMost"
			}
			wrappers = append(wrappers, func(val string) string {
				return fmt.Sprintf("%s.%s(%d, %s)", convertIdent, suffixSlice(methodName, isSlice), len, val)
			})
		}
	}
	if undOpt.Values().IsSome() {
		v := undOpt.Values().Value()
		switch {
		case v.Nonnull:
			if undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool { return lv.Op == undtag.LenOpEqEq }) {
				wrappers = append(wrappers, func(val string) string {
					return fmt.Sprintf(
						`%[1]s.Map(
							%[2]s,
							func(s [%[3]d]%[4]s.Option[%[5]s]) (r [%[3]d]%[5]s) {
								for i := 0; i < %[3]d; i++ {
									r[i] = s[i].Value()
								}
								return
							},
						)`,
						/*1*/ undIdent,
						/*2*/ val,
						/*3*/ undOpt.Len().Value().Len,
						/*4*/ optionIdent,
						/*5*/ typeParam,
					)
				})
			} else {
				wrappers = append(wrappers, func(val string) string {
					return fmt.Sprintf(
						`%s.NonNull%s(%s)`,
						convertIdent, sliceSuffix(isSlice), val,
					)
				})
			}
			// no other cases at the moment. I don't expect it to expand tho.
		}
	}

	// Then when len == 1, convert und.Und[[1]option.Option[T]] or und.Und[[1]T] to und.Und[option.Option[T]], und.Und[T] respectively
	if undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool { return lv.Op == undtag.LenOpEqEq && lv.Len == 1 }) {
		wrappers = append(wrappers, func(val string) string {
			return fmt.Sprintf("%s.UnwrapLen1%s(%s)", convertIdent, sliceSuffix(isSlice), val)
		})
	}

	// Finally unwrap value based on req,null,und
	if wrapper, _ := undToPlain(undOpt, importMap); wrapper != nil {
		wrappers = append(wrappers, wrapper)
	}

	return func(ident string) string {
		exp := c(ident)
		for _, wrapper := range wrappers {
			exp = wrapper(exp)
		}
		return exp
	}, true
}
