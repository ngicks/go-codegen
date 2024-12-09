package main

import (
	"os"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/packages"
)

func main() {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes,
	}
	pkgs, err := packages.Load(cfg, "./target")
	if err != nil {
		panic(err)
	}
	pkg := pkgs[0]

	dec := decorator.NewDecorator(pkg.Fset)
	df, err := dec.DecorateFile(pkg.Syntax[0])
	if err != nil {
		panic(err)
	}

	err = dst.Fprint(os.Stdout, df, nil)
	if err != nil {
		panic(err)
	}
}
