package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
)

func main() {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(
		fset,
		"./internal/enumlike",
		func(fi fs.FileInfo) bool {
			return fi.Mode().IsRegular() && !strings.HasSuffix(fi.Name(), "_test.go")
		},
		parser.ParseComments,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", pkgs)

	pkg := pkgs["enumlike"]

	for _, f := range pkg.Files {
		for _, c := range f.Comments {
			fmt.Printf("%s:%s\n", f.Name, c.Text())
		}
	}

}
