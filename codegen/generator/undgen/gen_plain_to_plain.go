package undgen

import (
	"fmt"

	"github.com/ngicks/go-codegen/codegen/pkg/imports"
	"github.com/ngicks/und/undtag"
)

func optionToPlain(undOpt undtag.UndOpt) (func(ident string) string, bool) {
	switch s := undOpt.States().Value(); {
	default:
		return nil, false
	case s.Def && (s.Null || s.Und):
		return nil, false
	case s.Def:
		return func(fieldName string) string {
			return fmt.Sprintf("%s.Value()", fieldName)
		}, true
	case s.Null || s.Und:
		return func(fieldName string) string { return "nil" }, false
	}
}

func undToPlain(undOpt undtag.UndOpt, importMap imports.ImportMap) (func(ident string) string, bool) {
	convertIdent, _ := importMap.Ident(UndPathConversion)
	switch s := undOpt.States().Value(); {
	default:
		return nil, false
	case s.Def && s.Null && s.Und:
		return nil, false
	case s.Def && (s.Null || s.Und):
		return func(ident string) string {
			return fmt.Sprintf("%s.Unwrap().Value()", ident)
		}, true
	case s.Null && s.Und:
		return func(ident string) string {
			return fmt.Sprintf("%s.UndNullish(%s)", convertIdent, ident)
		}, true
	case s.Def:
		return func(ident string) string {
			return fmt.Sprintf("%s.Value()", ident)
		}, true
	case s.Null || s.Und:
		return func(ident string) string { return "nil" }, false
	}
}

func elasticToPlain(isSlice bool, undOpt undtag.UndOpt, typeParam string, importMap imports.ImportMap) (func(ident string) string, bool) {
	optionIdent, _ := importMap.Ident(UndTargetTypeOption.ImportPath)
	convertIdent, _ := importMap.Ident(UndPathConversion)
	undIdent, _ := importMap.Ident(UndTargetTypeUnd.ImportPath)
	if isSlice {
		undIdent, _ = importMap.Ident(UndTargetTypeSliceUnd.ImportPath)
	}
	// very really simple case.
	if undOpt.States().IsSome() && undOpt.Len().IsNone() && undOpt.Values().IsNone() {
		switch s := undOpt.States().Value(); {
		default:
			return nil, false
		case s.Def && s.Null && s.Und:
			return nil, false
		case s.Def && (s.Null || s.Und):
			return func(ident string) string {
				return fmt.Sprintf(`%s.UnwrapElastic%s(%s).Unwrap().Value()`,
					convertIdent, sliceSuffix(isSlice), ident,
				)
			}, true
		case s.Null && s.Und:
			return func(ident string) string {
				return fmt.Sprintf("%s.UndNullish(%s)", convertIdent, ident)
			}, true
		case s.Def:
			return func(ident string) string {
				return fmt.Sprintf("%s.Unwrap().Value()", ident)
			}, true
		case s.Null || s.Und:
			return func(ident string) string {
				return "nil"
			}, false
		}
	}

	states := undOpt.States().Value()
	if !states.Def {
		// return early.
		switch s := states; {
		case s.Null && s.Und:
			return func(ident string) string {
				return fmt.Sprintf("%s.Unwrap().Value()", ident)
			}, true
		case s.Null || s.Und:
			return func(ident string) string {
				return "nil"
			}, false
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
	if wrapper, _ := undToPlain(undOpt, importMap); wrapper != nil {
		wrappers = append(wrappers, wrapper)
	}

	return func(ident string) string {
		exp := c(ident)
		for _, wrapper := range wrappers {
			exp = wrapper(exp)
		}
		return exp
	}, true
}
