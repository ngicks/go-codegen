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
	"github.com/ngicks/go-iterator-helper/hiter/iterable"
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
	for data, err := range xiter.Map2(
		wrapNonUndFields,
		generatorIter(
			imports,
			findTypes(pkg, targetTypeNames...),
		),
	) {
		if err != nil {
			return err
		}

		if len(data.targets) == 0 {
			continue
		}
		if verbose {
			slog.Debug(
				"found",
				slog.String("filename", data.filename),
				slog.Any("typesNames", slices.Collect(data.targets.typeNames())),
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

		for i, spec := range hiter.Enumerate(data.targets.typeSpecs()) {
			astSpec, ok := res.Ast.Nodes[spec]
			if !ok {
				panic(fmt.Errorf("implementation error: restored file does not contain type spec corresponding to %q", spec.Name.Name))
			}
			ts := astSpec.(*ast.TypeSpec)

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
					spec,
					data.targets[i].tsi.TypeInfo,
					data.targets[i].mt.Value(),
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

type methodGenFunc func(w io.Writer, ts *dst.TypeSpec, typeInfo types.Object, matchedFields RawMatchedType, imports importDecls, typeSuffix string) error

func wrapNonUndFields(data replaceData, err error) (replaceData, error) {
	if err != nil {
		return data, err
	}
	for i, ts := range hiter.Enumerate(data.targets.typeSpecs()) {
		wrapNonUndFieldsWithSliceUnd(ts, data.targets[i].replacerPerTypeData, data.importMap)
	}
	addMissingImports(data.df, data.importMap)

	return data, nil
}

func wrapNonUndFieldsWithSliceUnd(ts *dst.TypeSpec, target replacerPerTypeData, importMap importDecls) {
	typeName := ts.Name.Name
	ts.Name.Name = ts.Name.Name + "Patch"
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

				// later mutated
				// We need allocate one since field.Tag is nil when no tag is set.
				tag := &dst.BasicLit{}
				if field.Tag != nil {
					tag = field.Tag
				}
				field.Tag = tag

				isSliceType := true
				if f, ok := target.Field(field.Names[0].Name); ok && slices.Contains(UndTargetTypes, f.Type) {
					switch f.Type {
					case UndTargetTypeOption:
						c.Replace(&dst.Field{
							Names: field.Names,
							Type: &dst.IndexExpr{
								X:     importMap.DstExpr(UndTargetTypeSliceUnd),
								Index: field.Type.(*dst.IndexExpr).Index,
							},
							Tag:  field.Tag,
							Decs: field.Decs,
						})
					case UndTargetTypeUnd, UndTargetTypeElastic:
						isSliceType = false
					case UndTargetTypeSliceUnd, UndTargetTypeSliceElastic:
					}
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
				if tag != nil {
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
				}
				return false
			}
		},
		nil,
	)
}

func concatFieldNames(field *dst.Field) string {
	return hiter.StringsCollect(
		0,
		hiter.SkipLast(
			1,
			hiter.Decorate(
				nil,
				iterable.Repeatable[string]{V: ",", N: -1},
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

func typeObjectFieldsIter(typeInfo types.Object) iter.Seq2[int, *types.Var] {
	return func(yield func(int, *types.Var) bool) {
		structTy, ok := typeInfo.Type().Underlying().(*types.Struct)
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
	w io.Writer,
	ts *dst.TypeSpec,
	typeInfo types.Object,
	matchedFields RawMatchedType,
	imports importDecls,
	patcherTypeSuffix string,
) (err error) {
	patchTypeName := ts.Name.Name + printTypeParamVars(ts)
	orgTypeName := strings.TrimSuffix(ts.Name.Name, patcherTypeSuffix) + printTypeParamVars(ts)

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
	for i, f := range typeObjectFieldsIter(typeInfo) {
		printf("\t\t")
		// There's 3 possible conversions.
		// T -> sliceund.Und[T]
		// option.Option[T] -> sliceund.Und[T]
		// conserve type other than that e.g. for und.Und, elastic.Elastic.
		j := slices.IndexFunc(matchedFields.Field, func(mf MatchedField) bool { return mf.Pos == i })
		if j < 0 {
			// T -> sliceund.Und[T]
			sliceUndImportIdent, _ := imports.Ident(UndTargetTypeSliceUnd.ImportPath)
			printf("%[1]s: %[2]s.Defined(v.%[1]s),\n", f.Name(), sliceUndImportIdent)
			continue
		}
		undTypeField := matchedFields.Field[j]
		switch undTypeField.Type {
		case UndTargetTypeOption:
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
			continue
		case UndTargetTypeUnd, UndTargetTypeSliceUnd,
			UndTargetTypeElastic, UndTargetTypeSliceElastic:
			printf("%[1]s: v.%[1]s,\n", f.Name())
			continue
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
	w io.Writer,
	ts *dst.TypeSpec,
	typeInfo types.Object,
	matchedFields RawMatchedType,
	imports importDecls,
	patcherTypeSuffix string,
) (err error) {
	patchTypeName := ts.Name.Name + printTypeParamVars(ts)
	orgTypeName := strings.TrimSuffix(ts.Name.Name, patcherTypeSuffix) + printTypeParamVars(ts)

	printf, flush := bufPrintf(w)
	defer func() {
		err = flush()
	}()

	printf("//%s%s\n", UndDirectivePrefix, UndDirectiveCommentGenerated)
	printf("func (p %s) ToValue() %s {\n", patchTypeName, orgTypeName)
	// Same as FromValue, shut up linter. always explicitly note type params.
	printf("\t//nolint\n")
	printf("\treturn %s{\n", orgTypeName)
	for i, f := range typeObjectFieldsIter(typeInfo) {
		printf("\t\t")
		// Like FromValue, there's 3 possible back-conversions.
		// sliceund.Und[T] -> T
		// sliceund.Und[T] -> option.Option[T]
		// conserve type other than that e.g. for und.Und, elastic.Elastic.
		j := slices.IndexFunc(matchedFields.Field, func(mf MatchedField) bool { return mf.Pos == i })
		if j < 0 {
			// sliceund.Und[T] -> T
			printf("%[1]s: p.%[1]s.Value(),\n", f.Name())
			continue
		}
		undTypeField := matchedFields.Field[j]
		switch undTypeField.Type {
		case UndTargetTypeOption:
			// sliceund.Und[T] -> option.Option[T]
			optionImportIdent, _ := imports.Ident(UndTargetTypeOption.ImportPath)
			printf("%[1]s: %[2]s.Flatten(p.%[1]s.Unwrap()),\n", f.Name(), optionImportIdent)
			continue
		case UndTargetTypeUnd, UndTargetTypeSliceUnd,
			UndTargetTypeElastic, UndTargetTypeSliceElastic:
			printf("%[1]s: p.%[1]s,\n", f.Name())
			continue
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
	w io.Writer,
	ts *dst.TypeSpec,
	typeInfo types.Object,
	matchedFields RawMatchedType,
	imports importDecls,
	_ /*patcherTypeSuffix*/ string,
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
	for i, f := range typeObjectFieldsIter(typeInfo) {
		printf("\t\t")
		// Like FromValue, there's 2 possible Or logic.
		// both und like type.
		// both elastic like type.
		j := slices.IndexFunc(matchedFields.Field, func(mf MatchedField) bool { return mf.Pos == i })

		undImportIdent, _ := imports.Ident(UndTargetTypeSliceUnd.ImportPath)
		if j >= 0 {
			undTypeField := matchedFields.Field[j]
			switch undTypeField.Type {
			case UndTargetTypeUnd:
				undImportIdent, _ = imports.Ident(UndTargetTypeUnd.ImportPath)
			case UndTargetTypeElastic, UndTargetTypeSliceElastic:
				elasticImportIdent, _ := imports.Ident(UndTargetTypeElastic.ImportPath)
				undImportIdent, _ = imports.Ident(UndTargetTypeUnd.ImportPath)
				if undTypeField.Type == UndTargetTypeSliceElastic {
					elasticImportIdent, _ = imports.Ident(UndTargetTypeSliceElastic.ImportPath)
					undImportIdent, _ = imports.Ident(UndTargetTypeSliceUnd.ImportPath)
				}
				// or(elastic, elastic)
				printf(
					"%[1]s: %[2]s.FromUnd(%[3]s.FromOption(r.%[1]s.Unwrap().Unwrap().Or(p.%[1]s.Unwrap().Unwrap()))),\n",
					f.Name(), elasticImportIdent, undImportIdent,
				)
				continue
			}
		}
		// or(und,und)
		printf(
			"%[1]s: %[2]s.FromOption(r.%[1]s.Unwrap().Or(p.%[1]s.Unwrap())),\n",
			f.Name(), undImportIdent,
		)
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
	w io.Writer,
	ts *dst.TypeSpec,
	_ /*typeInfo*/ types.Object,
	_ /*matchedFields*/ RawMatchedType,
	_ /*imports*/ importDecls,
	patcherTypeSuffix string,
) (err error) {
	patchTypeName := ts.Name.Name + printTypeParamVars(ts)
	orgTypeName := strings.TrimSuffix(ts.Name.Name, patcherTypeSuffix) + printTypeParamVars(ts)

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
