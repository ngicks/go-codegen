package undgen

import (
	"go/ast"
	"go/token"
	"iter"
	"slices"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
)

type patchTypesGenerationData struct {
	df        *dst.File
	filename  string
	typeNames []string
}

func generatePatcher(pkg *packages.Package, imports []TargetImport, targetTypeNames ...string) iter.Seq2[patchTypesGenerationData, error] {
	return func(yield func(patchTypesGenerationData, error) bool) {
		for pkg, seq := range FindTypes(pkg, targetTypeNames...) {
			for file, seq := range seq {
				importMap := parseImports(file.Imports, imports)

				var firstErr error
				replaceData := slices.Collect(
					xiter.Map(
						func(tsi TypeSpecInfo) patchReplacerData {
							if tsi.Err != nil && firstErr == nil {
								firstErr = tsi.Err
							}
							mt, ok := parseUndType(tsi.TypeInfo, nil, importMap, ConversionMethodsSet{})
							return patchReplacerData{tsi, mt, ok}
						},
						seq,
					),
				)
				if firstErr != nil {
					if !yield(patchTypesGenerationData{}, firstErr) {
						return
					}
					continue
				}

				df, err := replaceNonUndTypes(
					file,
					pkg.Fset,
					importMap,
					replaceData,
				)
				if err != nil {
					if !yield(patchTypesGenerationData{}, firstErr) {
						return
					}
					// skip this file
					continue
				}

				if !yield(patchTypesGenerationData{
					df:       df,
					filename: pkg.Fset.Position(file.FileStart).Filename,
					typeNames: slices.Collect(
						xiter.Map(
							func(data patchReplacerData) string { return data.mt.Name },
							slices.Values(replaceData),
						),
					),
				}, nil) {
					return
				}

			}
		}
	}
}

type patchReplacerData struct {
	tsi TypeSpecInfo
	mt  RawMatchedType
	ok  bool
}

func (p patchReplacerData) Field(fieldName string) (MatchedField, bool) {
	if !p.ok {
		return MatchedField{}, false
	}
	idx := slices.IndexFunc(p.mt.Field, func(mf MatchedField) bool { return mf.Name == fieldName })
	if idx < 0 {
		return MatchedField{}, false
	}
	return p.mt.Field[idx], true
}

func replaceNonUndTypes(
	f *ast.File,
	fset *token.FileSet,
	imports importDecls,
	targets []patchReplacerData,
) (df *dst.File, err error) {
	dec := decorator.NewDecorator(fset)
	df, err = dec.DecorateFile(f)
	if err != nil {
		return
	}
	for _, target := range targets {
		dts, ok := dec.Dst.Nodes[target.tsi.TypeSpec].(*dst.TypeSpec)
		if !ok {
			continue
		}
		dstutil.Apply(
			dts.Type,
			func(c *dstutil.Cursor) bool {
				node := c.Node()
				switch field := node.(type) {
				default:
					return true
				case *dst.Field:
					if f, ok := target.Field(field.Names[0].Name); ok && slices.Contains(UndTargetTypes, f.Type) {
						switch f.Type {
						case UndTargetTypeOption:
							c.Replace(&dst.Field{
								Names: field.Names,
								Type: &dst.IndexExpr{
									X:     imports.DstExpr(UndTargetTypesSliceUnd.ImportPath, UndTargetTypesSliceUnd.Name),
									Index: field.Type.(*dst.IndexExpr).Index,
								},
								Tag:  field.Tag,
								Decs: field.Decs,
							})
						case UndTargetTypesUnd:
						case UndTargetTypeElastic:
						case UndTargetTypesSliceUnd:
						case UndTargetTypesSliceElastic:
						}
					} else {
						c.Replace(
							&dst.Field{
								Names: field.Names,
								Type: &dst.IndexExpr{
									X:     imports.DstExpr(UndTargetTypesSliceUnd.ImportPath, UndTargetTypesSliceUnd.Name),
									Index: field.Type,
								},
								Tag:  field.Tag,
								Decs: field.Decs,
							},
						)
					}
					return false
				}
			},
			nil,
		)
	}

	return df, nil
}
