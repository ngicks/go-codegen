package undgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"log/slog"
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

func generateConversionMethod(w io.Writer, data *replaceData, node *typeNode, exprMap map[string]fieldAstExprSet) (err error) {
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

	generateToRawOrToPlain(true, printf, plainTyName, rawTyName, ts, data, node, exprMap)
	generateToRawOrToPlain(false, printf, plainTyName, rawTyName, ts, data, node, exprMap)

	return
}

func generateToRawOrToPlain(
	toPlain bool,
	printf func(format string, args ...any),
	plainTyName, rawTyName string,
	ts *dst.TypeSpec,
	data *replaceData,
	node *typeNode,
	exprMap map[string]fieldAstExprSet,
) {
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

	named := node.typeInfo
	switch named.Underlying().(type) {
	case *types.Array, *types.Slice, *types.Map:
		generateConversionMethodElemTypes(toPlain, printf, node, data.importMap, exprMap)
	case *types.Struct:
		generateConversionMethodStructFields(toPlain, printf, ts, node, rawTyName, plainTyName, data.importMap, exprMap)
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
	node *typeNode,
	importMap importDecls,
	exprMap map[string]fieldAstExprSet,
) {
	conversionIndent, _ := importMap.Ident(UndPathConversion)

	_, edge := firstTypeIdent(node.children) // must be only one.

	rawExpr := typeToAst(
		edge.parentNode.typeInfo.Underlying(),
		edge.parentNode.typeInfo.Obj().Pkg().Path(),
		importMap,
	)

	var plainExpr ast.Expr
	for _, v := range exprMap {
		plainExpr = v.Wrapped
	}

	unwrapper := unwrapFieldAlongPath(
		or(toPlain, rawExpr, plainExpr),
		or(toPlain, plainExpr, rawExpr),
		edge,
		0,
	)

	if isUndType(edge.childType) {
		// matched, wrapped implementor
		printf(`return ` + unwrapper(
			func(s string) string {
				return fmt.Sprintf(
					`%s.Map(
						%s,
						%s.%s,
					)`,
					importIdent(namedTypeToTargetType(edge.childType), importMap),
					s,
					conversionIndent,
					or(toPlain, "ToPlain", "ToRaw"),
				)
			},
			"v",
		) + `
`)
		return
	} else {
		// implementor
		printf(`return ` + unwrapper(
			func(s string) string {
				return fmt.Sprintf(
					`%s.%s()`,
					s,
					or(toPlain, "UndPlain", "UndRaw"),
				)
			},
			"v",
		) + `
`)
		return
	}

}

func generateConversionMethodStructFields(
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

func generateConversionMethodDirect(toPlain bool, edge typeDependencyEdge, undOpt undtag.UndOpt, typeParam string, importMap importDecls) (convert func(ident string) string, needsArg bool) {
	matchUndTypeBool(
		namedTypeToTargetType(edge.childType),
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

	if edge.hasSingleNamedTypeArg(isUndConversionImplementor) {
		conversionIdent, _ := importMap.Ident(UndPathConversion)
		pkgIdent := importIdent(namedTypeToTargetType(edge.childType), importMap)
		inner := convert
		convert = or(
			toPlain,
			func(ident string) string {
				return inner(fmt.Sprintf(
					`%s.Map(
				%s,
				%s.%s,
			)`,
					pkgIdent, ident, conversionIdent, or(toPlain, "ToPlain", "ToRaw"),
				))
			},
			func(ident string) string {
				return fmt.Sprintf(
					`%s.Map(
				%s,
				%s.%s,
			)`,
					pkgIdent, inner(ident), conversionIdent, or(toPlain, "ToPlain", "ToRaw"),
				)
			},
		)
		needsArg = true
	}
	return
}
