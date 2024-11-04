package undgen

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"iter"
	"log/slog"
	"slices"
	"strconv"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"github.com/ngicks/und/undtag"
	"golang.org/x/tools/go/packages"
)

//go:generate go run ../ undgen plain --pkg ./internal/targettypes/ --pkg ./internal/targettypes/sub --pkg ./internal/targettypes/sub2
//go:generate go run ../ undgen plain --pkg ./internal/patchtarget/...
//go:generate go run ../ undgen plain --pkg ./internal/validatortarget/...
//go:generate go run ../ undgen plain --pkg ./internal/plaintarget/...

func GeneratePlain(
	sourcePrinter *suffixwriter.Writer,
	verbose bool,
	pkgs []*packages.Package,
	imports []TargetImport,
) error {
	replacerData, err := gatherPlainUndTypes(
		pkgs,
		imports,
		isUndPlainAllowedEdge,
		func(g *typeGraph) iter.Seq2[typeIdent, *typeNode] {
			return g.iterUpward(true, isUndPlainAllowedEdge)
		},
	)
	if err != nil {
		return err
	}

	for _, data := range xiter.Filter2(
		func(f *ast.File, data *replaceData) bool { return f != nil && data != nil },
		hiter.MapKeys(replacerData, enumerateFile(pkgs)),
	) {
		if verbose {
			slog.Debug(
				"found",
				slog.String("filename", data.filename),
			)
		}

		modified := slices.Collect(xiter.Filter(
			func(node *typeNode) bool {
				return _replaceToPlainTypes(data, node)
			},
			slices.Values(data.targetNodes),
		))

		if len(modified) == 0 {
			continue
		}

		res := decorator.NewRestorer()
		af, err := res.RestoreFile(data.df)
		if err != nil {
			return fmt.Errorf("converting dst to ast for %q: %w", data.filename, err)
		}

		buf := new(bytes.Buffer) // pool buf?

		_ = printPackage(buf, af)
		err = printImport(buf, af, res.Fset)
		if err != nil {
			return fmt.Errorf("%q: %w", data.filename, err)
		}

		for _, node := range modified {
			dts := data.dec.Dst.Nodes[node.ts].(*dst.TypeSpec)
			ats := res.Ast.Nodes[dts].(*ast.TypeSpec)

			buf.WriteString("//" + UndDirectivePrefix + UndDirectiveCommentGenerated + "\n")
			buf.WriteString(token.TYPE.String())
			buf.WriteByte(' ')
			err = printer.Fprint(buf, res.Fset, ats)
			if err != nil {
				return fmt.Errorf("print.Fprint failed for type %s in file %q: %w", data.filename, ats.Name.Name, err)
			}
			buf.WriteString("\n\n")

			// err = generateMethodToPlain(
			// 	buf,
			// 	data.dec,
			// 	dts,
			// 	ats.Name.Name[:len(ats.Name.Name)-len("Plain")]+printTypeParamVars(dts),
			// 	ats.Name.Name+printTypeParamVars(dts),
			// 	s,
			// 	data.importMap,
			// 	data.rawFields[idx],
			// 	data.plainFields[idx],
			// )
			// if err != nil {
			// 	return err
			// }

			// buf.WriteString("\n\n")

			// err = generateMethodToRaw(
			// 	buf,
			// 	data.dec,
			// 	dts,
			// 	ats.Name.Name[:len(ats.Name.Name)-len("Plain")]+printTypeParamVars(dts),
			// 	ats.Name.Name+printTypeParamVars(dts),
			// 	s,
			// 	data.importMap,
			// 	data.rawFields[idx],
			// 	data.plainFields[idx],
			// )
			// if err != nil {
			// 	return err
			// }

			buf.WriteString("\n\n")
		}

		err = sourcePrinter.Write(context.Background(), data.filename, buf.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func _replaceToPlainTypes(data *replaceData, node *typeNode) bool {
	ts := data.dec.Dst.Nodes[node.ts].(*dst.TypeSpec)
	ts.Name.Name += "Plain"
	named := node.typeInfo
	switch named.Underlying().(type) {
	case *types.Array, *types.Slice, *types.Map:
		unwrapElemTypes(ts, node, data.importMap)
		return true
	case *types.Struct:
		return unwrapStructFields(ts, node, data.importMap)
	}
	return false
}

func unwrapPath(expr *dst.Expr, edge typeDependencyEdge, skip int) *dst.Expr {
	unwrapped := expr
	for _, p := range edge.stack[skip:] {
		switch p.kind {
		case typeDependencyEdgeKindArray, typeDependencyEdgeKindSlice:
			next := (*unwrapped).(*dst.ArrayType)
			unwrapped = &next.Elt
		case typeDependencyEdgeKindMap:
			next := (*unwrapped).(*dst.MapType)
			unwrapped = &next.Value
		}
	}
	return unwrapped
}

func unwrapElemTypes(ts *dst.TypeSpec, node *typeNode, importMap importDecls) {
	var elem *dst.Expr
	switch x := ts.Type.(type) {
	case *dst.ArrayType: // slice or array. difference is Len expr.
		elem = &x.Elt
	case *dst.MapType:
		elem = &x.Value
	}
	// should be only one since we prohibit struct literals.
	_, edge := firstTypeIdent(node.children)
	if isUndType(edge.childType) {
		// matched, wrapped implementor
		unwrapped := unwrapPath(elem, edge, 1)
		index := (*unwrapped).(*dst.IndexExpr)
		converted, _ := ConstUnd.ConversionMethod.ConvertedType(edge.typeArgs[0].ty)
		index.Index = typeToDst(
			converted,
			node.typeInfo.Obj().Pkg().Path(),
			importMap,
		)
	} else {
		// implementor
		converted, _ := ConstUnd.ConversionMethod.ConvertedType(edge.childType)
		ts.Type = typeToDst(
			converted,
			node.typeInfo.Obj().Pkg().Path(),
			importMap,
		)
	}
}

func unwrapStructFields(ts *dst.TypeSpec, node *typeNode, importMap importDecls) bool {
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

				edge, _, tag, ok := node.byFieldName(field.Names[0].Name)
				if !ok {
					// not found
					return false
				}

				unwrapped := unwrapPath(&field.Type, edge, 1)

				undTagValue, hasTag := tag.Lookup(undtag.TagName)
				// edge.childNode.typeInfo.
				if hasTag {
					undOpt, err := undtag.ParseOption(undTagValue)
					if err != nil { // This case should be filtered when forming the graph.
						panic(err)
					}
					expr, modified := unwrapUndType((*unwrapped).(*dst.IndexExpr), edge, undOpt, importMap)
					if modified {
						atLeastOne = true
						*unwrapped = expr
					}
					return false
				}

				if named := edge.childType; ConstUnd.ConversionMethod.IsImplementor(named) {
					converted, _ := ConstUnd.ConversionMethod.ConvertedType(named)
					*unwrapped = typeToDst(
						converted,
						edge.parentNode.typeInfo.Obj().Pkg().Path(),
						importMap,
					)
					atLeastOne = true
				}

				return false
			}
		},
		nil,
	)
	return atLeastOne
}

