package cloner

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ngicks/go-codegen/codegen/codegen"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/typegraph"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/hiter/stringsiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
)

var (
	// errIgnored    = errors.New("ignored")
	errFieldNotOk = errors.New("field not ok") // TODO: will this really happen?
	errParamNotOk = errors.New("param not ok")
	errNotHandled = errors.New("not handled")
)

func generateMethod(
	c *Config,
	w io.Writer,
	g *typegraph.Graph,
	node *typegraph.Node,
	replacer *typegraph.ReplaceData,
) (err error) {
	ats := node.Ts

	buf := new(bytes.Buffer)
	printf, flush := codegen.BufPrintf(buf)

	err = generateCloner(c, printf, g, replacer.ImportMap, node, ats)
	if err != nil {
		return err
	}
	err = flush()
	if err != nil {
		return err
	}
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func generateCloner(
	c *Config,
	printf func(format string, args ...any),
	g *typegraph.Graph,
	importMap imports.ImportMap,
	node *typegraph.Node,
	ats *ast.TypeSpec,
) (err error) {
	typeName := node.Ts.Name.Name + codegen.PrintTypeParamsAst(node.Ts)

	var cloneCallbacks [][2]string

	printf("//" + codegen.DirectivePrefix + codegen.DirectiveCommentGenerated + "\n")
	if node.Type.TypeParams().Len() == 0 {
		printf("func (v %[1]s) Clone() %[1]s {\n", typeName)
	} else {
		// [][2]string{{"cloneT","T"}}
		cloneCallbacks = gatherCloneCallback(node.Type.TypeParams())

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
	default:
		return errNotHandled
	case *types.Struct:
		aStruct := ats.Type.(*ast.StructType)
		printf("return %s{\n", typeName)
		defer printf("}\n")
		var handled int
		for af := range codegen.FieldAst(aStruct) {
			i := af.Pos

			edge, _, _, _ := edges.ByFieldPos(i)

			_, _, handleKind, _ := c.matcherConfig().handleField(
				i,
				node,
				edge.ChildNode,
				node.Type.Underlying().(*types.Struct).Field(i).Type(),
			)

			if handleKind == handleKindIgnore {
				continue
			}
			handled++

			clonerExpr, callable, err := cloneTy(
				c,
				node.Type.Obj().Pkg().Path(),
				importMap,
				g,
				node,
				edge.ChildNode,
				i,
				af.Field.Type,
				node.Type.Underlying().(*types.Struct).Field(i).Type(),
				cloneCallbacks,
			)
			switch {
			case errors.Is(err, errParamNotOk):
				return err
			}

			printf(
				"%s:",
				af.Name,
			)
			if callable {
				printf(strings.ReplaceAll(clonerExpr("v."+af.Name)+"("+"v."+af.Name+")", "%", "%%"))
			} else {
				printf(strings.ReplaceAll(clonerExpr("v."+af.Name), "%", "%%"))
			}
			printf(",\n")
		}
		if handled == 0 {
			err = errNotHandled
		}
	case *types.Array, *types.Slice, *types.Map:
		_, edge, _ := edges.First()

		ty := node.Type.Underlying()
		_, _, handleKind, _ := c.matcherConfig().handleField(
			-1,
			node,
			edge.ChildNode,
			ty,
		)

		if handleKind == handleKindIgnore {
			return errNotHandled
		}

		clonerExpr, callable, err := cloneTy(
			c,
			node.Type.Obj().Pkg().Path(),
			importMap,
			g,
			node,
			edge.ChildNode,
			-1,
			ats.Type,
			ty,
			cloneCallbacks,
		)
		switch {
		case errors.Is(err, errParamNotOk):
			return err
		}

		printf("return ")
		if callable {
			printf(strings.ReplaceAll(clonerExpr("v")+"(v)", "%", "%%"))
		} else {
			printf(strings.ReplaceAll(clonerExpr("v"), "%", "%%"))
		}
		printf("\n")
	}
	return
}

func gatherCloneCallback(tyParams *types.TypeParamList) [][2]string {
	return slices.Collect(
		xiter.Map(
			func(p *types.TypeParam) [2]string {
				name := p.Obj().Name()
				first, size := utf8.DecodeRuneInString(name)
				return [2]string{
					"clone" + string(unicode.ToUpper(first)) + name[size:],
					name,
				}
			},
			hiter.OmitF(hiter.AtterAll(tyParams)),
		),
	)
}

func cloneTy(
	c *Config,
	pkgPath string,
	importMap imports.ImportMap,
	g *typegraph.Graph,
	parent *typegraph.Node,
	child *typegraph.Node,
	pos int,
	expr ast.Expr,
	ty types.Type,
	cloneCallbacks [][2]string,
) (clonerExpr func(s string) string, callable bool, err error) {
	unwrapped, stack, handleKind, idx := c.matcherConfig().handleField(
		pos,
		parent,
		child,
		ty,
	)

	if handleKind == handleKindIgnore {
		return nil, false, errNotHandled
	}

	if len(stack) > 0 && stack[0].Kind == typegraph.EdgeKindStruct {
		stack = stack[1:]
	}
	if idx := slices.IndexFunc(
		stack,
		func(node typegraph.EdgeRouteNode) bool {
			// There can be 2 or more aliases
			// But after that, there should not be other than that
			// since aliasing only points to named type.
			return node.Kind == typegraph.EdgeKindAlias
		}); idx >= 0 {
		stack = stack[:idx]
	}

	unwrappedExpr, unwrapper := unwrapFieldAlongPath(
		// af.Field.Type, af.Field.Type,
		expr, expr,
		stack,
		0,
	)

	var cloneExpr func(s string) string
	switch handleKind {
	default:
		panic(fmt.Errorf("unknown kind: %d", handleKind))
	case handleKindAssign:
		cloneExpr = func(s string) string { return s }
	case handleKindNewChannel:
		cloneExpr = func(s string) string {
			return fmt.Sprintf("make(%s, cap(%s))", codegen.PrintAstExprPanicking(unwrappedExpr), s)
		}
	case handleKindCallCb:
		cloneExpr = func(s string) string {
			return fmt.Sprintf("%s(%s)", cloneCallbacks[unwrapped.(*types.TypeParam).Index()][0], s)
		}
	case handleKindCallClone:
		cloneExpr = func(s string) string { return s + ".Clone()" }
	case handleKindCallCloneFunc:
		builder := strings.Builder{}
		builder.WriteString(".CloneFunc(\n")
		// always instantiated
		for i, t := range hiter.AtterAll(types.Unalias(unwrapped).(*types.Named).TypeArgs()) {
			switch x := t.(type) {
			case *types.TypeParam:
				builder.WriteString(cloneCallbacks[x.Index()][0])
			default:
				var childTy types.Type
				_ = typegraph.TraverseTypes(
					x,
					nil,
					func(ty types.Type, named *types.Named, stack []typegraph.EdgeRouteNode) error {
						childTy = ty
						return nil
					},
					nil,
				)
				child, _ := g.GetByType(childTy)

				var cbs [][2]string
				named, ok := types.Unalias(x).(*types.Named)
				if ok {
					cbs = gatherCloneCallback(named.TypeParams())
				}

				expr, callable, err := cloneTy(
					c,
					pkgPath,
					importMap,
					g,
					nil,
					child,
					-1,
					codegen.TypeToAst(x, pkgPath, importMap),
					x,
					cbs,
				)

				if err != nil {
					return nil, false, fmt.Errorf("%w: type param at index %d: %w", errParamNotOk, i, err)
				}

				if callable {
					builder.WriteString(expr("v"))
				} else {
					builder.WriteString(
						fmt.Sprintf(
							`func (v %[1]s) %[1]s {
								return %[2]s
							}`,
							codegen.PrintAstExprPanicking(codegen.TypeToAst(x, pkgPath, importMap)), expr("v"),
						),
					)
				}
			}
			builder.WriteString(",\n")
		}
		builder.WriteString(")")

		cloneExpr = func(s string) string {
			return s + builder.String()
		}
	case handleKindUseCustomHandler:
		cloneExpr, callable = c.matcherConfig().
			CustomHandlers[idx].
			Expr(CustomHandlerExprData{
				ImportMap: importMap,
				AstExpr:   unwrappedExpr,
				Ty:        unwrapped,
			})
	case handleKindStructLiteral:
		callable = true
		builder := strings.Builder{}
		printf, flush := codegen.BufPrintf(&builder)
		printf(
			`func (v %[1]s) %[1]s {
	return %[1]s{
`,
			codegen.PrintAstExprPanicking(unwrappedExpr),
		)
		for i, f := range pkgsutil.EnumerateFields(unwrapped.(*types.Struct)) {
			var childTy types.Type
			_ = typegraph.TraverseTypes(
				f.Type(),
				func(ty types.Type, currentStack []typegraph.EdgeRouteNode) bool {
					_, isStructLit := ty.(*types.Struct)
					return isStructLit
				},
				func(ty types.Type, named *types.Named, stack []typegraph.EdgeRouteNode) error {
					childTy = ty
					return nil
				},
				nil,
			)
			child, _ := g.GetByType(childTy)

			expr, callable, err := cloneTy(
				c,
				pkgPath,
				importMap,
				g,
				nil,
				child,
				-1,
				codegen.TypeToAst(f.Type(), pkgPath, importMap),
				f.Type(),
				nil,
			)

			if err != nil {
				return nil, false, fmt.Errorf("%w: field of struct literal at index %d: %w", errFieldNotOk, i, err)
			}

			printf(
				"%s:",
				f.Name(),
			)
			if callable {
				printf(strings.ReplaceAll(expr("v."+f.Name())+"("+"v."+f.Name()+")", "%", "%%"))
			} else {
				printf(strings.ReplaceAll(expr("v."+f.Name()), "%", "%%"))
			}
			printf(",\n")
		}
		printf("}\n}")

		if err := flush(); err != nil {
			panic(err)
		}

		cloneExpr = func(s string) string {
			return builder.String()
		}
	}

	if unwrapper != nil {
		if callable {
			inner := cloneExpr
			cloneExpr = func(s string) string {
				return inner("") + "(" + s + ")"
			}
		}
		return func(s string) string { return unwrapper(cloneExpr) }, true, nil
	} else {
		return cloneExpr, callable, nil
	}
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
) (unwrapped ast.Expr, unwrapper func(wrappee func(string) string) string) {
	if fromExpr == nil || toExpr == nil {
		return toExpr, nil
	}
	input := codegen.PrintAstExprPanicking(fromExpr)
	output := codegen.PrintAstExprPanicking(toExpr)

	s := stack[skip:]
	if len(s) == 0 {
		return toExpr, nil
	}

	initializer := func(expr ast.Expr, kind typegraph.EdgeKind, variable string) (s string) {
		defer func() {
			if s == "" {
				return
			}
			s = "if v != nil {\n" + variable + "=" + s + "\n}"
		}()
		switch kind {
		default: // case typegraph.EdgeKindArray, typegraph.EdgeKindPointer:
			return ""
		case typegraph.EdgeKindSlice:
			return fmt.Sprintf("make(%s, len(v), cap(v))", codegen.PrintAstExprPanicking(expr))
		case typegraph.EdgeKindMap:
			return fmt.Sprintf("make(%s, len(v))", codegen.PrintAstExprPanicking(expr))
		}
	}

	var wrappers []func(string) string
	unwrapped = toExpr
	for p := range hiter.Window(s, 2) {
		unwrapped = unwrapExprOne(unwrapped, p[0].Kind)
		tyExpr := codegen.PrintAstExprPanicking(unwrapped)
		initializerExpr := initializer(unwrapped, p[1].Kind, "inner")
		switch p[0].Kind {
		case typegraph.EdgeKindPointer:
			wrappers = append(wrappers, func(s string) string {
				return fmt.Sprintf(
					`if v != nil {
						v := *v
						outer := &inner
						var inner %s
						%s
						%s
						(*outer) = &inner
					}`,
					tyExpr, initializerExpr, s,
				)
			})
		case typegraph.EdgeKindArray, typegraph.EdgeKindMap, typegraph.EdgeKindSlice:
			wrappers = append(wrappers, func(s string) string {
				return fmt.Sprintf(
					`for k, v := range v {
						outer := &inner
						var inner %s
						%s
						%s
						(*outer)[k] = inner
					}`,
					tyExpr, initializerExpr, s,
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
						vv := %s
						inner = &vv
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

	unwrapped = unwrapExprOne(unwrapped, s[len(s)-1].Kind)
	return unwrapped, func(wrappee func(string) string) string {
		wrappers = slices.Insert(wrappers, 0, func(expr string) string {
			return fmt.Sprintf(
				`func (v %[1]s) %[2]s {
					var out %[1]s

					%[3]s

					inner := out
					%s
					out = inner

					return out
				}`,
				input, output, initializer(toExpr, s[0].Kind, "out"), expr)
		})
		expr := wrappee("v")
		for _, wrapper := range slices.Backward(wrappers) {
			expr = wrapper(expr)
		}
		return expr
	}
}
