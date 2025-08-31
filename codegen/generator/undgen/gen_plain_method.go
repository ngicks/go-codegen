package undgen

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"log/slog"
	"slices"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/go-codegen/codegen/pkg/astutil"
	"github.com/ngicks/go-codegen/codegen/pkg/directive"
	"github.com/ngicks/go-codegen/codegen/pkg/imports"
	"github.com/ngicks/go-codegen/codegen/pkg/typegraph"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/und/undtag"
)

type fieldAstExprSet struct {
	Wrapped   ast.Expr
	Unwrapped ast.Expr
}

func unwrapExprOne(expr ast.Expr, kind typegraph.EdgeKind) ast.Expr {
	switch kind {
	case typegraph.EdgeKindArray, typegraph.EdgeKindSlice:
		return expr.(*ast.ArrayType).Elt
	case typegraph.EdgeKindMap:
		return expr.(*ast.MapType).Value
	}
	return expr
}

func unwrapFieldAlongPath(
	fromExpr, toExpr ast.Expr,
	edge typegraph.Edge,
	skip int,
) func(wrappee func(string) string, fieldExpr string) string {
	if fromExpr == nil || toExpr == nil {
		return nil
	}
	input := astutil.PrintAstExprPanicking(fromExpr)
	output := astutil.PrintAstExprPanicking(toExpr)

	s := edge.Stack[skip:]
	if len(s) > 0 && s[len(s)-1].Kind == typegraph.EdgeKindPointer {
		s = s[:len(s)-1]
	}
	if len(s) == 0 {
		return nil
	}

	initializer := func(expr ast.Expr, kind typegraph.EdgeKind) string {
		switch kind {
		case typegraph.EdgeKindArray:
			return fmt.Sprintf("%s{}", astutil.PrintAstExprPanicking(expr))
		default:
			return fmt.Sprintf("make(%s, len(v))", astutil.PrintAstExprPanicking(expr))
		}
	}

	var wrappers []func(string) string
	unwrapped := toExpr
	for p := range hiter.Window(s, 2) {
		unwrapped = unwrapExprOne(unwrapped, p[0].Kind)
		initializerExpr := initializer(unwrapped, p[1].Kind)
		wrappers = append(wrappers, func(s string) string {
			return fmt.Sprintf(
				`for k, v := range v {
					outer := &inner
					inner := %s
					%s
					(*outer)[k] = inner
				}`,
				initializerExpr, s,
			)
		})

	}
	wrappers = append(wrappers, func(s string) string {
		return fmt.Sprintf(
			`for k, v := range v {
				inner[k] = %s
			}`,
			s,
		)
	})
	return func(wrappee func(string) string, fieldExpr string) string {
		wrappers = slices.Insert(wrappers, 0, func(expr string) string {
			return fmt.Sprintf(
				`(func (v %s) %s {
					out := %s

					inner := out
					%s

					return out
				})(%s)`,
				input, output, initializer(toExpr, s[0].Kind), expr, fieldExpr)
		})
		expr := wrappee("v")
		for _, wrapper := range slices.Backward(wrappers) {
			expr = wrapper(expr)
		}
		return expr
	}
}

func generateConversionMethod(w io.Writer, data *typegraph.ReplaceData, node *typegraph.Node, exprMap map[string]fieldAstExprSet) (err error) {
	ts := data.Dec.Dst.Nodes[node.Ts].(*dst.TypeSpec)
	plainTyName := ts.Name.Name + astutil.PrintTypeParamsDst(ts)
	rawTyName, _ := strings.CutSuffix(ts.Name.Name, "Plain")
	rawTyName += astutil.PrintTypeParamsDst(ts)

	printf, flush := astutil.BufPrintf(w)
	defer func() {
		if err == nil {
			err = flush()
		}
	}()

	generateToRawOrToPlain(true, printf, plainTyName, rawTyName, ts, data, node, exprMap)
	generateToRawOrToPlain(false, printf, plainTyName, rawTyName, ts, data, node, exprMap)

	return
}

