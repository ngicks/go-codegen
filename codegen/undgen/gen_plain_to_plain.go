package undgen

import (
	"fmt"
	"go/ast"
	"io"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/und/undtag"
)

func generateMethodToPlain(
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
	return err
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
	convertIdent, _ := importMap.Ident(UndPathConversion)
	isSlice := mf.Type == UndTargetTypeSliceElastic
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
				return fmt.Sprintf(`%s.UnwrapElastic%s(%s).Unwrap().Value()`,
					convertIdent, sliceSuffix(isSlice), ident,
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
		return fmt.Sprintf(`%s.UnwrapElastic%s(%s)`,
			convertIdent, sliceSuffix(isSlice), ident,
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
					/*1*/ undIdent,
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
						/*1*/ undIdent,
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
						convertIdent, sliceSuffix(isSlice), val,
					)
				})
			}
			// no other cases at the moment. I don't expect it to expand tho.
		}
	}

	// Then when len == 1, convert und.Und[[1]option.Option[T]] or und.Und[[1]T] to und.Und[option.Option[T]], und.Und[T] respectively
	if undOpt.Len().IsSomeAnd(func(lv undtag.LenValidator) bool { return lv.Op == undtag.LenOpEqEq && lv.Len == 1 }) {
		wrappers = append(wrappers, func(val string) string {
			return fmt.Sprintf("%s.UnwrapLen1%s(%s)", convertIdent, sliceSuffix(isSlice), val)
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
