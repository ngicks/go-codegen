package matcher

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"gotest.tools/v3/assert"
)

var (
	noCopySrc = `package main

import (
	"sync"
)

type A struct {
	mu *sync.Mutex
}

type B struct{
	locker sync.Locker
}

type C struct{
	sync.Locker
}
`
)

func parseStringSource(src string) (*token.FileSet, *ast.File, *types.Package) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "hello.go", src, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(err)
	}
	conf := &types.Config{
		Importer: importer.Default(),
	}
	pkg := types.NewPackage("hello", "main")
	chk := types.NewChecker(conf, fset, pkg, nil)
	err = chk.Files([]*ast.File{f})
	if err != nil {
		panic(err)
	}
	return fset, f, pkg
}

func TestIsNoCopy(t *testing.T) {
	_, _, pkg := parseStringSource(noCopySrc)
	a := pkg.Scope().Lookup("A")
	assert.Assert(t, !IsNoCopy(a.Type()))
	assert.Assert(t, !IsNoCopy(a.Type().Underlying()))
	assert.Assert(t, IsNoCopy(a.Type().Underlying().(*types.Struct).Field(0).Type()))
	b := pkg.Scope().Lookup("B")
	assert.Assert(t, IsNoCopy(b.Type().Underlying().(*types.Struct).Field(0).Type()))
	c := pkg.Scope().Lookup("C")
	assert.Assert(t, IsNoCopy(c.Type()))
}