func generateToRawOrToPlain(
	toPlain bool,
	printf func(format string, args ...any),
	plainTyName, rawTyName string,
	ts *dst.TypeSpec,
	data *typegraph.ReplaceData,
	node *typegraph.Node,
	exprMap map[string]fieldAstExprSet,
) {
	printf("//%s%s\n", directive.DirectivePrefix, directive.DirectiveCommentGenerated)
	printf(`func (v %s) %s() %s {
`,
		or(
			toPlain,
			[]any{rawTyName, "UndPlain", plainTyName},
			[]any{plainTyName, "UndRaw", rawTyName},
		)...,
	)
	defer printf(`}

`)

	named := node.Type
	switch named.Underlying().(type) {
	case *types.Array, *types.Slice, *types.Map:
		generateConversionMethodElemTypes(toPlain, printf, node, data.ImportMap, exprMap)
	case *types.Struct:
		generateConversionMethodStructFields(toPlain, printf, ts, node, rawTyName, plainTyName, data.ImportMap, exprMap)
	default:
		slog.Default().Error(
			"implementation error",
			slog.String("rawTyName", rawTyName),
			slog.String("plainTyName", plainTyName),
			slog.Any("type", named),
		)
		panic("implementation error")
	}
}

func generateConversionMethodElemTypes(
	toPlain bool,
	printf func(format string, args ...any),
	node *typegraph.Node,
	importMap imports.ImportMap,
	exprMap map[string]fieldAstExprSet,
) {
	_, edge := typegraph.FirstTypeIdent(node.Children) // must be only one.

	rawExpr := astutil.TypeToAst(
		edge.ParentNode.Type.Underlying(),
		edge.ParentNode.Type.Obj().Pkg().Path(),
		importMap,
	)

	var plainExprWrapped, plainExprUnwrapped ast.Expr
	for _, v := range exprMap {
		plainExprWrapped = v.Wrapped
		plainExprUnwrapped = v.Unwrapped
	}

	unwrapper := unwrapFieldAlongPath(
		or(toPlain, rawExpr, plainExprWrapped),
		or(toPlain, plainExprWrapped, rawExpr),
		edge,
		0,
	)

	if isUndType(edge.ChildType) {
		_, isPointer := edge.HasSingleNamedTypeArg(func(named *types.Named) bool { return true })
		// matched, wrapped implementor
		converter, _ := _generateConversionMethodImplementorMapper(
			toPlain,
			edge,
			prefixPointer(isPointer, edge.PrintChildArg(0, importMap)),
			prefixPointer(
				isPointer,
				edge.PrintChildArgConverted(
					plainConverter,
					importMap,
				)),
			importMap,
			isPointer,
			func(ident string) string {
				return ident
			},
		)
		printf(`return ` + unwrapper(converter, "v"))
		return
	} else {
		isPointer := edge.LastPointer().IsSomeAnd(func(tdep typegraph.EdgeRouteNode) bool {
			return tdep.Kind == typegraph.EdgeKindPointer
		})
		// implementor
		printf(`return ` + unwrapper(
			_generateConversionMethodInvocationExpr(
				toPlain,
				isPointer,
				prefixPointer(isPointer, edge.PrintChildType(importMap)),
				astutil.PrintAstExprPanicking(plainExprUnwrapped),
			),
			"v"),
		)
		return
	}

}

func prefixPointer(isPointer bool, s string) string {
	if isPointer {
		return "*" + s
	}
	return s
}

