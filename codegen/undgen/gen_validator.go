package undgen

import (
	"bufio"
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
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/und/undtag"
	"golang.org/x/tools/go/packages"
)

//go:generate go run ../ undgen validator -v --pkg ./testdata/targettypes/ --pkg ./testdata/targettypes/sub --pkg ./testdata/targettypes/sub2

func GenerateValidator(
	sourcePrinter *suffixwriter.Writer,
	verbose bool,
	pkgs []*packages.Package,
	imports []TargetImport,
) error {
	rawTypes, err := findValidatableTypes(pkgs, imports)
	if err != nil {
		return err
	}
	for data, err := range generatorRawIter(imports, rawTypes) {
		if err != nil {
			return err
		}

		if verbose {
			slog.Debug(
				"found",
				slog.String("filename", data.filename),
			)
		}

		res := decorator.NewRestorer()
		af, err := res.RestoreFile(data.df)
		if err != nil {
			return fmt.Errorf("converting dst to ast for %q: %w", data.filename, err)
		}

		buf := new(bytes.Buffer) // pool buf?

		buf.WriteString(token.PACKAGE.String())
		buf.WriteByte(' ')
		buf.WriteString(af.Name.Name)
		buf.WriteString("\n\n")

		// TODO: split these lines to function.
		for i, dec := range af.Decls {
			genDecl, ok := dec.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok != token.IMPORT {
				// it's possible that the file has multiple import spec.
				// but it always starts with import spec.
				break
			}
			err = printer.Fprint(buf, res.Fset, genDecl)
			if err != nil {
				return fmt.Errorf("print.Fprint failed printing %dth import spec in file %q: %w", i, data.filename, err)
			}
			buf.WriteString("\n\n")
		}

		var atLeastOne bool
		for _, matchedType := range data.targets {
			atLeastOne = true
			err := generateUndValidate(
				buf,
				data.dec.Dst.Nodes[matchedType.TypeSpec].(*dst.TypeSpec),
				matchedType,
				data.importMap,
			)
			if err != nil {
				return fmt.Errorf("generating UndValidate for type %s in file %q: %w", data.filename, matchedType.TypeSpec.Name.Name, err)
			}
			buf.WriteString("\n\n")
		}

		if atLeastOne {
			err = sourcePrinter.Write(context.Background(), data.filename, buf.Bytes())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// generates methods on the patch type
//
//	func (v Ty[T, U,...]) UndValidate(v OrgType[T, U,...]) {
//		if err := undtag.UndOpt{}.Validate(v.Field); err != nil  {
//			return err
//		}
//		return nil
//	}
func generateUndValidate(
	w io.Writer,
	ts *dst.TypeSpec,
	matchedFields RawMatchedType,
	imports importDecls,
) error {
	typeName := ts.Name.Name + printTypeParamVars(ts)
	undtagImportIdent, _ := imports.Ident(UndPathUndTag)
	optionImportIdent, _ := imports.Ident(UndTargetTypeOption.ImportPath)

	bufw := bufio.NewWriter(w)

	printf := func(format string, args ...any) {
		fmt.Fprintf(bufw, format, args...)
	}

	printf("//%s%s\n", UndDirectivePrefix, UndDirectiveCommentGenerated)
	printf("func (v %s) UndValidate() error {\n", typeName)
	switch matchedFields.Variant {
	case MatchedAsArray, MatchedAsSlice, MatchedAsMap:
		switch x := matchedFields.Field[0]; x.As {
		default:
			// do return an error
		case MatchedAsArray, MatchedAsSlice, MatchedAsMap:
			printf("for _, val := range v {\n")
			switch x.Type {
			case UndTargetTypeElastic, UndTargetTypeSliceElastic:
				printf("if !validator.ValidElastic(val)")
			case UndTargetTypeUnd, UndTargetTypeSliceUnd:
				printf("if !validator.ValidUnd(val)")
			case UndTargetTypeOption:
				printf("if !validator.ValidOpt(val)")
			}
			printf("{" +
				"\treturn fmt.Errorf(\"nay\")\n" +
				"}")
			printf("}\n")
		}
	case MatchedAsStruct:
		for _, f := range matchedFields.Field {
			if f.Tag.IsSome() {
				printf("{\n")
				printf("validator := %s\n\n", printValidator(undtagImportIdent, optionImportIdent, f.Tag.Value().Opt))
				switch f.Type {
				default:
					switch f.As {
					default:
						// do return an error
					case MatchedAsArray, MatchedAsSlice, MatchedAsMap:
						printf("for _, val := range v.%s {\n", f.Name)
						switch f.Elem.Type {
						case UndTargetTypeElastic, UndTargetTypeSliceElastic:
							printf("if !validator.ValidElastic(val)", f.Name)
						case UndTargetTypeUnd, UndTargetTypeSliceUnd:
							printf("if !validator.ValidUnd(val)", f.Name)
						case UndTargetTypeOption:
							printf("if !validator.ValidOpt(val)", f.Name)
						}
						printf("{" +
							"\treturn fmt.Errorf(\"nay\")\n" +
							"}")
						printf("}\n")
					}
				case UndTargetTypeElastic, UndTargetTypeSliceElastic:
					printf("if !validator.ValidElastic(v.%s)", f.Name)
				case UndTargetTypeUnd, UndTargetTypeSliceUnd:
					printf("if !validator.ValidUnd(v.%s)", f.Name)
				case UndTargetTypeOption:
					printf("if !validator.ValidOpt(v.%s)", f.Name)
				}
				printf("{\n" +
					`return fmt.Errorf("yay")` + "\n" +
					"}\n",
				)
				printf("}\n")
			} else {
				if f.As == MatchedAsImplementor {
					printf("if err := v.%s.UndValidate(); err != nil {\n", f.Name)
					printf("\treturn err\n")
					printf("}\n")
				}
			}
		}
	}
	printf("\nreturn nil\n")
	printf("}")

	return bufw.Flush()
}

type rawTypeReplacerData struct {
	filename  string
	af        *ast.File
	dec       *decorator.Decorator
	df        *dst.File
	importMap importDecls
	targets   iter.Seq2[int, RawMatchedType]
}

func generatorRawIter(imports []TargetImport, types RawTypes) iter.Seq2[rawTypeReplacerData, error] {
	return func(yield func(rawTypeReplacerData, error) bool) {
		for pkg, seq := range types.Iter() {
			for file, seq := range seq {
				dec := decorator.NewDecorator(pkg.Fset)
				df, err := dec.DecorateFile(file)
				if err != nil {
					if !yield(rawTypeReplacerData{}, err) {
						return
					}
					continue
				}

				importMap := parseImports(file.Imports, imports)

				addMissingImports(df, importMap)

				if !yield(
					rawTypeReplacerData{
						filename:  pkg.Fset.Position(file.FileStart).Filename,
						af:        file,
						dec:       dec,
						df:        df,
						importMap: importMap,
						targets:   seq,
					},
					nil,
				) {
					return
				}
			}
		}
	}
}

func findValidatableTypes(pkgs []*packages.Package, imports []TargetImport) (RawTypes, error) {
	validatorMethod := ValidatorMethod{"UndValidate"}
	// 1st path, find other than implementor
	matched, err := findRawTypes(pkgs, imports, validatorMethod, nil)
	if err != nil {
		return matched, err
	}
	// 2nd path, find including implementor
	matched, err = findRawTypes(pkgs, imports, validatorMethod, matched)
	if err != nil {
		return matched, err
	}
	return matched, nil
}

type ValidatorMethod struct {
	Name string
}

func (method ValidatorMethod) IsImplementor(ty *types.Named) bool {
	return isValidatorImplementor(ty, method.Name)
}

func isValidatorImplementor(ty *types.Named, methodName string) bool {
	ms := types.NewMethodSet(ty)
	for i := range ms.Len() {
		sel := ms.At(i)
		if sel.Obj().Name() == methodName {
			sig, ok := sel.Obj().Type().Underlying().(*types.Signature)
			if !ok {
				return false
			}
			tup := sig.Results()
			if tup.Len() != 1 {
				return false
			}
			v := tup.At(0)

			named, ok := v.Type().(*types.Named)
			if !ok {
				return false
			}
			return named.Obj().Name() == "error"
		}
	}
	return false
}

func printValidator(undtagImportIdent, optionImportIdent string, tagOpt undtag.UndOpt) string {
	var builder strings.Builder
	builder.WriteString(undtagImportIdent + ".UndOpt{\n")
	if tagOpt.States.IsSome() {
		s := tagOpt.States.Value()
		builder.WriteString(fmt.Sprintf("\tStates: %s.Some(%s.StateValidator{\n", optionImportIdent, undtagImportIdent))
		if s.Def {
			builder.WriteString("\t\tDef: true,\n")
		}
		if s.Null {
			builder.WriteString("\t\tNull: true,\n")
		}
		if s.Und {
			builder.WriteString("\t\tUnd: true,\n")
		}
		builder.WriteString("\t}),\n")
	}
	if tagOpt.Len.IsSome() {
		l := tagOpt.Len.Value()
		builder.WriteString(fmt.Sprintf("\tLen: %s.Some(%s.LenValidator{\n", optionImportIdent, undtagImportIdent))
		builder.WriteString(fmt.Sprintf("\t\tLen: %d,\n", l.Len))
		op := ""
		switch l.Op {
		case undtag.LenOpEqEq: // ==
			op = "LenOpEqEq"
		case undtag.LenOpGr: // >
			op = "LenOpGr"
		case undtag.LenOpGrEq: // >=
			op = "LenOpGrEq"
		case undtag.LenOpLe: // <
			op = "LenOpLe"
		case undtag.LenOpLeEq: // <=
			op = "LenOpLeEq"
		}
		if op != "" {
			builder.WriteString(fmt.Sprintf("\t\tOp: %s.%s,\n", undtagImportIdent, op))
		}
		builder.WriteString("\t}),\n")
	}
	if tagOpt.Values.IsSome() {
		v := tagOpt.Values.Value()
		builder.WriteString(fmt.Sprintf("\tValues: %s.Some(%s.ValuesValidator{\n", optionImportIdent, undtagImportIdent))
		if v.Nonnull {
			builder.WriteString("\t\tNonnull: true,\n")
		}
		builder.WriteString("\t}),\n")
	}
	builder.WriteString("}")
	return builder.String()
}
