package undgen

import (
	"context"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"iter"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/pkg/astutil"
	"github.com/ngicks/go-codegen/codegen/pkg/imports"
	"github.com/ngicks/go-codegen/codegen/internal/bufpool"
	"github.com/ngicks/go-codegen/codegen/pkg/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/pkg/suffixwriter"
	"github.com/ngicks/go-codegen/codegen/pkg/typegraph"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/und/undtag"
	"golang.org/x/tools/go/packages"
)

func GenerateValidator(
	sourcePrinter *suffixwriter.Writer,
	verbose bool,
	pkgs []*packages.Package,
	extra []imports.TargetImport,
) error {
	parser := imports.NewParserPackages(pkgs)
	parser.AppendExtra(extra...)
	// The generated code uses fmt.Errorf.
	parser.AppendExtra(imports.TargetImport{Import: imports.Import{Path: "fmt", Name: "fmt"}})

	replacerData, err := gatherValidatableUndTypes(
		pkgs,
		parser,
		isUndValidatorAllowedEdge,
		func(g *typegraph.Graph) iter.Seq2[typegraph.Ident, *typegraph.Node] {
			return g.IterUpward(true, isUndValidatorAllowedEdge)
		},
	)
	if err != nil {
		return err
	}

	buf := bufpool.GetBuf()
	defer bufpool.PutBuf(buf)

	for _, data := range hiter.Filter2(
		func(f *ast.File, data *typegraph.ReplaceData) bool { return f != nil && data != nil },
		hiter.MapsKeys(replacerData, pkgsutil.EnumerateFile(pkgs)),
	) {
		buf.Reset()

		if verbose {
			slog.Debug(
				"found",
				slog.String("filename", data.Filename),
			)
		}

		data.ImportMap.AddMissingImports(data.DstFile)
		res := decorator.NewRestorer()
		af, err := res.RestoreFile(data.DstFile)
		if err != nil {
			return fmt.Errorf("converting dst to ast for %q: %w", data.Filename, err)
		}

		if err := astutil.PrintFileHeader(buf, af, res.Fset); err != nil {
			return fmt.Errorf("%q: %w", data.Filename, err)
		}

		var atLeastOne bool
		for _, node := range data.TargetNodes {
			dts := data.Dec.Dst.Nodes[node.Ts].(*dst.TypeSpec)
			written, err := generateUndValidate(
				buf,
				dts,
				node,
				data.ImportMap,
			)
			if written {
				atLeastOne = true
			}
			if err != nil {
				return fmt.Errorf("generating UndValidate for type %s in file %q: %w", node.Ts.Name.Name, data.Filename, err)
			}
			buf.WriteString("\n\n")
		}

		if atLeastOne {
			err = sourcePrinter.Write(context.Background(), data.Filename, buf.Bytes())
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
	node *typegraph.Node,
	imports imports.ImportMap,
) (written bool, err error) {
	typeName := ts.Name.Name + astutil.PrintTypeParamsDst(ts)
	undtagImportIdent, _ := imports.Ident(UndPathUndTag)
	validateImportIdent, _ := imports.Ident(UndPathValidate)

	buf := bufpool.GetBuf()
	defer bufpool.PutBuf(buf)

	// true only when validator is meaningful.
	var shouldPrint bool
	defer func() {
		if err != nil || !shouldPrint {
			return
		}
		written = true
		_, err = w.Write(buf.Bytes())
	}()

	printf, flush := astutil.BufPrintf(buf)
	defer func() {
		fErr := flush()
		if err != nil {
			return
		}
		err = fErr
	}()

	printf("//%s%s\n", astutil.DirectivePrefix, astutil.DirectiveCommentGenerated)
	printf("func (v %s) UndValidate() (err error) {\n", typeName)
	defer printf(`return
	}
`)

	// unwrappers to reach final destination type(implementor or und types.)
	validatorUnwrappers := func(pointer []typegraph.EdgeRouteNode) []func(exp string) string {
		var wrappers []func(exp string) string
		if len(pointer) > 0 && pointer[len(pointer)-1].Kind == typegraph.EdgeKindPointer {
			pointer = pointer[:len(pointer)-1]
		}
		for range pointer {
			wrappers = append(wrappers, func(exp string) string {
				return fmt.Sprintf(
					`for k, v := range v {
						%s
						if err != nil {
							err = %s.AppendValidationErrorIndex(
								err,
								fmt.Sprintf("%%v", k),
							)
							break
						}
					}
`,
					exp, validateImportIdent,
				)
			})
		}
		return wrappers
	}

	edgeMap := node.ChildEdgeMap(isUndValidatorAllowedEdge)
	switch x := node.Type.Underlying().(type) {
	case *types.Map, *types.Array, *types.Slice:
		// should be only one since we prohibit struct literals.
		ident, edge, _ := edgeMap.First()
		isPointer := edge.LastPointer().IsSomeAnd(func(tdep typegraph.EdgeRouteNode) bool {
			return tdep.Kind == typegraph.EdgeKindPointer
		})
		// An implementor or implementor wrapped in und types
		exp := fmt.Sprintf(
			`%s err = v.UndValidate() %s`,
			or(isPointer, fmt.Sprintf("if %s {\n", _printUndValidateCallableChecker(namedTypeToTargetType(edge.ChildType), "v")), ""),
			or(isPointer, "\n}", ""),
		)
		_ = matchUndTypeBool(
			ident.TargetType(),
			false,
			func() {
				_, isPointer := edge.HasSingleNamedTypeArg(nil)
				exp = fmt.Sprintf(
					`%s err = %s.UndValidate(v,%s) %s`,
					or(
						isPointer,
						fmt.Sprintf("if %s {\n", _printUndValidateCallableChecker(namedTypeToTargetType(edge.ChildType), "v")),
						"",
					),
					importIdent(ident.TargetType(), imports),
					or(isPointer, _printUndValidateElasticSkipIndices(namedTypeToTargetType(edge.ChildType), "v"), ""),
					or(isPointer, "\n}", ""),
				)
			},
			nil,
			nil,
		)
		for _, w := range slices.Backward(validatorUnwrappers(edge.Stack)) {
			exp = w(exp)
		}
		shouldPrint = true
		// later processed through fmt.*printf functions.
		printf(strings.ReplaceAll(exp, "%", "%%"))
	case *types.Struct:
		for i, f := range pkgsutil.EnumerateFields(x) {
			edge, _, _, ok := edgeMap.ByFieldPos(i)
			if !ok {
				// nothing to validate
				continue
			}

			undTagValue, hasTag := reflect.StructTag(x.Tag(i)).Lookup(undtag.TagName)

			shouldPrint = true

			func() {
				// isolate each field with block scope.
				printf("{\n")
				defer printf("}\n")

				if hasTag {
					undOpt, err := undtag.ParseOption(undTagValue)
					if err != nil { // This case should be filtered when forming the graph.
						panic(err)
					}
					printf("validator := %s\n\n", printValidator(undtagImportIdent, undOpt))
				}

				var nodeValidator func(ident string) string
				if hasTag && matchUndTypeBool(
					namedTypeToTargetType(edge.ChildType),
					false,
					func() {
						nodeValidator = func(ident string) string {
							return fmt.Sprintf(`validator.ValidOpt(%s)`, ident)
						}
					},
					func(isSlice bool) {
						nodeValidator = func(ident string) string {
							return fmt.Sprintf(`validator.ValidUnd(%s)`, ident)
						}
					},
					func(isSlice bool) {
						nodeValidator = func(ident string) string {
							return fmt.Sprintf(`validator.ValidElastic(%s)`, ident)
						}
					},
				) {
					validatorInvocation := nodeValidator

					undTypeValidator := func(ident string) string {
						return fmt.Sprintf(
							`if !%s {
							err = fmt.Errorf("%%s: value is %%s", validator.Describe(), %s.ReportState(%s))
						}`,
							validatorInvocation(ident), validateImportIdent, ident,
						)
					}

					var wrappeeValidator func(ident string) string
					isImpl, isPointer := edge.HasSingleNamedTypeArg(isUndValidatorImplementor)
					if isImpl || edge.IsTypeArgMatched() {
						wrappeeValidator = func(ident string) string {
							return fmt.Sprintf(
								`
								if err == nil %s {
									err = %s.UndValidate(%s%s)
								}
`,
								or(isPointer, "&&"+_printUndValidateCallableChecker(namedTypeToTargetType(edge.ChildType), ident), ""),
								importIdent(
									namedTypeToTargetType(edge.ChildType),
									imports,
								),
								ident,
								or(isPointer, ","+_printUndValidateElasticSkipIndices(namedTypeToTargetType(edge.ChildType), ident), ""),
							)
						}
					}

					nodeValidator = func(ident string) string {
						exp := undTypeValidator(ident)
						if wrappeeValidator != nil {
							exp += wrappeeValidator(ident)
						}
						return exp
					}
				} else {
					nodeValidator = func(ident string) string {
						isPointer := edge.LastPointer().IsSomeAnd(func(tdep typegraph.EdgeRouteNode) bool {
							return tdep.Kind == typegraph.EdgeKindPointer
						})
						// An implementor or implementor wrapped in und types
						return fmt.Sprintf(
							`%s err = %s.UndValidate() %s`,
							or(isPointer, fmt.Sprintf("if %s != nil {\n", ident), ""),
							ident,
							or(isPointer, "\n}", ""),
						)
					}
				}

				var exp string
				wrappers := validatorUnwrappers(edge.Stack[1:]) // skip first one; is always prefixed with struct kind.
				if len(wrappers) == 0 {
					exp = nodeValidator(fmt.Sprintf("v.%s", f.Name()))
				} else {
					printf("v := v.%s\n\n", f.Name())
					exp = nodeValidator("v")
					for _, w := range slices.Backward(wrappers) {
						exp = w(exp)
					}
				}
				printf(strings.ReplaceAll(exp, "%", "%%")) // later processed through fmt.*printf kind functions.
				printf(
					`
if err != nil {
						return %s.AppendValidationErrorDot(
							err,
							%q,
						)
					}
`,
					validateImportIdent, fieldJsonName(x, i),
				)
			}()
		}
	}
	return
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

func _printUndValidateCallableChecker(ty imports.TargetType, ident string) string {
	return matchUndType(
		ty,
		true,
		func() string {
			return ident + ".Value() != nil"
		},
		nil,
		func(isSlice bool) string {
			return "true"
		},
	)
}

func _printUndValidateElasticSkipIndices(ty imports.TargetType, ident string) string {
	return matchUndType(
		ty,
		true,
		func() string {
			return ""
		},
		nil,
		func(isSlice bool) string {
			return fmt.Sprintf(`func() []int {
				var skip []int
				for i, v := range %s.Values() {
					if v == nil {
						skip = append(skip, i)
					}
				}
				return skip
			}()...`,
				ident,
			)
		},
	)
}
