package undgen

import (
	"fmt"
	"go/ast"
	"io"
	"slices"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/und/undtag"
)

func generateMethodToRaw(
	w io.Writer,
	dec *decorator.Decorator,
	ts *dst.TypeSpec,
	tyName string,
	modifiedTyName string, // must include type param
	target RawMatchedType,
	importMap importDecls,
) (err error) {
	if target.Variant != MatchedAsStruct { //TODO remove this constraint
		return nil
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
		`func (v %[1]s) UndRaw() %[2]s {
	return %[2]s{
`,
		modifiedTyName, tyName,
	)
	defer func() {
		printf(`}
		}
`)
	}()

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

					fieldConverter = generateMethodToRawDirect(mf, undOpt, param, importMap)
					return false
				}
			}
			return false
		},
		nil,
	)
	return err

}

func generateMethodToRawDirect(mf MatchedField, undOpt undtag.UndOpt, typeParam string, importMap importDecls) func(ident string) string {
	switch mf.Type {
	case UndTargetTypeOption:
		return optionToRaw(undOpt, typeParam, importMap)
	case UndTargetTypeUnd, UndTargetTypeSliceUnd:
		return undToRaw(mf, undOpt, typeParam, importMap)
	case UndTargetTypeElastic, UndTargetTypeSliceElastic:
		return elasticToRaw(mf, undOpt, typeParam, importMap)
	}
	return nil
}

func optionToRaw(undOpt undtag.UndOpt, typeParam string, importMap importDecls) func(ident string) string {
	optionIdent, _ := importMap.Ident(UndTargetTypeOption.ImportPath)
	switch s := undOpt.States().Value(); {
	default:
		return nil
	case s.Def && (s.Null || s.Und):
		return nil
	case s.Def:
		return func(ident string) string {
			return fmt.Sprintf("%s.Some(%s)", optionIdent, ident)
		}
	case s.Null || s.Und:
		return func(ident string) string {
			return fmt.Sprintf("%s.None[%s]()", optionIdent, typeParam)
		}
	}
}

func undToRaw(mf MatchedField, undOpt undtag.UndOpt, typeParam string, importMap importDecls) func(ident string) string {
	convertIdent, _ := importMap.Ident(UndPathConversion)
	undIdent, _ := importMap.Ident(UndTargetTypeUnd.ImportPath)
	isSlice := targetTypeIsSlice(mf.Type)
	if isSlice {
		undIdent, _ = importMap.Ident(UndTargetTypeSliceUnd.ImportPath)
	}
	switch s := undOpt.States().Value(); {
	default:
		return nil
	case s.Def && s.Null && s.Und:
		return nil
	case s.Def && (s.Null || s.Und):
		return func(ident string) string {
			return fmt.Sprintf(
				"%s.OptionUnd%s(%t, %s)",
				convertIdent,
				sliceSuffix(isSlice),
				s.Null,
				ident,
			)
		}
	case s.Null && s.Und:
		return func(ident string) string {
			return fmt.Sprintf("%s.NullishUnd%s[%s](%s)", convertIdent, sliceSuffix(isSlice), typeParam, ident)
		}
	case s.Def:
		return func(ident string) string {
			return fmt.Sprintf("%s.Defined(%s)", undIdent, ident)
		}
	case s.Null || s.Und:
		if s.Null {
			return func(ident string) string {
				return fmt.Sprintf("%s.Null[%s]()", undIdent, typeParam)
			}
		} else {
			return func(ident string) string {
				return fmt.Sprintf("%s.Undefined[%s]()", undIdent, typeParam)
			}
		}
	}
}

