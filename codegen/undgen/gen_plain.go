package undgen

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"log/slog"
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

//go:generate go run ../ undgen plain --pkg ./testdata/targettypes/ --pkg ./testdata/targettypes/sub --pkg ./testdata/targettypes/sub2

func GeneratePlain(
	sourcePrinter *suffixwriter.Writer,
	verbose bool,
	pkgs []*packages.Package,
	imports []TargetImport,
) error {
	rawTypes, err := FindRawTypes(pkgs, imports, ConstUnd.ConversionMethod)
	if err != nil {
		return err
	}
	for data, err := range xiter.Map2(
		replaceToPlainTypes,
		preprocessRawTypes(imports, rawTypes),
	) {
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

		for _, s := range hiter.Values2(data.targets) {
			dts := data.dec.Dst.Nodes[s.TypeSpec].(*dst.TypeSpec)
			ts := res.Ast.Nodes[dts].(*ast.TypeSpec)
			buf.WriteString("//" + UndDirectivePrefix + UndDirectiveCommentGenerated + "\n")
			buf.WriteString(token.TYPE.String())
			buf.WriteByte(' ')
			err = printer.Fprint(buf, res.Fset, ts)
			if err != nil {
				return fmt.Errorf("print.Fprint failed for type %s in file %q: %w", data.filename, ts.Name.Name, err)
			}
			buf.WriteString("\n\n")

			_, err = generateMethodToPlain(
				buf,
				data.dec,
				dts,
				ts.Name.Name[:len(ts.Name.Name)-len("Plain")]+printTypeParamVars(dts),
				ts.Name.Name+printTypeParamVars(dts),
				s,
				data.importMap,
			)
			if err != nil {
				return err
			}

			buf.WriteString("\n\n")
		}

		if len(data.targets) > 0 {
			err = sourcePrinter.Write(context.Background(), data.filename, buf.Bytes())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// replaceToPlainTypes replaces type spec in *dst.File. Later replaced types are printed to output file.
func replaceToPlainTypes(ty rawTypeReplacerData, err error) (rawTypeReplacerData, error) {
	if err != nil {
		return ty, err
	}
	var targets hiter.KeyValues[int, RawMatchedType]
	for idx, rawTy := range hiter.Values2(ty.targets) {
		modified, err := unwrapUndFields(ty.dec.Dst.Nodes[rawTy.TypeSpec].(*dst.TypeSpec), rawTy, ty.importMap)
		if err != nil {
			return ty, err
		}
		if !modified {
			continue
		}
		targets = append(targets, hiter.KeyValue[int, RawMatchedType]{K: idx, V: rawTy})
	}
	ty.targets = targets
	return ty, err
}

func unwrapUndFields(ts *dst.TypeSpec, target RawMatchedType, importMap importDecls) (bool, error) {
	if target.Variant != MatchedAsStruct { //TODO remove this constraint
		return false, nil
	}

	// typeName := ts.Name.Name
	ts.Name.Name = ts.Name.Name + "Plain"

	var (
		err        error
		atLeastOne bool
	)
	dstutil.Apply(
		ts.Type,
		func(c *dstutil.Cursor) bool {
			if err != nil {
				return false
			}

			node := c.Node()
			switch field := node.(type) {
			default:
				return true
			case *dst.Field:
				if len(field.Names) == 0 {
					return false
				}

				mf, ok := target.FieldByName(field.Names[0].Name)
				if !ok {
					return false
				}

				if mf.UndTag.IsNone() {
					return false
				}
				undOptParseResult := mf.UndTag.Value()
				if undOptParseResult.Err != nil {
					if err == nil {
						err = undOptParseResult.Err
					}
					return false
				}

				undOpt := undOptParseResult.Opt

				// later mutated
				// We need allocate one since field.Tag is nil when no tag is set.
				tag := &dst.BasicLit{}
				if field.Tag != nil {
					tag = field.Tag
				}
				field.Tag = tag

				switch mf.As {
				// TODO add more match pattern
				case MatchedAsDirect:
					field, modified := unwrapUndFieldsDirect(field, mf, undOpt, importMap)
					if modified {
						atLeastOne = true
					}
					c.Replace(field)
				}
			}
			return false
		},
		nil,
	)
	return atLeastOne, err
}

func unwrapUndFieldsDirect(field *dst.Field, mf MatchedField, undOpt undtag.UndOpt, importMap importDecls) (*dst.Field, bool) {
	modified := true

	fieldTy := field.Type.(*dst.IndexExpr) // X.Sel[Index]
	sel := fieldTy.X.(*dst.SelectorExpr)   // X.Sel
	switch mf.Type {
	case UndTargetTypeOption:
		switch s := undOpt.States().Value(); {
		default:
			modified = false
		case s.Def && (s.Null || s.Und):
			modified = false
		case s.Def:
			field.Type = fieldTy.Index // unwrap, simply T.
		case s.Null || s.Und:
			field.Type = startStructExpr() // *struct{}
		}
	case UndTargetTypeUnd, UndTargetTypeSliceUnd:
		switch s := undOpt.States().Value(); {
		case s.Def && s.Null && s.Und:
			modified = false
		case s.Def && (s.Null || s.Und):
			*sel = *importMap.DstExpr(UndTargetTypeOption)
		case s.Null && s.Und:
			fieldTy.Index = startStructExpr()
			*sel = *importMap.DstExpr(UndTargetTypeOption)
		case s.Def:
			// unwrap
			field.Type = fieldTy.Index
		case s.Null || s.Und:
			field.Type = startStructExpr()
		}
	case UndTargetTypeElastic, UndTargetTypeSliceElastic:
		isSlice := mf.Type == UndTargetTypeSliceElastic

		// early return if nothing to change
		if (undOpt.States().IsSomeAnd(func(s undtag.StateValidator) bool {
			return s.Def && s.Null && s.Und
		})) && (undOpt.Len().IsNone() || undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool {
			// when opt is eq, we'll narrow its type to [n]T. but otherwise it remains []T
			return lv.Op != undtag.LenOpEqEq
		})) && (undOpt.Values().IsNone()) {
			return field, false
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
			fieldTy.Index = startStructExpr()
			fieldTy.X = importMap.DstExpr(UndTargetTypeOption)
		case s.Def:
			// und.Und[[]option.Option[T]] -> []option.Option[T]
			field.Type = fieldTy.Index
		case s.Null || s.Und:
			field.Type = startStructExpr()
		}
	}

	return field, modified
}

func generateMethodToPlain(
	w io.Writer,
	dec *decorator.Decorator,
	ts *dst.TypeSpec,
	tyName string,
	modifiedTyName string, // must include type param
	target RawMatchedType,
	importMap importDecls,
) (ok bool, err error) {
	if target.Variant != MatchedAsStruct { //TODO remove this constraint
		return false, nil
	}

	printf, flush := bufPrintf(w)
	defer func() {
		fErr := flush()
		if err != nil {
			return
		}
		err = fErr
	}()

	printf(
		`func (v %[1]s) UndPlain() %[2]s {
	return %[2]s{
`,
		tyName, modifiedTyName,
	)
	defer func() {
		printf(`}
		}
		
`)
	}()

	var (
		atLeastOne bool
	)
	dstutil.Apply(
		ts.Type,
		func(c *dstutil.Cursor) bool {
			if err != nil {
				return false
			}

			node := c.Node()
			switch field := node.(type) {
			default:
				return true
			case *dst.Field:
				if len(field.Names) == 0 { // Is it possible?
					return false
				}

				var fieldConverter func(ident string) string
				defer func() {
					if fieldConverter == nil {
						fieldConverter = func(ident string) string {
							return ident
						}
					}
					// TODO: move this line to somewhere when adding conversion other than "direct".
					for _, n := range field.Names {
						printf("\t%s: %s,\n", n.Name, fieldConverter("v."+n.Name))
					}
				}()

				mf, ok := target.FieldByName(field.Names[0].Name)
				if !ok || mf.UndTag.IsNone() {
					return false
				}

				undOptParseResult := mf.UndTag.Value()
				if undOptParseResult.Err != nil {
					if err == nil {
						err = undOptParseResult.Err
					}
					return false
				}

				undOpt := undOptParseResult.Opt

				switch mf.As {
				// TODO add more match pattern
				case MatchedAsDirect:
					var param string
					param, err = printTypeParamForField(dec.Fset, dec.Ast.Nodes[ts].(*ast.TypeSpec), field.Names[0].Name)
					if err != nil {
						return false
					}

					fieldConverter = generateMethodToPlainDirect(mf, undOpt, param, importMap)
					return false
				}
			}
			return false
		},
		nil,
	)
	return atLeastOne, err
}

func generateMethodToPlainDirect(mf MatchedField, undOpt undtag.UndOpt, typeParam string, importMap importDecls) func(ident string) string {
	switch mf.Type {
	case UndTargetTypeOption:
		return optionToPlain(undOpt)
	case UndTargetTypeUnd, UndTargetTypeSliceUnd:
		return undToPlain(undOpt, importMap)
	case UndTargetTypeElastic, UndTargetTypeSliceElastic:
		return elasticToPlain(mf, undOpt, typeParam, importMap)
	}
	return nil
}

func optionToPlain(undOpt undtag.UndOpt) func(ident string) string {
	switch s := undOpt.States().Value(); {
	default:
		return nil
	case s.Def && (s.Null || s.Und):
		return nil
	case s.Def:
		return func(fieldName string) string {
			return fmt.Sprintf("%s.Value()", fieldName)
		}
	case s.Null || s.Und:
		return func(fieldName string) string { return "nil" }
	}
}

func undToPlain(undOpt undtag.UndOpt, importMap importDecls) func(ident string) string {
	convertIdent, _ := importMap.Ident(UndPathConversion)
	switch s := undOpt.States().Value(); {
	default:
		return nil
	case s.Def && s.Null && s.Und:
		return nil
	case s.Def && (s.Null || s.Und):
		return func(ident string) string {
			return fmt.Sprintf("%s.Unwrap().Value()", ident)
		}
	case s.Null && s.Und:
		return func(ident string) string {
			return fmt.Sprintf("%s.UndNullish(%s)", convertIdent, ident)
		}
	case s.Def:
		return func(ident string) string {
			return fmt.Sprintf("%s.Value()", ident)
		}
	case s.Null || s.Und:
		return func(ident string) string { return "nil" }
	}
}

func elasticToPlain(mf MatchedField, undOpt undtag.UndOpt, typeParam string, importMap importDecls) func(ident string) string {
	optionIdent, _ := importMap.Ident(UndTargetTypeOption.ImportPath)
	undIdent, _ := importMap.Ident(UndTargetTypeUnd.ImportPath)
	sliceUndIdent, _ := importMap.Ident(UndTargetTypeSliceUnd.ImportPath)
	convertIdent, _ := importMap.Ident(UndPathConversion)
	isSlice := mf.Type == UndTargetTypeSliceElastic
	// very really simple case.
	if undOpt.States().IsSome() && undOpt.Len().IsNone() && undOpt.Values().IsNone() {
		switch s := undOpt.States().Value(); {
		default:
			return nil
		case s.Def && s.Null && s.Und:
			return nil
		case s.Def && (s.Null || s.Und):
			return func(ident string) string {
				return fmt.Sprintf(`%s.UnwrapElastic%s(%s).Unwrap().Value()`,
					convertIdent, orIsSlice("", "Slice", isSlice), ident,
				)
			}
		case s.Null && s.Und:
			return func(ident string) string {
				return fmt.Sprintf("%s.UndNullish(%s)", convertIdent, ident)
			}
		case s.Def:
			return func(ident string) string {
				return fmt.Sprintf("%s.Unwrap().Value()", ident)
			}
		case s.Null || s.Und:
			return func(ident string) string {
				return "nil"
			}
		}
	}

	states := undOpt.States().Value()
	if !states.Def {
		// return early.
		switch s := states; {
		case s.Null && s.Und:
			return func(ident string) string {
				return fmt.Sprintf("%s.Unwrap().Value()", ident)
			}
		case s.Null || s.Und:
			return func(ident string) string {
				return "nil"
			}
		}
	}
	// fist, converts Elastic[T] -> Und[[]option.Option[T]]
	c := func(ident string) string {
		return fmt.Sprintf(`%s.%s(%s)`,
			convertIdent, suffixSlice("UnwrapElastic", isSlice), ident,
		)
	}
	wrappers := []func(val string) string{}

	if undOpt.Len().IsSome() {
		lv := undOpt.Len().Value()

		switch lv.Op {
		default:
			panic("unknown len op")
		case undtag.LenOpEqEq:
			// to [n]option.Option[T]
			wrappers = append(wrappers, func(val string) string {
				return fmt.Sprintf(
					`%[1]s.Map(
						%[2]s,
						func(o []%[3]s.Option[%[4]s]) (out [%[5]d]option.Option[%[4]s]) {
					 		copy(out[:], o)
					 		return out
						},
					)`,
					/*1*/ orIsSlice(undIdent, sliceUndIdent, isSlice),
					/*2*/ val,
					/*3*/ optionIdent,
					/*4*/ typeParam,
					/*5*/ lv.Len,
				)
			})
		case undtag.LenOpGr, undtag.LenOpGrEq, undtag.LenOpLe, undtag.LenOpLeEq:
			// other then trim down or append it to the size at most or at least.
			var methodName string
			len := lv.Len
			switch lv.Op {
			case undtag.LenOpGr:
				methodName = "LenNAtLeast"
				len += 1
			case undtag.LenOpGrEq:
				methodName = "LenNAtLeast"
			case undtag.LenOpLe:
				methodName = "LenNAtMost"
				len -= 1
			case undtag.LenOpLeEq:
				methodName = "LenNAtMost"
			}
			wrappers = append(wrappers, func(val string) string {
				return fmt.Sprintf("%s.%s(%d, %s)", convertIdent, suffixSlice(methodName, isSlice), len, val)
			})
		}
	}
	if undOpt.Values().IsSome() {
		v := undOpt.Values().Value()
		switch {
		case v.Nonnull:
			if undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool { return lv.Op == undtag.LenOpEqEq }) {
				/*
					{{.UndPkg}}.Map(
						{{.Arg}},
						func(s [{{.Size}}]{{.OptionPkg}}.Option[{{.TypeParam}}]) (r [{{.Size}}]{{.TypeParam}}) {
							for i := 0; i < {{.Size}}; i++ {
								r[i] = s[i].Value()
							}
							return
						},
					)
				*/
				wrappers = append(wrappers, func(val string) string {
					return fmt.Sprintf(
						`%[1]s.Map(
							%[2]s,
							func(s [%[3]d]%[4]s.Option[%[5]s]) (r [%[3]d]%[5]s) {
								for i := 0; i < %[3]d; i++ {
									r[i] = s[i].Value()
								}
								return
							},
						)`,
						/*1*/ orIsSlice(undIdent, sliceUndIdent, isSlice),
						/*2*/ val,
						/*3*/ undOpt.Len().Value().Len,
						/*4*/ optionIdent,
						/*5*/ typeParam,
					)
				})
			} else {
				wrappers = append(wrappers, func(val string) string {
					return fmt.Sprintf(
						`%s.NonNull%s(%s)`,
						convertIdent, orIsSlice("", "Slice", isSlice), val,
					)
				})
			}
			// no other cases at the moment. I don't expect it to expand tho.
		}
	}

	// Then when len == 1, convert und.Und[[1]option.Option[T]] or und.Und[[1]T] to und.Und[option.Option[T]], und.Und[T] respectively
	if undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool { return lv.Op == undtag.LenOpEqEq && lv.Len == 1 }) {
		wrappers = append(wrappers, func(val string) string {
			return fmt.Sprintf("%s.UnwrapLen1%s(%s)", convertIdent, orIsSlice("", "Slice", isSlice), val)
		})
	}

	// Finally unwrap value based on req,null,und
	if wrapper := undToPlain(undOpt, importMap); wrapper != nil {
		wrappers = append(wrappers, wrapper)
	}

	return func(ident string) string {
		exp := c(ident)
		for _, wrapper := range wrappers {
			exp = wrapper(exp)
		}
		return exp
	}
}

func orIsSlice[T any](l, r T, isSlice bool) T {
	if !isSlice {
		return l
	}
	return r
}

func suffixSlice(s string, isSlice bool) string {
	if isSlice {
		s += "Slice"
	}
	return s
}
