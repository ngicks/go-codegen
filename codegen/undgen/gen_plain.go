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
	"maps"
	"slices"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
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

		modified := hiter.Collect2(xiter.Filter2(
			func(node *typeNode, exprMap map[string]fieldDstExprSet) bool {
				return node != nil && exprMap != nil
			},
			hiter.Divide(
				func(node *typeNode) (*typeNode, map[string]fieldDstExprSet) {
					exprMap, ok := _replaceToPlainTypes(data, node)
					if !ok {
						return nil, nil
					}
					return node, exprMap
				},
				slices.Values(data.targetNodes),
			),
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

		for node, exprMap := range hiter.Values2(modified) {
			dts := data.dec.Dst.Nodes[node.ts].(*dst.TypeSpec)
			ats := res.Ast.Nodes[dts].(*ast.TypeSpec)

			astExprMap := maps.Collect(
				xiter.Map2(
					func(s string, expr fieldDstExprSet) (string, fieldAstExprSet) {
						return s, fieldAstExprSet{
							Wrapped:   res.Ast.Nodes[expr.Wrapped].(ast.Expr),
							Unwrapped: res.Ast.Nodes[expr.Unwrapped].(ast.Expr),
						}
					},
					maps.All(exprMap),
				),
			)

			buf.WriteString("//" + UndDirectivePrefix + UndDirectiveCommentGenerated + "\n")
			buf.WriteString(token.TYPE.String())
			buf.WriteByte(' ')
			err = printer.Fprint(buf, res.Fset, ats)
			if err != nil {
				return fmt.Errorf("print.Fprint failed for type %s in file %q: %w", data.filename, ats.Name.Name, err)
			}
			buf.WriteString("\n\n")

			err = generateMethodToPlain(buf, data, node, astExprMap)
			if err != nil {
				return err
			}

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
