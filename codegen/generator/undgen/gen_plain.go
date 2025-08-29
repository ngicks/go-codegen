package undgen

import (
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
	"github.com/ngicks/go-codegen/codegen/codegen"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-codegen/codegen/internal/bufpool"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-codegen/codegen/typegraph"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
)

func GeneratePlain(
	sourcePrinter *suffixwriter.Writer,
	verbose bool,
	pkgs []*packages.Package,
	extra []imports.TargetImport,
) error {
	parser := imports.NewParserPackages(pkgs)
	parser.AppendExtra(extra...)
	replacerData, err := gatherPlainUndTypes(
		pkgs,
		parser,
		isUndPlainAllowedEdge,
		func(g *typegraph.Graph) iter.Seq2[typegraph.Ident, *typegraph.Node] {
			return g.IterUpward(true, isUndPlainAllowedEdge)
		},
	)
	if err != nil {
		return err
	}

	buf := bufpool.GetBuf()
	defer bufpool.PutBuf(buf)

	for _, data := range xiter.Filter2(
		func(f *ast.File, data *typegraph.ReplaceData) bool { return f != nil && data != nil },
		hiter.MapsKeys(replacerData, pkgsutil.EnumerateFile(pkgs)),
	) {
		buf.Reset()

		slog.Debug(
			"found",
			slog.String("filename", data.Filename),
		)

		modified := hiter.Collect2(xiter.Filter2(
			func(node *typegraph.Node, exprMap map[string]fieldDstExprSet) bool {
				return node != nil && exprMap != nil
			},
			hiter.Divide(
				func(node *typegraph.Node) (*typegraph.Node, map[string]fieldDstExprSet) {
					exprMap, ok := _replaceToPlainTypes(data, node)
					if !ok {
						return nil, nil
					}
					slog.Debug(
						"rewritten",
						slog.String("package", node.Type.Obj().Pkg().Path()),
						slog.String("type", node.Type.Obj().Name()),
					)
					return node, exprMap
				},
				slices.Values(data.TargetNodes),
			),
		))

		if len(modified) == 0 {
			continue
		}

		data.ImportMap.AddMissingImports(data.DstFile)
		res := decorator.NewRestorer()
		af, err := res.RestoreFile(data.DstFile)
		if err != nil {
			return fmt.Errorf("converting dst to ast for %q: %w", data.Filename, err)
		}

		if err := codegen.PrintFileHeader(buf, af, res.Fset); err != nil {
			return fmt.Errorf("%q: %w", data.Filename, err)
		}

		for node, exprMap := range hiter.Values2(modified) {
			dts := data.Dec.Dst.Nodes[node.Ts].(*dst.TypeSpec)
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

			buf.WriteString("//" + codegen.DirectivePrefix + codegen.DirectiveCommentGenerated + "\n")
			buf.WriteString(token.TYPE.String())
			buf.WriteByte(' ')
			err = printer.Fprint(buf, res.Fset, ats)
			if err != nil {
				return fmt.Errorf("print.Fprint failed for type %s in file %q: %w", data.Filename, ats.Name.Name, err)
			}
			buf.WriteString("\n\n")

			err = generateConversionMethod(buf, data, node, astExprMap)
			if err != nil {
				return err
			}

			buf.WriteString("\n\n")
		}

		err = sourcePrinter.Write(context.Background(), data.Filename, buf.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
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

func plainConverter(ty *types.Named, isMatched bool) (*types.Named, bool) {
	if ty == nil {
		return nil, false
	}
	if !isMatched {
		return ConstUnd.ConversionMethod.ConvertedType(ty)
	}
	return makeRenamedType(
		ty,
		ty.Obj().Name()+"Plain",
		ty.Obj().Pkg(),
		func(typeName *types.TypeName) []*types.Func {
			return []*types.Func{
				types.NewFunc(
					0,
					ty.Obj().Pkg(),
					"UndRaw",
					types.NewSignatureType(
						types.NewVar(
							0,
							ty.Obj().Pkg(),
							"v",
							typeName.Type(),
						),
						nil,
						nil,
						nil,
						types.NewTuple(
							types.NewVar(
								0,
								ty.Obj().Pkg(),
								"",
								ty,
							),
						),
						false,
					),
				),
			}
		},
	), true
}
