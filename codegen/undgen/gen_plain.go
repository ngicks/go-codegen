package undgen

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
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

//go:generate go run ../ undgen plain --pkg ./internal/targettypes/ --pkg ./internal/targettypes/sub --pkg ./internal/targettypes/sub2
//go:generate go run ../ undgen plain --pkg ./internal/patchtarget/...
//go:generate go run ../ undgen plain --pkg ./internal/validatortarget/...

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

			err = generateMethodToPlain(
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

			err = generateMethodToRaw(
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

func sliceSuffix(isSlice bool) string {
	if isSlice {
		return "Slice"
	}
	return ""
}

func suffixSlice(s string, isSlice bool) string {
	if isSlice {
		s += "Slice"
	}
	return s
}
