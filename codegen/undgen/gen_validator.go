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
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
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
	// append imports other than und imports.
	// The generated code uses fmt.Errorf.
	imports = AppendTargetImports(imports, TargetImport{ImportPath: "fmt"})

	replacerData, err := gatherValidatableUndTypes(
		pkgs,
		imports,
		isUndValidatorAllowedEdge,
		func(g *typeGraph) iter.Seq2[typeIdent, *typeNode] {
			return g.iterUpward(true, isUndValidatorAllowedEdge)
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
		for _, node := range data.targetNodes {
			dts := data.dec.Dst.Nodes[node.ts].(*dst.TypeSpec)
			written, err := generateUndValidate(
				buf,
				dts,
				node,
				data.importMap,
			)
			if written {
				atLeastOne = true
			}
			if err != nil {
				return fmt.Errorf("generating UndValidate for type %s in file %q: %w", node.ts.Name.Name, data.filename, err)
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
	node *typeNode,
	imports importDecls,
) (written bool, err error) {
	typeName := ts.Name.Name + printTypeParamVars(ts)
	undtagImportIdent, _ := imports.Ident(UndPathUndTag)
	validateImportIdent, _ := imports.Ident(UndPathValidate)

	buf := new(bytes.Buffer)

	// true only when validator is meaningful.
	var shouldPrint bool
	defer func() {
		if err != nil || !shouldPrint {
			return
		}
		written = true
		_, err = w.Write(buf.Bytes())
	}()

	printf, flush := bufPrintf(buf)
	defer func() {
		fErr := flush()
		if err != nil {
			return
		}
		err = fErr
	}()

	printf("//%s%s\n", UndDirectivePrefix, UndDirectiveCommentGenerated)
	printf("func (v %s) UndValidate() (err error) {\n", typeName)
	defer printf(`return
	}
`)

	loopName := func(s string) string {
		if s == "" {
			return "LOOP"
		}
		return "LOOP_" + s
	}

	// unwrappers to reach final destination type(implementor or und types.)
	validatorUnwrappers := func(fieldName string, pointer []typeDependencyEdgePointer) []func(exp string) string {
		var wrappers []func(exp string) string
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
							break %s
						}
					}
`,
					exp, validateImportIdent, loopName(fieldName),
				)
			})
		}
		if len(wrappers) > 0 {
			// label with loopName, later break-ed when non nil error is returned from validator implementation.
			wrappers = slices.Insert(
				wrappers,
				0,
				func(exp string) string {
					return fmt.Sprintf("%s:\n%s", loopName(fieldName), exp)
				},
			)
		}
		return wrappers
	}

	switch x := node.typeInfo.Underlying().(type) {
	case *types.Map, *types.Array, *types.Slice:
		// should be only one since we prohibit struct literals.
		ident, edge := firstTypeIdent(node.children)
		// An implementor or implementor wrapped in und types
		exp := `err = v.UndValidate()`
		_ = matchUndTypeBool(
			ident.targetType(),
			false,
			func() {
				exp = fmt.Sprintf("err = %s.UndValidate(v)", importIdent(ident.targetType(), imports))
			},
			nil,
			nil,
		)
		for _, w := range slices.Backward(validatorUnwrappers("", edge.stack)) {
			exp = w(exp)
		}
		shouldPrint = true
		// later processed through fmt.*printf functions.
		printf(strings.ReplaceAll(exp, "%", "%%"))
	case *types.Struct:
		edges := maps.Collect(node.fields())
		for i, f := range pkgsutil.EnumerateFields(x) {
			edge, ok := edges[i]
			if !ok {
				// nothing to validate
				continue
			}
			if !isUndValidatorAllowedEdge(edge) {
				continue
			}

			undTagValue, hasTag := reflect.StructTag(x.Tag(i)).Lookup(undtag.TagName)

			// There's cases where matched by but needed to be rejected.
			// 1. not tagged, being und filed, type arg is not a implementor
			// TODO: remove this line and check test result.
			// For now this guard is supposed to have been done by transitive edge filtering.
			if !hasTag && isUndType(edge.childType) &&
				!edge.hasSingleNamedTypeArg(isUndValidatorImplementor) {
				continue
			}

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
					namedTypeToTargetType(edge.childType),
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
					if edge.hasSingleNamedTypeArg(isUndValidatorImplementor) {
						wrappeeValidator = func(ident string) string {
							return fmt.Sprintf(
								`
								if err == nil {
									err = %s.UndValidate(%s)
								}
`,
								importIdent(
									namedTypeToTargetType(edge.childType),
									imports,
								),
								ident,
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
						return fmt.Sprintf(`err = %s.UndValidate()`, ident)
					}
				}

				var exp string
				wrappers := validatorUnwrappers(f.Name(), edge.stack[1:]) // skip first one; is always prefixed with struct kind.
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