func elasticToRaw(mf MatchedField, undOpt undtag.UndOpt, typeParam string, importMap importDecls) func(ident string) string {
	optionIdent, _ := importMap.Ident(UndTargetTypeOption.ImportPath)
	convertIdent, _ := importMap.Ident(UndPathConversion)

	isSlice := targetTypeIsSlice(mf.Type)

	elasticIdent, _ := importMap.Ident(UndTargetTypeElastic.ImportPath)
	if isSlice {
		elasticIdent, _ = importMap.Ident(UndTargetTypeSliceElastic.ImportPath)
	}

	undIdent, _ := importMap.Ident(UndTargetTypeUnd.ImportPath)
	if isSlice {
		undIdent, _ = importMap.Ident(UndTargetTypeSliceUnd.ImportPath)
	}

	// very really simple case.
	if undOpt.States().IsSome() && undOpt.Len().IsNone() && undOpt.Values().IsNone() {
		switch s := undOpt.States().Value(); {
		default:
			return nil
		case s.Def && s.Null && s.Und:
			return nil
		case s.Def && (s.Null || s.Und):
			return func(ident string) string {
				return fmt.Sprintf("%s.OptionOptionElastic(%t, %s)", convertIdent, s.Null, ident)
			}
		case s.Null && s.Und:
			return func(ident string) string {
				return fmt.Sprintf("%s.NullishElastic[%s](%s)", convertIdent, typeParam, ident)
			}
		case s.Def:
			return func(ident string) string {
				return fmt.Sprintf("%s.FromOptions(%s...)", elasticIdent, ident)
			}
		case s.Null || s.Und:
			if s.Null {
				return func(ident string) string {
					return fmt.Sprintf("%s.Null[%s]()", elasticIdent, typeParam)
				}
			} else {
				return func(ident string) string {
					return fmt.Sprintf("%s.Undefined[%s]()", elasticIdent, typeParam)
				}
			}
		}
	}

	states := undOpt.States().Value()
	if !states.Def {
		// return early.
		switch s := states; {
		case s.Null && s.Und:
			return func(ident string) string {
				return fmt.Sprintf("%s.NullishUnd%s[%s](%s)", convertIdent, sliceSuffix(isSlice), typeParam, ident)
			}
		case s.Null || s.Und:
			if s.Null {
				return func(ident string) string {
					return fmt.Sprintf("%s.Null[%s]()", elasticIdent, typeParam)
				}
			} else {
				return func(ident string) string {
					return fmt.Sprintf("%s.Undefined[%s]()", elasticIdent, typeParam)
				}
			}
		}
	}

	c := func(ident string) string {
		return fmt.Sprintf("%s.FromUnd(%s)", elasticIdent, ident)
	}
	wrappers := []func(val string) string{}

	// Below converts much like UndPlain but reversed order.
	// At last, Und[[]option.Option[T]] -> converts Elastic[T]
	if undOpt.Len().IsSome() {
		// if len is EqEq, map Und[[n]option.Option[T]] -> Und[[]option.Option[T]]
		lv := undOpt.Len().Value()
		switch lv.Op {
		case undtag.LenOpEqEq:
			wrappers = append(wrappers, func(val string) string {
				return fmt.Sprintf(
					`%[1]s.Map(
					%[2]s,
					func(s [%[3]d]%[4]s.Option[%[5]s]) []%[4]s.Option[%[5]s] {
						return s[:]
					},
				)`,
					undIdent, val, lv.Len, optionIdent, typeParam)
			})
		}
	}
	if undOpt.Values().IsSome() {
		v := undOpt.Values().Value()
		// Und[[n]T] -> Und[[n]option.Option[T]]
		switch {
		case v.Nonnull:
			if undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool { return lv.Op == undtag.LenOpEqEq }) {
				wrappers = append(wrappers, func(val string) string {
					return fmt.Sprintf(
						`%[1]s.Map(
							%[2]s,
							func(s [%[3]d]%[4]s) (out [%[3]d]%[5]s.Option[%[4]s]) {
								for i := 0; i < %[3]d; i++ {
									out[i] = %[5]s.Some(s[i])
								}
								return
							},
)`,
						/*1*/ undIdent,
						/*2*/ val,
						/*3*/ undOpt.Len().Value().Len,
						/*4*/ typeParam,
						/*5*/ optionIdent,
					)
				})
			} else {
				wrappers = append(wrappers, func(val string) string {
					return fmt.Sprintf(
						"%s.Nullify%s(%s)",
						convertIdent, sliceSuffix(isSlice), val,
					)
				})
			}
			// no other cases at the moment. I don't expect it to expand tho.
		}
	}

	// When len == 1, convert und.Und[[1]option.Option[T]] or und.Und[[1]T] to und.Und[option.Option[T]], und.Und[T] respectively
	if undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool { return lv.Op == undtag.LenOpEqEq && lv.Len == 1 }) {
		wrappers = append(wrappers, func(val string) string {
			return fmt.Sprintf(
				"%s.WrapLen1%s(%s)",
				convertIdent, sliceSuffix(isSlice), val,
			)
		})
	}

	// Finally wrap value based on req,null,und
	if wrapper := undToRaw(mf, undOpt, typeParam, importMap); wrapper != nil {
		wrappers = append(wrappers, wrapper)
	}

	return func(ident string) string {
		exp := ident
		for _, wrapper := range slices.Backward(wrappers) {
			exp = wrapper(exp)
		}
		return c(exp)
	}
}
