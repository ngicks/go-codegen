package pkgsutil

import (
	"errors"
	"fmt"
	"go/ast"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
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

func RemoveSuffixedFiles(pkgs []*packages.Package, cwd, suffix string, dry bool) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		if cwd == "" {
			var err error
			cwd, err = os.Getwd()
			if err != nil {
				yield("", fmt.Errorf("getwd: %w", err))
				return
			}
		}
		var err error
		cwd, err = filepath.Abs(cwd)
		if err != nil {
			yield("", fmt.Errorf("filepath.Abs: %w", err))
			return
		}
		for pkg, seq := range EnumeratePackages(pkgs) {
			if err := LoadError(pkg); err != nil {
				if !yield("", err) {
					return
				}
				continue
			}
			for file := range seq {
				filename := pkg.Fset.Position(file.FileStart).Filename

				rel, err := filepath.Rel(cwd, filename)
				if err != nil {
					yield("", fmt.Errorf("cwd = %q, filename = %q: %w", cwd, filename, err))
					return
				}

				if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
					yield(filename, fmt.Errorf("not under cwd: %q", rel))
					return
				}

				s, err := os.Lstat(filename)
				if err != nil {
					if !yield(filename, fmt.Errorf("stat %q: %w", filename, err)) {
						return
					}
					continue
				}

				if !s.Mode().IsRegular() {
					if !yield(filename, fmt.Errorf("ignoring non regular file: %q", filename)) {
						return
					}
					continue
				}

				withoutExt, _ := strings.CutSuffix(filename, filepath.Ext(filename))
				if strings.HasSuffix(withoutExt, suffix) {
					if dry {
						if !yield(filename, nil) {
							return
						}
					} else {
						err = os.Remove(filename)
						if errors.Is(err, fs.ErrNotExist) {
							err = nil
						}
						if err != nil {
							err = fmt.Errorf("remove %q: %w", filename, err)
						}
						if !yield(filename, err) {
							return
						}
					}
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