func unwrapUndType(fieldTy *dst.IndexExpr, edge typeDependencyEdge, undOpt undtag.UndOpt, importMap importDecls) (expr dst.Expr, modified bool) {
	modified = true

	// default: unchanged.
	// maybe below lines writes expr entirely.
	expr = fieldTy

	// fieldTy -> X.Sel[Index]
	sel := fieldTy.X.(*dst.SelectorExpr) // X.Sel

	if edge.hasSingleNamedTypeArg(isUndConversionImplementor) {
		arg := edge.typeArgs[0].ty
		named, _ := ConstUnd.ConversionMethod.ConvertedType(arg)
		fieldTy.Index = typeToDst(
			named,
			edge.parentNode.typeInfo.Obj().Pkg().Path(),
			importMap,
		)
	}

	_ = matchUndTypeBool(
		namedTypeToTargetType(edge.childType),
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

func conversionTargetOfImplementorAst(target RawMatchedType, fieldTypeNamed *types.Named, importMap importDecls) ast.Expr {
	ty, ok := ConstUnd.ConversionMethod.ConvertedType(fieldTypeNamed)
	if ok {
		return typeToAst(
			ty,
			target.TypeInfo.Type().(*types.Named).Obj().Pkg().Path(),
			importMap,
		)
	} else {
		return typeToAst(
			types.NewNamed(
				types.NewTypeName(
					0,
					fieldTypeNamed.Obj().Pkg(),
					fieldTypeNamed.Obj().Name()+"Plain",
					nil,
				),
				nil,
				nil,
			),
			fieldTypeNamed.Obj().Pkg().Path(),
			importMap,
		)
	}

}

func sliceSuffix(isSlice bool) string {
	if isSlice {
		return "Slice"
	}
	return ""
}

func suffixSlice(s string, isSlice bool) string {
	if isSlice {
		s += "Slice"
	}
	return s
}
