package undgen

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"iter"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/go-codegen/codegen/structtag"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
)

//go:generate go run ../ undgen patch --pkg ./internal/patchtarget All Ignored Hmm NameOverlapping
//go:generate go run ../ undgen patch --pkg ./internal/targettypes All WithTypeParam A B IncludesSubTarget

func GeneratePatcher(
	sourcePrinter *suffixwriter.Writer,
	verbose bool,
	pkg *packages.Package,
	imports []TargetImport,
	targetTypeNames ...string,
) error {
	if verbose {
		slog.Debug(
			"target type names",
			slog.Any("names", targetTypeNames),
		)
	}

	replacerData, err := gatherPlainUndTypes(
		[]*packages.Package{pkg},
		imports,
		nil, // no transitive type marking; it is not needed here.
		func(g *typeGraph) iter.Seq2[typeIdent, *typeNode] {
			return g.enumerateTypesKeys(
				xiter.Map(func(s string) typeIdent {
					return typeIdent{pkgPath: pkg.PkgPath, typeName: s}
				},
					slices.Values(targetTypeNames),
				),
			)
		},
	)
	if err != nil {
		return err
	}

	for _, data := range xiter.Filter2(
		func(f *ast.File, data *replaceData) bool { return f != nil && data != nil },
		hiter.MapKeys(replacerData, slices.Values(pkg.Syntax)),
	) {
		wrapNonUndFields(data)

		if verbose {
			slog.Debug(
				"found",
				slog.String("filename", data.filename),
				slog.Any(
					"typesNames",
					slices.Collect(xiter.Map(
						func(n *typeNode) string { return n.typeInfo.Obj().Name() },
						slices.Values(data.targetNodes),
					)),
				),
			)
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

		for _, node := range data.targetNodes {
			dts := data.dec.Dst.Nodes[node.ts].(*dst.TypeSpec)
			ts := res.Ast.Nodes[dts].(*ast.TypeSpec)
			// type keyword is attached to *ast.GenDecl
			// But we are not printing gen decl itself since
			// it could have multiple specs inside it (type (spec1; spec2;...))
			// surely at least a spec of them is converted but we can't tell all of them were.
			buf.WriteString("//" + UndDirectivePrefix + UndDirectiveCommentGenerated + "\n")
			buf.WriteString(token.TYPE.String())
			buf.WriteByte(' ')
			err = printer.Fprint(buf, res.Fset, ts)
			if err != nil {
				return fmt.Errorf("print.Fprint failed for type %s in file %q: %w", data.filename, ts.Name.Name, err)
			}
			buf.WriteString("\n\n")

			for _, gen := range []methodGenSet{
				{
					generateFromValue,
					func() error {
						return fmt.Errorf("generating FromValue for type %s in file %q: %w", data.filename, ts.Name.Name, err)
					},
				},
				{
					generateToValue,
					func() error {
						return fmt.Errorf("generating ToValue for type %s in file %q: %w", data.filename, ts.Name.Name, err)
					},
				},
				{
					generateMerge,
					func() error {
						return fmt.Errorf("generating Merge for type %s in file %q: %w", data.filename, ts.Name.Name, err)
					},
				},
				{
					generateApplyPatch,
					func() error {
						return fmt.Errorf("generating ApplyPatch for type %s in file %q: %w", data.filename, ts.Name.Name, err)
					},
				},
			} {
				err = gen.fn(
					buf,
					dts,
					node,
					data.importMap,
					"Patch",
				)
				if err != nil {
					return gen.errFunc()
				}
			}
		}
		err = sourcePrinter.Write(context.Background(), data.filename, buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

type methodGenSet struct {
	fn      methodGenFunc
	errFunc func() error
}

type methodGenFunc func(w io.Writer, ts *dst.TypeSpec, node *typeNode, imports importDecls, typeSuffix string) error

func wrapNonUndFields(data *replaceData) {
	for _, node := range data.targetNodes {
		wrapNonUndFieldsWithSliceUnd(data.dec.Dst.Nodes[node.ts].(*dst.TypeSpec), node, data.importMap)
	}
	addMissingImports(data.df, data.importMap)
}

func wrapNonUndFieldsWithSliceUnd(ts *dst.TypeSpec, target *typeNode, importMap importDecls) {
	typeName := ts.Name.Name
	ts.Name.Name = ts.Name.Name + "Patch"
	edges := edgesDirectFields(target)
	dstutil.Apply(
		ts.Type,
		func(c *dstutil.Cursor) bool {
			node := c.Node()
			switch field := node.(type) {
			default:
				return true
			case *dst.Field:
				if len(field.Names) == 0 {
					return false
				}

				edge, ok := edges[field.Names[0].Name]

				if field.Tag == nil {
					field.Tag = &dst.BasicLit{}
				}
				tag := field.Tag

				isSliceType := true
				if ok {
					matchUndType(
						namedTypeToTargetType(edge.childType),
						false,
						func() bool {
							c.Replace(&dst.Field{
								Names: field.Names,
								Type: &dst.IndexExpr{
									X:     importMap.DstExpr(UndTargetTypeSliceUnd),
									Index: field.Type.(*dst.IndexExpr).Index,
								},
								Tag:  field.Tag,
								Decs: field.Decs,
							})
							return true
						},
						func(isSlice bool) bool {
							isSliceType = isSlice
							return true
						},
						func(isSlice bool) bool {
							isSliceType = isSlice
							return true
						},
					)
				} else {
					c.Replace(
						&dst.Field{
							Names: field.Names,
							Type: &dst.IndexExpr{
								X:     importMap.DstExpr(UndTargetTypeSliceUnd),
								Index: field.Type,
							},
							Tag:  field.Tag,
							Decs: field.Decs,
						},
					)
				}
				tagOpt, err := structtag.ParseStructTag(
					reflect.StructTag(unquoteBasicLitString(tag.Value)),
				)
				if err != nil {
					panic(fmt.Errorf(
						"malformed struct tag on field %s of type %q: %w",
						concatFieldNames(field), typeName, err,
					))
				}
				tagOpt, _ = tagOpt.Delete("json", "omitempty")
				tagOpt, _ = tagOpt.Delete("json", "omitzero")
				omitOpt := "omitempty"
				if !isSliceType {
					omitOpt = "omitzero"
				}
				tagOpt, _ = tagOpt.Add("json", omitOpt, "")
				tag.Value = "`" + string(tagOpt.StructTag()) + "`"
				return false
			}
		},
		nil,
	)
}

func edgesDirectFields(node *typeNode) map[string]typeDependencyEdge {
	return maps.Collect(xiter.Filter2(
		func(_ string, edge typeDependencyEdge) bool {
			return len(edge.stack) == 1 && edge.stack[0].kind == typeDependencyEdgeKindStruct
		},
		node.fieldsName(),
	))
}

func concatFieldNames(field *dst.Field) string {
	return hiter.StringsCollect(
		0,
		hiter.SkipLast(
			1,
			hiter.Decorate(
				nil,
				hiter.WrapSeqIterable(hiter.Once(",")),
				xiter.Map(
					func(i *dst.Ident) string { return strconv.Quote(i.Name) },
					slices.Values(field.Names),
				),
			),
		),
	)
}

// addMissingImports adds missing imports from imports to df,
// both [*dst.File.Imports] and tge first import decl in [*dst.File.Decls].
func addMissingImports(df *dst.File, imports importDecls) {
	var replaced bool
	dstutil.Apply(
		df,
		func(c *dstutil.Cursor) bool {
			if replaced {
				return false
			}
			node := c.Node()
			switch x := node.(type) {
			default:
				return true
			case *dst.GenDecl:
				if x.Tok != token.IMPORT {
					return false
				}
				for ident, path := range imports.MissingImports() {
					spec := &dst.ImportSpec{
						Name: dst.NewIdent(ident),
						Path: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(path)},
					}
					if ident == "" {
						spec.Name = nil
					}
					df.Imports = append(df.Imports, spec)
					x.Specs = append(x.Specs, spec)
				}
				replaced = true
				return false
			}
		},
		nil,
	)
}

func printTypeParamVars(ts *dst.TypeSpec) string {
	if ts.TypeParams == nil {
		return ""
	}
	// appends [TypeParams, ...] without type constraint to type names.
	// Uses same _FieldName_ for type param vars for some sort of pretty printing.
	var typeParams strings.Builder
	for _, f := range ts.TypeParams.List {
		if typeParams.Len() > 0 {
			typeParams.WriteByte(',')
		}
		typeParams.WriteString(f.Names[0].Name)
	}
	return "[" + typeParams.String() + "]"
}

func typeObjectFieldsIter(typeInfo types.Type) iter.Seq2[int, *types.Var] {
	return func(yield func(int, *types.Var) bool) {
		structTy, ok := typeInfo.Underlying().(*types.Struct)
		if !ok {
			return
		}
		for i := range structTy.NumFields() {
			if !yield(i, structTy.Field(i)) {
				return
			}
		}
	}
}

// generates methods on the patch type
//
//	func (p *Patch[T, U,...]) FromValue(v OrgType[T, U,...]) {
//		*p = Patch[T, U,...]{
//			fields: Conversion(v.fields),
//			// ...
//		}
//	}
func generateFromValue(
	w io.Writer, ts *dst.TypeSpec, node *typeNode, imports importDecls, typeSuffix string,
) (err error) {
	patchTypeName := ts.Name.Name + printTypeParamVars(ts)
	orgTypeName := strings.TrimSuffix(ts.Name.Name, typeSuffix) + printTypeParamVars(ts)

	printf, flush := bufPrintf(w)
	defer func() {
		err = flush()
	}()

	printf("//%s%s\n", UndDirectivePrefix, UndDirectiveCommentGenerated)
	printf("func (p *%s) FromValue(v %s) {\n", patchTypeName, orgTypeName)
	// shut up linter. sometimes linter warns you should directly convert type to type using T(u).
	// It is possible that the patch type is exactly same as org type.
	printf("\t//nolint\n")
	printf("\t*p = %s{\n", patchTypeName)

	edges := edgesDirectFields(node)
	for _, f := range typeObjectFieldsIter(node.typeInfo) {
		printf("\t\t")
		// There's 3 possible conversions.
		// T -> sliceund.Und[T]
		// option.Option[T] -> sliceund.Und[T]
		// conserve type other than that e.g. for und.Und, elastic.Elastic.
		edge, ok := edges[f.Name()]
		if !ok || !matchUndType(
			namedTypeToTargetType(edge.childType),
			false,
			func() bool {
				// convert option -> und
				t := f.Type().(*types.Named).TypeArgs().At(0).String()
				sliceUndImportIdent, _ := imports.Ident(UndTargetTypeSliceUnd.ImportPath)
				optionImportIdent, _ := imports.Ident(UndTargetTypeOption.ImportPath)
				printf(
					"%[1]s: %[2]s.MapOr("+
						"v.%[1]s,"+
						" %[3]s.Null[%[4]s](),"+
						"%[3]s.Defined[%[4]s]),\n",
					f.Name(), optionImportIdent, sliceUndImportIdent, t,
				)
				return true
			},
			func(isSlice bool) bool {
				printf("%[1]s: v.%[1]s,\n", f.Name())
				return true
			},
			func(isSlice bool) bool {
				printf("%[1]s: v.%[1]s,\n", f.Name())
				return true
			},
		) {
			// T -> sliceund.Und[T]
			sliceUndImportIdent, _ := imports.Ident(UndTargetTypeSliceUnd.ImportPath)
			printf("%[1]s: %[2]s.Defined(v.%[1]s),\n", f.Name(), sliceUndImportIdent)
		}
	}
	printf("\t}\n")
	printf("}\n\n")

	return
}

// generates methods on the patch type
//
//	func (p Patch[T, U,...]) ToValue() OrgType[T, U,...] {
//		return OrgType[T, U,...]{
//			fields: Conversion(p.fields),
//			// ...
//		}
//	}
func generateToValue(
	w io.Writer, ts *dst.TypeSpec, node *typeNode, imports importDecls, typeSuffix string,
) (err error) {
	patchTypeName := ts.Name.Name + printTypeParamVars(ts)
	orgTypeName := strings.TrimSuffix(ts.Name.Name, typeSuffix) + printTypeParamVars(ts)

	printf, flush := bufPrintf(w)
	defer func() {
		err = flush()
	}()

	printf("//%s%s\n", UndDirectivePrefix, UndDirectiveCommentGenerated)
	printf("func (p %s) ToValue() %s {\n", patchTypeName, orgTypeName)
	// Same as FromValue, shut up linter. always explicitly note type params.
	printf("\t//nolint\n")
	printf("\treturn %s{\n", orgTypeName)

	edges := edgesDirectFields(node)
	for _, f := range typeObjectFieldsIter(node.typeInfo) {
		printf("\t\t")
		edge, ok := edges[f.Name()]
		// Like FromValue, there's 3 possible back-conversions.
		// sliceund.Und[T] -> T
		// sliceund.Und[T] -> option.Option[T]
		// conserve type other than that e.g. for und.Und, elastic.Elastic.
		if !ok || !matchUndType(
			namedTypeToTargetType(edge.childType),
			false,
			func() bool {
				// sliceund.Und[T] -> option.Option[T]
				optionImportIdent, _ := imports.Ident(UndTargetTypeOption.ImportPath)
				printf("%[1]s: %[2]s.Flatten(p.%[1]s.Unwrap()),\n", f.Name(), optionImportIdent)
				return true
			},
			func(isSlice bool) bool {
				printf("%[1]s: p.%[1]s,\n", f.Name())
				return true
			},
			func(isSlice bool) bool {
				printf("%[1]s: p.%[1]s,\n", f.Name())
				return true
			},
		) {
			// sliceund.Und[T] -> T
			printf("%[1]s: p.%[1]s.Value(),\n", f.Name())
		}
	}
	printf("\t}\n")
	printf("}\n\n")

	return
}

// generates methods on the patch type
//
//	func (p Patch[T, U,...]) Merge(r Patch[T, U,...]) Patch[T, U,...] {
//		return Patch[T, U,...]{
//			fields: Or(p.fields, r.fields),
//			// ...
//		}
//	}
func generateMerge(
	w io.Writer, ts *dst.TypeSpec, node *typeNode, imports importDecls, _ string,
) (err error) {
	patchTypeName := ts.Name.Name + printTypeParamVars(ts)

	printf, flush := bufPrintf(w)
	defer func() {
		err = flush()
	}()

	printf("//%s%s\n", UndDirectivePrefix, UndDirectiveCommentGenerated)
	printf("func (p %[1]s) Merge(r %[1]s) %[1]s {\n", patchTypeName)
	// Same as FromValue, shut up linter. always explicitly note type params.
	printf("\t//nolint\n")
	printf("\treturn %s{\n", patchTypeName)
	edges := edgesDirectFields(node)
	for _, f := range typeObjectFieldsIter(node.typeInfo) {
		printf("\t\t")
		edge, ok := edges[f.Name()]
		// Like FromValue, there's 2 possible Or logic.
		// both und like type.
		// both elastic like type.
		undImportIdent, _ := imports.Ident(UndTargetTypeSliceUnd.ImportPath)
		if !ok || !matchUndType(
			namedTypeToTargetType(edge.childType),
			false,
			func() bool {
				return false
			},
			func(isSlice bool) bool {
				if !isSlice {
					undImportIdent, _ = imports.Ident(UndTargetTypeUnd.ImportPath)
				}
				return false
			},
			func(isSlice bool) bool {
				elasticImportIdent, _ := imports.Ident(UndTargetTypeSliceElastic.ImportPath)
				undImportIdent, _ = imports.Ident(UndTargetTypeSliceUnd.ImportPath)
				if !isSlice {
					elasticImportIdent, _ = imports.Ident(UndTargetTypeElastic.ImportPath)
					undImportIdent, _ = imports.Ident(UndTargetTypeUnd.ImportPath)
				}
				// or(elastic, elastic)
				printf(
					"%[1]s: %[2]s.FromUnd(%[3]s.FromOption(r.%[1]s.Unwrap().Unwrap().Or(p.%[1]s.Unwrap().Unwrap()))),\n",
					f.Name(), elasticImportIdent, undImportIdent,
				)
				return true
			},
		) {
			// or(und,und)
			printf(
				"%[1]s: %[2]s.FromOption(r.%[1]s.Unwrap().Or(p.%[1]s.Unwrap())),\n",
				f.Name(), undImportIdent,
			)
		}
	}
	printf("\t}\n")
	printf("}\n\n")

	return
}

// generates methods on the patch type
//
//	func (p Patch[T, U,...]) ApplyPatch(v OrgType[T, U,...]) OrgType[T, U,...] {
//		var orgP Patch[T, U,...]
//		orgP.FromValue(v)
//		merged := orgP.Merge(p)
//		return merged.ToValue()
//	}
func generateApplyPatch(
	w io.Writer, ts *dst.TypeSpec, _ *typeNode, _ importDecls, typeSuffix string,
) (err error) {
	patchTypeName := ts.Name.Name + printTypeParamVars(ts)
	orgTypeName := strings.TrimSuffix(ts.Name.Name, typeSuffix) + printTypeParamVars(ts)

	printf, flush := bufPrintf(w)
	defer func() {
		err = flush()
	}()

	printf("//%s%s\n", UndDirectivePrefix, UndDirectiveCommentGenerated) // note this is generated method.
	printf("func (p %[1]s) ApplyPatch(v %[2]s) %[2]s {\n", patchTypeName, orgTypeName)
	printf("\tvar orgP %s\n", patchTypeName)
	printf("\torgP.FromValue(v)\n")
	printf("\tmerged := orgP.Merge(p)\n")
	printf("\treturn merged.ToValue()\n")
	printf("}\n\n")

	return
}
