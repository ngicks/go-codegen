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
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"github.com/ngicks/und/option"
	"golang.org/x/tools/go/packages"
)

type enumTypeSeq iter.Seq2[*packages.Package, iter.Seq2[*ast.File, iter.Seq[TypeSpecInfo]]]

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
func enumerateTypeSpec(pkgs []*packages.Package) enumTypeSeq {
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

func findTypes(pkg *packages.Package, typeNames ...string) enumTypeSeq {
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
) enumTypeSeq {
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
	filename  string
	af        *ast.File
	dec       *decorator.Decorator
	df        *dst.File
	importMap importDecls
	targets   replacerTargets
}

type replacerTargets []replacerTarget

func (t replacerTargets) typeNames() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, tt := range t {
			if !yield(tt.typeName) {
				return
			}
		}
	}
}

func (t replacerTargets) typeSpecs() iter.Seq[*dst.TypeSpec] {
	return func(yield func(*dst.TypeSpec) bool) {
		for _, tt := range t {
			if !yield(tt.typeSpec) {
				return
			}
		}
	}
}

type replacerTarget struct {
	replacerPerTypeData
	typeSpec *dst.TypeSpec
}

type replacerPerTypeData struct {
	tsi      TypeSpecInfo
	mt       option.Option[RawMatchedType]
	typeName string
}

func (p replacerPerTypeData) Field(fieldName string) (MatchedField, bool) {
	if p.mt.IsNone() {
		return MatchedField{}, false
	}
	idx := slices.IndexFunc(p.mt.Value().Field, func(mf MatchedField) bool { return mf.Name == fieldName })
	if idx < 0 {
		return MatchedField{}, false
	}
	return p.mt.Value().Field[idx], true
}

func generatorIter(imports []TargetImport, seq enumTypeSeq) iter.Seq2[replaceData, error] {
	return func(yield func(replaceData, error) bool) {
		for pkg, seq := range seq {
			if err := pkgsutil.LoadError(pkg); err != nil {
				if !yield(replaceData{}, err) {
					return
				}
				continue
			}

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

				var targets replacerTargets
				targets, err = hiter.TryCollect(
					xiter.Map2(
						func(tsi TypeSpecInfo, err error) (replacerTarget, error) {
							if err != nil {
								return replacerTarget{}, err
							}
							mt, ok := parseUndType(tsi.TypeInfo, nil, importMap, ConversionMethodsSet{})
							return replacerTarget{
								replacerPerTypeData{
									tsi,
									option.FromOk(mt, ok),
									mt.Name,
								},
								dec.Dst.Nodes[tsi.TypeSpec].(*dst.TypeSpec),
							}, nil
						},
						hiter.Divide(
							func(tsi TypeSpecInfo) (TypeSpecInfo, error) {
								return tsi, tsi.Err
							},
							seq,
						),
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
						filename:  pkg.Fset.Position(file.FileStart).Filename,
						af:        file,
						dec:       dec,
						df:        df,
						importMap: importMap,
						targets:   targets,
					},
					nil,
				) {
					return
				}
			}
		}
	}
}
