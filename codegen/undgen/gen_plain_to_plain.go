package undgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
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
	rawFields map[string]string,
	plainFields map[string]string,
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
				if !ok {
					return false
				}
				if mf.UndTag.IsNone() && mf.As != MatchedAsImplementor && (mf.Elem != nil && mf.Elem.As != MatchedAsImplementor) {
					return false
				}

				var undOpt undtag.UndOpt
				if mf.UndTag.IsSome() {
					undOptParseResult := mf.UndTag.Value()
					if undOptParseResult.Err != nil {
						if err == nil {
							err = undOptParseResult.Err
						}
						return false
					}
					undOpt = undOptParseResult.Opt
				}

				var param string
				switch {
				case mf.Elem != nil:
					switch mf.Elem.As {
					case MatchedAsImplementor:
						var elem types.Type
						switch x := mf.TypeInfo.(type) {
						case *types.Named:
							elem = x
						case *types.Array:
							elem = x.Elem()
						case *types.Slice:
							elem = x.Elem()
						}
						expr := conversionTargetOfImplementorAst(
							target,
							elem.(*types.Named).TypeArgs().At(0).(*types.Named),
							importMap,
						)
						buf := new(bytes.Buffer)
						err = printer.Fprint(buf, token.NewFileSet(), expr)
						if err != nil {
							return false
						}
						param = buf.String()
					case MatchedAsDirect:
						if mf.Elem.Elem != nil && mf.Elem.Elem.As == MatchedAsImplementor {
							expr := conversionTargetOfImplementorAst(
								target,
								mf.Elem.Elem.TypeInfo.(*types.Named),
								importMap,
							)
							buf := new(bytes.Buffer)
							err = printer.Fprint(buf, token.NewFileSet(), expr)
							if err != nil {
								return false
							}
							param = buf.String()
						} else {
							ts := dec.Ast.Nodes[ts].(*ast.TypeSpec)
							param, err = printTypeParamForField(dec.Fset, ts, field.Names[0].Name)
							if err != nil {
								return false
							}
						}
					}
				default:
					param, err = printTypeParamForField(dec.Fset, dec.Ast.Nodes[ts].(*ast.TypeSpec), field.Names[0].Name)
					if err != nil {
						return false
					}
				}
				switch mf.As {
				// TODO add more match pattern
				case MatchedAsDirect:
					fieldConverter, _ = generateMethodToPlainDirect(mf, undOpt, param, importMap)
					return false
				case MatchedAsArray:
					mapper, needsArg := generateMethodToPlainDirect(*mf.Elem, undOpt, param, importMap)
					if mapper == nil {
						mapper = func(ident string) string {
							return ident + ".UndPlain()"
						}
						needsArg = true
					}
					fieldConverter = func(ident string) string {
						return fmt.Sprintf(
							`func(in %[1]s) (out %[2]s) {
								for k %[3]s := range in {
									out[k] = %[4]s
								}
								return out
							}(%[5]s)`,
							/*1*/ rawFields[mf.Name],
							/*2*/ plainFields[mf.Name],
							/*3*/ func() string {
								if needsArg {
									return ", v"
								} else {
									return ""
								}
							}(),
							/*4*/ mapper("v"),
							/*5*/ ident,
						)
					}
				case MatchedAsSlice, MatchedAsMap:
					mapper, needsArg := generateMethodToPlainDirect(*mf.Elem, undOpt, param, importMap)
					fieldConverter = func(ident string) string {
						return fmt.Sprintf(
							`func(in %[1]s) %[2]s {
								out := make(%[2]s, len(in))
								for k %[3]s  := range in {
									out[k] = %[4]s
								}
								return out
							}(%[5]s)`,
							/*1*/ rawFields[mf.Name],
							/*2*/ plainFields[mf.Name],
							/*3*/ func() string {
								if needsArg {
									return ", v"
								} else {
									return ""
								}
							}(),
							/*4*/ mapper("v"),
							/*5*/ ident,
						)
					}
				case MatchedAsImplementor:
					fieldConverter = func(ident string) string {
						return ident + ".UndPlain()"
					}
				}
			}
			return false
		},
		nil,
	)
	return err
}

func generateMethodToPlainDirect(mf MatchedField, undOpt undtag.UndOpt, typeParam string, importMap importDecls) (convert func(ident string) string, needsArg bool) {
	switch mf.Type {
	case UndTargetTypeOption:
		convert, needsArg = optionToPlain(undOpt)
	case UndTargetTypeUnd, UndTargetTypeSliceUnd:
		convert, needsArg = undToPlain(undOpt, importMap)
	case UndTargetTypeElastic, UndTargetTypeSliceElastic:
		convert, needsArg = elasticToPlain(mf, undOpt, typeParam, importMap)
	}
	if mf.Elem != nil && mf.Elem.As == MatchedAsImplementor {
		conversionIdent, _ := importMap.Ident(UndPathConversion)
		pkgIdent := importIdent(mf.Type, importMap)
		inner := convert
		convert = func(ident string) string {
			return inner(fmt.Sprintf(
				`%s.Map(
				%s,
				%s.ToPlain,
			)`,
				pkgIdent, ident, conversionIdent,
			))
		}
		needsArg = true
	}
	return
}

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

func undToPlain(undOpt undtag.UndOpt, importMap importDecls) (func(ident string) string, bool) {
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

func elasticToPlain(mf MatchedField, undOpt undtag.UndOpt, typeParam string, importMap importDecls) (func(ident string) string, bool) {
	optionIdent, _ := importMap.Ident(UndTargetTypeOption.ImportPath)
	convertIdent, _ := importMap.Ident(UndPathConversion)
	isSlice := targetTypeIsSlice(mf.Type)
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
