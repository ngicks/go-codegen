package cloner

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dave/dst"
	"github.com/ngicks/go-codegen/codegen/codegen"
	"github.com/ngicks/go-codegen/codegen/typegraph"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/hiter/stringsiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
)

func generateMethod(
	c *Config,
	w io.Writer,
	node *typegraph.Node,
	replacer *typegraph.ReplaceData,
) (err error) {
	ats := node.Ts
	dts := replacer.Dec.Dst.Nodes[ats].(*dst.TypeSpec)

	printf, flush := codegen.BufPrintf(w)
	defer func() {
		fErr := flush()
		if err == nil {
			err = fErr
		}
	}()

	err = generateCloner(c, printf, node, ats, dts)
	return
}

func generateCloner(
	c *Config,
	printf func(format string, args ...any),
	node *typegraph.Node,
	ats *ast.TypeSpec,
	dts *dst.TypeSpec,
) error {
	typeName := node.Ts.Name.Name + codegen.PrintTypeParamsAst(node.Ts)

	var cloneCallbacks [][2]string

	if node.Type.TypeParams().Len() == 0 {
		printf("func (v %[1]s) Clone() %[1]s {\n", typeName)
	} else {
		cloneCallbacks = slices.Collect(
			xiter.Map(
				func(p *types.TypeParam) [2]string {
					name := p.Obj().Name()
					first, size := utf8.DecodeRuneInString(name)
					return [2]string{
						"clone" + string(unicode.ToUpper(first)) + name[size:],
						name,
					}
				},
				hiter.OmitF(hiter.AtterAll(node.Type.TypeParams())),
			),
		)

		printf(
			"func (v %[1]s) CloneFunc(%[2]s) %[1]s {\n",
			typeName,
			stringsiter.Join(
				1024,
				", ",
				xiter.Map(
					func(s [2]string) string {
						return fmt.Sprintf("%[1]s func(%[2]s) %[2]s", s[0], s[1])
					},
					slices.Values(cloneCallbacks),
				),
			),
		)
	}
	defer printf("}\n\n")

	edges := node.ChildEdgeMap(c.MatcherConfig.MatchEdge)

	switch node.Type.Underlying().(type) {
	case *types.Struct:
		aStruct := ats.Type.(*ast.StructType)
		dStruct := dts.Type.(*dst.StructType)
		printf("return %s{\n", typeName)
		defer printf("}\n")
		for af, _ := range hiter.Pairs(codegen.FieldAst(aStruct), codegen.FieldDst(dStruct)) {
			i := af.Pos

			edge, _, _, _ := edges.ByFieldPos(i)

			unwrapped, stack, handleKind := c.matcherConfig().handleField(
				i,
				node,
				edge,
				node.Type.Underlying().(*types.Struct).Field(i).Type(),
			)

			if len(stack) > 0 && stack[0].Kind == typegraph.EdgeKindStruct {
				stack = stack[1:]
			}

			if handleKind == handleKindIgnore {
				continue
			}

			// box each field. allow defer to wrap it up.
			func() {
				printf(
					"%s:",
					af.Name,
				)
				defer printf(",\n")

				unwrapper := unwrapFieldAlongPath(
					af.Field.Type, af.Field.Type,
					stack,
					0,
				)

				var cloneExpr func(s string) string
				switch handleKind {
				default:
					panic(fmt.Errorf("unknown kind: %d", handleKind))
				case handleKindAssign:
					cloneExpr = func(s string) string { return s }
				case handleKindCallCb:
					cloneExpr = func(s string) string {
						return fmt.Sprintf("%s(%s)", cloneCallbacks[unwrapped.(*types.TypeParam).Index()][0], s)
					}
				case handleKindCallClone:
					cloneExpr = func(s string) string { return s + ".Clone()" }
				case handleKindCallCloneFunc:
					// handle cases
					cloneExpr = func(s string) string { return s + "CloneFunc()" }
				}

				if unwrapper != nil {
					printf(unwrapper(cloneExpr, "v."+af.Name))
				} else {
					printf(strings.ReplaceAll(cloneExpr("v."+af.Name), "%", "%%"))
				}
			}()
		}
	}
	return nil
}

func unwrapExprOne(expr ast.Expr, kind typegraph.EdgeKind) ast.Expr {
	switch kind {
	case typegraph.EdgeKindArray:
		return expr.(*ast.ArrayType).Elt
	case typegraph.EdgeKindMap:
		return expr.(*ast.MapType).Value
	case typegraph.EdgeKindPointer:
		return expr.(*ast.StarExpr).X
	case typegraph.EdgeKindSlice:
		return expr.(*ast.ArrayType).Elt
	}
	return expr
}

func unwrapFieldAlongPath(
	fromExpr, toExpr ast.Expr,
	stack []typegraph.EdgeRouteNode,
	skip int,
) func(wrappee func(string) string, fieldExpr string) string {
	if fromExpr == nil || toExpr == nil {
		return nil
	}
	input := codegen.PrintAstExprPanicking(fromExpr)
	output := codegen.PrintAstExprPanicking(toExpr)

	s := stack[skip:]
	if len(s) == 0 {
		return nil
	}

	initializer := func(expr ast.Expr, kind typegraph.EdgeKind) string {
		switch kind {
		case typegraph.EdgeKindPointer:
			return fmt.Sprintf("new(%s)", codegen.PrintAstExprPanicking(expr.(*ast.StarExpr).X))
		case typegraph.EdgeKindArray:
			return fmt.Sprintf("%s{}", codegen.PrintAstExprPanicking(expr))
		default:
			return fmt.Sprintf("make(%s, len(v))", codegen.PrintAstExprPanicking(expr))
		}
	}

	var wrappers []func(string) string
	unwrapped := toExpr
	for p := range hiter.Window(s, 2) {
		unwrapped = unwrapExprOne(unwrapped, p[0].Kind)
		initializerExpr := initializer(unwrapped, p[1].Kind)
		switch p[0].Kind {
		case typegraph.EdgeKindPointer:
			wrappers = append(wrappers, func(s string) string {
				return fmt.Sprintf(
					`if v != nil {
						v := *v
						outer := &inner
						inner := %s
						%s
						(*outer) = &inner
					}`,
					initializerExpr, s,
				)
			})
		case typegraph.EdgeKindArray, typegraph.EdgeKindMap, typegraph.EdgeKindSlice:
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
	}

	// inner most
	switch s[len(s)-1].Kind {
	case typegraph.EdgeKindPointer:
		wrappers = append(wrappers, func(s string) string {
			return fmt.Sprintf(
				`if v != nil {
						v := *v
						*inner = %s
					}`,
				s,
			)
		})
	case typegraph.EdgeKindArray, typegraph.EdgeKindMap, typegraph.EdgeKindSlice:
		wrappers = append(wrappers, func(s string) string {
			return fmt.Sprintf(
				`for k, v := range v {
				inner[k] = %s
			}`,
				s,
			)
		})
	}

	return func(wrappee func(string) string, fieldExpr string) string {
		wrappers = slices.Insert(wrappers, 0, func(expr string) string {
			return fmt.Sprintf(
				`func (v %s) %s {
					out := %s

					inner := out
					%s

					return out
				}(%s)`,
				input, output, initializer(toExpr, s[0].Kind), expr, fieldExpr)
		})
		expr := wrappee("v")
		for _, wrapper := range slices.Backward(wrappers) {
			expr = wrapper(expr)
		}
		return expr
	}
}
