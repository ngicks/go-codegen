package undgen

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"iter"
	"log/slog"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/und/undtag"
	"golang.org/x/tools/go/packages"
)

//go:generate go run ../ undgen validator --pkg ./internal/targettypes/ --pkg ./internal/targettypes/sub --pkg ./internal/targettypes/sub2
//go:generate go run ../ undgen validator --pkg ./internal/validatortarget/...

func GenerateValidator(
	sourcePrinter *suffixwriter.Writer,
	verbose bool,
	pkgs []*packages.Package,
	imports []TargetImport,
) error {
	imports = AppendTargetImports(imports, TargetImport{ImportPath: "fmt"})

	rawTypes, err := findValidatableTypes(pkgs, imports)
	if err != nil {
		return err
	}
	for data, err := range preprocessRawTypes(imports, rawTypes) {
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

		_ = printPackage(buf, af)
		err = printImport(buf, af, res.Fset)
		if err != nil {
			return fmt.Errorf("%q: %w", data.filename, err)
		}

		var atLeastOne bool
		for _, matchedType := range hiter.Values2(data.targets) {
			written, err := generateUndValidate(
				buf,
				data.dec.Dst.Nodes[matchedType.TypeSpec].(*dst.TypeSpec),
				matchedType,
				data.importMap,
			)
			if written {
				atLeastOne = true
			}
			if err != nil {
				return fmt.Errorf("generating UndValidate for type %s in file %q: %w", matchedType.TypeSpec.Name.Name, data.filename, err)
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
//		if err := undtag.UndOptExport{}.Into().ValidateOpt(v.Field); err != nil  {
//			return err
//		}
//		return nil
//	}
func generateUndValidate(
	w io.Writer,
	ts *dst.TypeSpec,
	matchedFields RawMatchedType,
	imports importDecls,
) (written bool, err error) {
	typeName := ts.Name.Name + printTypeParamVars(ts)
	undtagImportIdent, _ := imports.Ident(UndPathUndTag)
	validateImportIdent, _ := imports.Ident(UndPathValidate)

	buf := new(bytes.Buffer)

	printf, flush := bufPrintf(w)
	defer func() {
		fErr := flush()
		if err != nil {
			return
		}
		err = fErr
	}()

	// true only when validator is meaningful.
	var shouldPrint bool
	defer func() {
		if err != nil || !shouldPrint {
			return
		}
		written = true
		_, err = w.Write(buf.Bytes())
	}()

	printf("//%s%s\n", UndDirectivePrefix, UndDirectiveCommentGenerated)
	printf("func (v %s) UndValidate() error {\n", typeName)
	switch matchedFields.Variant {
	case MatchedAsArray, MatchedAsSlice, MatchedAsMap:
		f := matchedFields.Field[0]
		printf("for i, val := range v {\n")
		if ident := importIdent(f.Type, imports); ident != "" && f.Elem != nil && f.Elem.As == MatchedAsImplementor {
			shouldPrint = true
			printf(
				`if err := %s.UndValidate(val); err != nil {
									return %s.AppendValidationErrorIndex(
										err,
										fmt.Sprintf("%%v", i),
									)
								}
							`,
				ident, validateImportIdent,
			)
		}
		printf("}\n")
	case MatchedAsStruct:
		for _, f := range matchedFields.Field {
			if f.UndTag.IsSome() {
				shouldPrint = true
				printf("{\n")
				printf("validator := %s\n\n", printValidator(undtagImportIdent, f.UndTag.Value().Opt))
				switch f.Type {
				default:
					switch f.As {
					default:
						return false, fmt.Errorf("und struct tag on non eligible field")
					case MatchedAsArray, MatchedAsSlice, MatchedAsMap:
						printf("for i, val := range v.%s {\n", f.Name)
						switch f.Elem.Type {
						case UndTargetTypeElastic, UndTargetTypeSliceElastic:
							printf("if !validator.ValidElastic(val)")
						case UndTargetTypeUnd, UndTargetTypeSliceUnd:
							printf("if !validator.ValidUnd(val)")
						case UndTargetTypeOption:
							printf("if !validator.ValidOpt(val)")
						}
						printf(
							`{
								return %[1]s.AppendValidationErrorDot(
									%[1]s.AppendValidationErrorIndex(
										fmt.Errorf("%%s: value is %%s", validator.Describe(), %[1]s.ReportState(i)),
										fmt.Sprintf("%%v", i),
									),
									%[2]q,
								)
							}`,
							validateImportIdent, f.JsonFieldName(),
						)
						if ident := importIdent(f.Elem.Type, imports); ident != "" &&
							f.Elem != nil &&
							f.Elem.Elem != nil &&
							f.Elem.Elem.As == MatchedAsImplementor {
							printf(
								`
								if err := %[1]s.UndValidate(val); err != nil {
									return %[2]s.AppendValidationErrorDot(
										%[2]s.AppendValidationErrorIndex(
											err,
											fmt.Sprintf("%%v", i),
										),
										%[3]q,
									)
								}
							`,
								ident, validateImportIdent, f.JsonFieldName(),
							)
						}
						printf("}\n")
					}
				case UndTargetTypeElastic, UndTargetTypeSliceElastic, UndTargetTypeUnd, UndTargetTypeSliceUnd, UndTargetTypeOption:
					shouldPrint = true
					switch f.Type {
					case UndTargetTypeElastic, UndTargetTypeSliceElastic:
						printf("if !validator.ValidElastic(v.%s)", f.Name)
					case UndTargetTypeUnd, UndTargetTypeSliceUnd:
						printf("if !validator.ValidUnd(v.%s)", f.Name)
					case UndTargetTypeOption:
						printf("if !validator.ValidOpt(v.%s)", f.Name)
					}
					printf(
						`{
							return %[1]s.AppendValidationErrorDot(
								fmt.Errorf("%%s: value is %%s", validator.Describe(), %[1]s.ReportState(v.%[2]s)),
								%[3]q,
							)
						}
							`,
						validateImportIdent, f.Name, f.JsonFieldName(),
					)
					if f.Elem != nil && f.Elem.As == MatchedAsImplementor {
						if ident := importIdent(f.Type, imports); ident != "" {
							printf(
								`if err := %s.UndValidate(v.%s); err != nil {
									return %s.AppendValidationErrorDot(
										err,
										%q,
									)
								}
							`,
								ident, f.Name, validateImportIdent, f.JsonFieldName(),
							)
						}
					}
				}
				printf("}\n")
			} else {
				if f.As == MatchedAsImplementor {
					shouldPrint = true
					printf(
						`if err := v.%s.UndValidate(); err != nil {
							return %s.AppendValidationErrorDot(err,	%q)
						}
					`,
						f.Name, validateImportIdent, f.JsonFieldName(),
					)
				}
			}
		}
	}
	printf(`
	return nil
	}`)

	return
}

type rawTypeReplacerData struct {
	filename    string
	af          *ast.File
	dec         *decorator.Decorator
	df          *dst.File
	importMap   importDecls
	targets     hiter.KeyValues[int, RawMatchedType]
	rawFields   map[int]map[string]string
	plainFields map[int]map[string]string
}

func preprocessRawTypes(imports []TargetImport, rawTypes RawTypes) iter.Seq2[rawTypeReplacerData, error] {
	return func(yield func(rawTypeReplacerData, error) bool) {
		for pkg, seq := range rawTypes.Iter() {
			for file, seq := range seq {
				dec := decorator.NewDecorator(pkg.Fset)
				df, err := dec.DecorateFile(file)
				if err != nil {
					if !yield(rawTypeReplacerData{}, err) {
						return
					}
					continue
				}

				targets := hiter.Collect2(seq)
				for _, matched := range hiter.Values2(targets) {
					switch matched.Variant {
					case MatchedAsStruct:
						for _, f := range matched.Field {
							if f.As == MatchedAsImplementor {
								ty, ok := ConstUnd.ConversionMethod.ConvertedType(f.TypeInfo.(*types.Named))
								if !ok {
									continue
								}
								imports = appendTypeAndTypeParams(imports, pkg.PkgPath, ty)
							}
							if f.Elem != nil && f.Elem.As == MatchedAsImplementor {
								var elem types.Type
								switch x := f.TypeInfo.(type) {
								case *types.Named:
									elem = x
								case *types.Array:
									elem = x.Elem()
								case *types.Slice:
									elem = x.Elem()
								case *types.Map:
									elem = x.Elem()
								}
								ty, ok := ConstUnd.ConversionMethod.ConvertedType(elem.(*types.Named).TypeArgs().At(0).(*types.Named))
								if !ok {
									continue
								}
								imports = appendTypeAndTypeParams(imports, pkg.PkgPath, ty)
							}
						}
					case MatchedAsArray, MatchedAsSlice, MatchedAsMap:
						f := matched.Field[0]
						if f.As == MatchedAsImplementor {
							ty, ok := ConstUnd.ConversionMethod.ConvertedType(f.Elem.TypeInfo.(*types.Named))
							if !ok {
								continue
							}
							imports = appendTypeAndTypeParams(imports, pkg.PkgPath, ty)
						}
					}
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
						targets:   targets,
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
	matched, err := findRawTypes(pkgs, imports, validatorMethod, nil, false, nil)
	if err != nil {
		return matched, err
	}

	// TODO: use filter instead
	matched = collectRawTypes(
		filterRawTypes(
			nil,
			nil,
			func(rmt RawMatchedType) bool {
				if rmt.Variant != MatchedAsStruct {
					return true
				}
				var count int
				for _, f := range rmt.Field {
					if (f.UndTag.IsSome() && f.As != MatchedAsImplementor) || f.As == MatchedAsImplementor {
						count++
					}
				}
				return count > 0
			},
			matched.Iter(),
		),
	)

	// 2nd path, find including implementor
	matched, err = findRawTypes(pkgs, imports, validatorMethod, matched, true, nil)
	if err != nil {
		return matched, err
	}

	matched = collectRawTypes(
		filterRawTypes(
			nil,
			nil,
			func(rmt RawMatchedType) bool {
				if rmt.Variant != MatchedAsStruct {
					return true
				}
				var count int
				for _, f := range rmt.Field {
					if (f.UndTag.IsSome() && f.As != MatchedAsImplementor) || f.As == MatchedAsImplementor {
						count++
					}
				}
				return count > 0
			},
			matched.Iter(),
		),
	)

	return matched, nil
}

type ValidatorMethod struct {
	Name string
}

func (method ValidatorMethod) IsImplementor(ty *types.Named) bool {
	return isValidatorImplementor(ty, method.Name)
}

func isValidatorImplementor(ty *types.Named, methodName string) bool {
	ms := types.NewMethodSet(types.NewPointer(ty))
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
			return named.Obj().Pkg() == nil && named.Obj().Name() == "error"
		}
	}
	return false
}

func printValidator(undtagImportIdent string, tagOpt undtag.UndOpt) string {
	var builder strings.Builder
	builder.WriteString(undtagImportIdent + ".UndOptExport{\n")
	if tagOpt.States().IsSome() {
		s := tagOpt.States().Value()
		builder.WriteString(fmt.Sprintf("\tStates: &%s.StateValidator{\n", undtagImportIdent))
		if s.Def {
			builder.WriteString("\t\tDef: true,\n")
		}
		if s.Null {
			builder.WriteString("\t\tNull: true,\n")
		}
		if s.Und {
			builder.WriteString("\t\tUnd: true,\n")
		}
		builder.WriteString("\t},\n")
	}
	if tagOpt.Len().IsSome() {
		l := tagOpt.Len().Value()
		builder.WriteString(fmt.Sprintf("\tLen: &%s.LenValidator{\n", undtagImportIdent))
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
		builder.WriteString("\t},\n")
	}
	if tagOpt.Values().IsSome() {
		v := tagOpt.Values().Value()
		builder.WriteString(fmt.Sprintf("\tValues: &%s.ValuesValidator{\n", undtagImportIdent))
		if v.Nonnull {
			builder.WriteString("\t\tNonnull: true,\n")
		}
		builder.WriteString("\t},\n")
	}
	builder.WriteString("}.Into()")
	return builder.String()
}
