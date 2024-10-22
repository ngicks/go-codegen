package pkgsutil

import (
	"errors"
	"fmt"
	"go/ast"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"strings"

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

func RemoveSuffixedFiles(pkgs []*packages.Package, cwd, suffix string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		if cwd == "" {
			var err error
			cwd, err = os.Getwd()
			if err != nil {
				yield("", fmt.Errorf("getwd: %w", err))
				return
			}
		}
		for pkg, seq := range EnumeratePackages(pkgs) {
			for file := range seq {
				filename := pkg.Fset.Position(file.FileStart).Filename

				rel, err := filepath.Rel(cwd, filename)
				if err != nil {
					yield("", err)
					return
				}

				if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
					yield(rel, fmt.Errorf("not under cwd: %q", rel))
					return
				}

				s, err := os.Lstat(rel)
				if err != nil {
					if !yield(rel, fmt.Errorf("stat %q: %w", rel, err)) {
						return
					}
					continue
				}

				if !s.Mode().IsRegular() {
					if !yield(rel, fmt.Errorf("ignoring non regular file: %q", rel)) {
						return
					}
					continue
				}

				withoutExt, _ := strings.CutSuffix(rel, filepath.Ext(rel))
				if strings.HasSuffix(withoutExt, suffix) {
					err = os.Remove(rel)
					if errors.Is(err, fs.ErrNotExist) {
						err = nil
					}
					if err != nil {
						err = fmt.Errorf("remove %q: %w", rel, err)
					}
					if !yield(rel, err) {
						return
					}
				}
			}
		}
	}
}