func generateConversionMethodStructFields(
	toPlain bool,
	printf func(format string, args ...any),
	ts *dst.TypeSpec,
	node *typegraph.Node,
	rawTyName, plainTyName string,
	importMap imports.ImportMap,
	exprMap map[string]fieldAstExprSet,
) {
	printf(`return %s{
`,
		or(toPlain, plainTyName, rawTyName),
	)
	defer printf(`}
`)
	edgeMap := node.ChildEdgeMap(isUndPlainAllowedEdge)
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

				edge, typeVar, tag, ok := edgeMap.ByFieldName(field.Names[0].Name)
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

					ty := edge.PrintChildArgConverted(plainConverter, importMap)
					fieldConverter, needsArg = generateConversionMethodDirect(toPlain, edge, undOpt, ty, importMap)
				} else if isUndConversionImplementor(edge.ChildType) || edge.IsChildMatched() {
					isPointer := edge.LastPointer().IsSomeAnd(func(tdep typegraph.EdgeRouteNode) bool {
						return tdep.Kind == typegraph.EdgeKindPointer
					})
					fieldConverter = _generateConversionMethodInvocationExpr(
						toPlain,
						isPointer,
						prefixPointer(isPointer, edge.PrintChildType(importMap)),
						astutil.PrintAstExprPanicking(plainExpr.Unwrapped),
					)
					needsArg = true
				} else if ok, isPointer := edge.HasSingleNamedTypeArg(nil); ok && isUndType(edge.ChildType) {
					_, ok := plainConverter(edge.TypeArgs[0].Ty, edge.IsTypeArgMatched())
					if ok {
						fieldConverter, needsArg = _generateConversionMethodImplementorMapper(
							toPlain,
							edge,
							edge.PrintChildArg(0, importMap),
							edge.PrintChildArgConverted(plainConverter, importMap),
							importMap,
							isPointer,
							func(ident string) string { return ident },
						)
					}
				}

				rawExpr := astutil.TypeToAst(
					edge.ParentNode.Type.Underlying().(*types.Struct).Field(edge.Stack[0].Pos.Value()).Type(),
					edge.ParentNode.Type.Obj().Pkg().Path(),
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

func generateConversionMethodDirect(toPlain bool, edge typegraph.Edge, undOpt undtag.UndOpt, typeParam string, importMap imports.ImportMap) (convert func(ident string) string, needsArg bool) {
	matchUndTypeBool(
		namedTypeToTargetType(edge.ChildType),
		false,
		func() {
			convert, needsArg = or(
				toPlain,
				func() (func(ident string) string, bool) { return optionToPlain(undOpt) },
				func() (func(ident string) string, bool) { return optionToRaw(undOpt, typeParam, importMap) },
			)()
		},
		func(isSlice bool) {
			convert, needsArg = or(
				toPlain,
				func() (func(ident string) string, bool) { return undToPlain(undOpt, importMap) },
				func() (func(ident string) string, bool) { return undToRaw(isSlice, undOpt, typeParam, importMap) },
			)()
		},
		func(isSlice bool) {
			convert, needsArg = or(
				toPlain,
				func() (func(ident string) string, bool) { return elasticToPlain(isSlice, undOpt, typeParam, importMap) },
				func() (func(ident string) string, bool) { return elasticToRaw(isSlice, undOpt, typeParam, importMap) },
			)()
		},
	)

	if ok, isPointer := edge.HasSingleNamedTypeArg(nil); ok {
		var isMatched bool
		if node := edge.TypeArgs[0].Node; node != nil {
			isMatched = (node.Matched &^ typegraph.MatchKindExternal) > 0
		}
		_, ok := plainConverter(edge.TypeArgs[0].Ty, isMatched)
		if ok {
			convert, needsArg = _generateConversionMethodImplementorMapper(
				toPlain,
				edge,
				edge.PrintChildArg(0, importMap),
				typeParam,
				importMap,
				isPointer,
				convert,
			)
		}
	}

	return
}

func _generateConversionMethodImplementorMapper(
	toPlain bool,
	edge typegraph.Edge,
	rawType, plainTy string,
	importMap imports.ImportMap,
	isPointer bool,
	inner func(ident string) string,
) (func(ident string) string, bool) {
	pkgIdent := importIdent(namedTypeToTargetType(edge.ChildType), importMap)

	return or(
		toPlain,
		func(ident string) string {
			return inner(fmt.Sprintf(
				`%s.Map(
						%s,
						func(v %s) %s {
							%s vv := v.UndPlain()
							%s
						},
					)`,
				pkgIdent, ident,
				rawType, plainTy,
				or(isPointer, "if v == nil { return nil }\n", ""),
				or(isPointer, "return &vv", "return vv"),
			))
		},
		func(ident string) string {
			return fmt.Sprintf(
				`%s.Map(
						%s,
						func(v %s) %s {
							%s vv := v.UndRaw()
							%s
						},
					)`,
				pkgIdent, inner(ident),
				plainTy, rawType,
				or(isPointer, "if v == nil { return nil }\n", ""),
				or(isPointer, "return &vv", "return vv"),
			)
		},
	), true
}

func _generateConversionMethodInvocationExpr(
	toPlain bool, isPointer bool,
	rawTy, plainTy string,
) func(expr string) string {
	if !isPointer {
		return func(expr string) string {
			return expr + or(toPlain, ".UndPlain()", ".UndRaw()")
		}
	} else {
		return func(expr string) string {
			return fmt.Sprintf(
				`func(v %s) %s {
					if v == nil {
						return nil
					}
					vv := v.%s()
					return &vv
				}(%s)`,
				or(toPlain, rawTy, plainTy),
				or(toPlain, plainTy, rawTy),
				or(toPlain, "UndPlain", "UndRaw"),
				expr,
			)
		}
	}
}
