package pkgsutil

import (
	"fmt"
	"go/ast"
	"iter"
	"slices"
	"strings"

	"github.com/ngicks/go-iterator-helper/hiter"
	"golang.org/x/tools/go/packages"
)

func EnumeratePackages(pkgs []*packages.Package) iter.Seq2[*packages.Package, iter.Seq[*ast.File]] {
	return func(yield func(*packages.Package, iter.Seq[*ast.File]) bool) {
		for _, pkg := range pkgs {
			if !yield(pkg, func(yield func(*ast.File) bool) {
				for _, file := range pkg.Syntax {
					if !yield(file) {
						return
					}
				}
			}) {
				return
			}
		}
	}
}

func EnumerateGenDecls(pkgs []*packages.Package) iter.Seq2[*packages.Package, iter.Seq2[*ast.File, iter.Seq[*ast.GenDecl]]] {
	return func(yield func(*packages.Package, iter.Seq2[*ast.File, iter.Seq[*ast.GenDecl]]) bool) {
		for _, pkg := range pkgs {
			if !yield(pkg, func(yield func(*ast.File, iter.Seq[*ast.GenDecl]) bool) {
				for _, file := range pkg.Syntax {
					if !yield(file, func(yield func(*ast.GenDecl) bool) {
						for _, decl := range file.Decls {
							genDecl, ok := decl.(*ast.GenDecl)
							if !ok {
								continue
							}
							if !yield(genDecl) {
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

func EnumerateFile(pkgs []*packages.Package) iter.Seq[*ast.File] {
	return func(yield func(*ast.File) bool) {
		for _, pkg := range pkgs {
			for _, f := range pkg.Syntax {
				if !yield(f) {
					return
				}
			}
		}
	}
}

func LoadError(pkg *packages.Package) error {
	if len(pkg.Errors) > 0 {
		format, _ := strings.CutSuffix(strings.Repeat("%w,\n", len(pkg.Errors)), ",\n")
		return fmt.Errorf(
			"*packages.Package load error: "+format+"\n",
			slices.Collect(
				hiter.Map(
					func(e packages.Error) any { return e },
					slices.Values(pkg.Errors),
				),
			)...,
		)
	}
	return nil
}

func CheckLoadError(pkgs []*packages.Package) error {
	var errs []any
	for _, pkg := range pkgs {
		if err := LoadError(pkg); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		msg, _ := strings.CutSuffix(strings.Repeat("%w, ", len(errs)), ", ")
		return fmt.Errorf("load error: "+msg, errs...)
	}
	return nil
}
