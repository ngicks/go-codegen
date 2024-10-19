package undgen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"slices"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-iterator-helper/hiter"
	"golang.org/x/tools/go/packages"
)

type TypeSpecInfo struct {
	Pos      int
	Pkg      *packages.Package
	File     *ast.File
	GenDecl  *ast.GenDecl
	TypeSpec *ast.TypeSpec
	TypeInfo types.Object
	Err      error
}

// enumerateTypeSpec returns an iterator over every *ast.TypeSpec and corresponding types.Object.
func enumerateTypeSpec(pkgs []*packages.Package) iter.Seq2[*packages.Package, iter.Seq2[*ast.File, iter.Seq[TypeSpecInfo]]] {
	return func(yield func(*packages.Package, iter.Seq2[*ast.File, iter.Seq[TypeSpecInfo]]) bool) {
		for _, pkg := range pkgs {
			if !yield(pkg, func(yield func(*ast.File, iter.Seq[TypeSpecInfo]) bool) {
				for _, file := range pkg.Syntax {
					if !yield(file, func(yield func(TypeSpecInfo) bool) {
						// type decl position per file.
						// incremented at every occurrence of type decl.
						// (`type` keyword itself does not count)
						var pos int
						for _, dec := range file.Decls {
							genDecl, ok := dec.(*ast.GenDecl)
							if !ok {
								continue
							}
							if genDecl.Tok != token.TYPE {
								continue
							}

							direction, _, err := ParseUndComment(genDecl.Doc)
							if err != nil {
								if !yield(TypeSpecInfo{
									Pos:     pos,
									Pkg:     pkg,
									File:    file,
									GenDecl: genDecl,
									Err: fmt.Errorf(
										"in file %q at %dth type spec(maybe group): bad comment: %w",
										file.Name, pos, err,
									),
								}) {
									return
								}
								continue
							}

							if direction.MustIgnore() {
								continue
							}

							for _, s := range genDecl.Specs {
								currentPos := pos
								pos++

								ts := s.(*ast.TypeSpec)
								direction, _, err := ParseUndComment(ts.Doc)
								if err != nil {
									if !yield(TypeSpecInfo{
										Pos:      currentPos,
										Pkg:      pkg,
										File:     file,
										GenDecl:  genDecl,
										TypeSpec: ts,
										Err: fmt.Errorf(
											"in file %q at type spec for %q: bad comment: %w",
											file.Name, ts.Name, err,
										),
									}) {
										return
									}
									continue
								}

								if direction.MustIgnore() {
									continue
								}

								obj := pkg.TypesInfo.Defs[ts.Name]
								if obj == nil {
									continue
								}
								if !yield(TypeSpecInfo{
									Pos:      currentPos,
									Pkg:      pkg,
									File:     file,
									GenDecl:  genDecl,
									TypeSpec: ts,
									TypeInfo: obj,
								}) {
									return
								}
							}
						}
					}) {
						return
					}
				}
			}) {
				return
			}
		}
	}
}

func findTypes(pkg *packages.Package, typeNames ...string) iter.Seq2[*packages.Package, iter.Seq2[*ast.File, iter.Seq[TypeSpecInfo]]] {
	return filterEnumerateTypeSpec(
		[]*packages.Package{pkg},
		nil,
		nil,
		func(tsi TypeSpecInfo) bool {
			return slices.Contains(typeNames, tsi.TypeSpec.Name.Name)
		},
	)
}

func filterEnumerateTypeSpec(
	pkgs []*packages.Package,
	pkgFilter func(*packages.Package) bool,
	fileFilter func(*ast.File) bool,
	elemFilter func(TypeSpecInfo) bool,
) iter.Seq2[*packages.Package, iter.Seq2[*ast.File, iter.Seq[TypeSpecInfo]]] {
	return func(yield func(*packages.Package, iter.Seq2[*ast.File, iter.Seq[TypeSpecInfo]]) bool) {
		for pkg, seq := range enumerateTypeSpec(pkgs) {
			if pkgFilter != nil && !pkgFilter(pkg) {
				continue
			}
			if !yield(pkg, func(yield func(*ast.File, iter.Seq[TypeSpecInfo]) bool) {
				for f, seq := range seq {
					if fileFilter != nil && !fileFilter(f) {
						continue
					}
					if !yield(f, func(yield func(TypeSpecInfo) bool) {
						for ti := range seq {
							if elemFilter != nil && !elemFilter(ti) {
								continue
							}
							if !yield(ti) {
								return
							}
						}
					}) {
						return
					}
				}
			}) {
				return
			}
		}
	}
}

type replaceData struct {
	filename    string
	af          *ast.File
	dec         *decorator.Decorator
	df          *dst.File
	importMap   importDecls
	perTypeData []replacerPerTypeData
	typeNames   []string
	typeSpecs   []*dst.TypeSpec
}

type replacerPerTypeData struct {
	tsi TypeSpecInfo
	mt  RawMatchedType
	ok  bool
}

func (p replacerPerTypeData) Field(fieldName string) (MatchedField, bool) {
	if !p.ok {
		return MatchedField{}, false
	}
	idx := slices.IndexFunc(p.mt.Field, func(mf MatchedField) bool { return mf.Name == fieldName })
	if idx < 0 {
		return MatchedField{}, false
	}
	return p.mt.Field[idx], true
}

func generatorSeq(pkg *packages.Package, imports []TargetImport, targetTypeNames ...string) iter.Seq2[replaceData, error] {
	return func(yield func(replaceData, error) bool) {
		for pkg, seq := range findTypes(pkg, targetTypeNames...) {
			for file, seq := range seq {
				dec := decorator.NewDecorator(pkg.Fset)
				df, err := dec.DecorateFile(file)
				if err != nil {
					if !yield(replaceData{}, err) {
						return
					}
					continue
				}

				importMap := parseImports(file.Imports, imports)
				var (
					perTypeData []replacerPerTypeData
					typeNames   []string
					typeSpecs   []*dst.TypeSpec
				)
				err = hiter.TryForEach(
					func(tsi TypeSpecInfo) {
						mt, ok := parseUndType(tsi.TypeInfo, nil, importMap, ConversionMethodsSet{})
						perTypeData = append(perTypeData, replacerPerTypeData{tsi, mt, ok})
						typeNames = append(typeNames, mt.Name)
						typeSpecs = append(typeSpecs, dec.Dst.Nodes[tsi.TypeSpec].(*dst.TypeSpec))
					},
					hiter.Divide(
						func(tsi TypeSpecInfo) (TypeSpecInfo, error) {
							return tsi, tsi.Err
						},
						seq,
					),
				)
				if err != nil {
					if !yield(replaceData{}, err) {
						return
					}
					continue
				}

				addMissingImports(df, importMap)

				if !yield(
					replaceData{
						filename:    pkg.Fset.Position(file.FileStart).Filename,
						af:          file,
						dec:         dec,
						df:          df,
						importMap:   importMap,
						perTypeData: perTypeData,
						typeNames:   typeNames,
						typeSpecs:   typeSpecs,
					},
					nil,
				) {
					return
				}
			}
		}
	}
}
