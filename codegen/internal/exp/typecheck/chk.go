package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	fset := token.NewFileSet()

	pkgs := make(map[string][]*ast.File)

	var readFilesRecursive func(dir string)
	readFilesRecursive = func(dir string) {
		dir, err := filepath.Abs(dir)
		if err != nil {
			panic(err)
		}

		dirents, err := os.ReadDir(dir)
		if err != nil {
			panic(err)
		}

		for _, dirent := range dirents {
			direntPath := filepath.Join(dir, dirent.Name())
			if dirent.IsDir() {
				readFilesRecursive(direntPath)
			}
			if dirent.Type().IsRegular() {
				if !strings.HasSuffix(direntPath, ".go") || strings.HasSuffix(direntPath, "_test.go") {
					continue
				}
				bin, err := os.ReadFile(direntPath)
				if err != nil {
					panic(err)
				}
				f, err := parser.ParseFile(fset, direntPath, bin, parser.ParseComments|parser.AllErrors)
				if err != nil {
					panic(err)
				}
				pkgs[dir] = append(pkgs[dir], f)
			}
		}
	}

	readFilesRecursive("./")

	// list files

	fmt.Printf("ast parsing results:----\n\n")
	for _, pkgPath := range slices.Sorted(maps.Keys(pkgs)) {
		files := pkgs[pkgPath]
		fmt.Printf("package:%s\n", pkgPath)
		for _, f := range files {
			fileBasename, _ := strings.CutPrefix(fset.Position(f.FileStart).Filename, pkgPath)
			fileBasename, _ = strings.CutPrefix(fileBasename, "/")
			fmt.Printf("\t%s\n", fileBasename)
		}
	}

	fmt.Printf("\n\ntype checker results:----\n\n")
	for _, pkgPath := range slices.Sorted(maps.Keys(pkgs)) {
		files := pkgs[pkgPath]
		conf := &types.Config{
			Importer: importer.Default(),
			Sizes:    types.SizesFor("gc", "amd64"),
		}
		pkg := types.NewPackage(pkgPath, files[0].Name.Name)
		typeInfo := &types.Info{
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
			Implicits:  make(map[ast.Node]types.Object),
			Instances:  make(map[*ast.Ident]types.Instance),
			Scopes:     make(map[ast.Node]*types.Scope),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
		}
		chk := types.NewChecker(conf, fset, pkg, typeInfo)
		err := chk.Files(files)
		if err != nil {
			fmt.Printf("package check error: %v\n", err)
		} else {
			fmt.Printf("package checked:%s\n", pkgPath)
		}
	}
}
