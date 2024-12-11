package generationtests

import (
	"os"
	"slices"

	"github.com/ngicks/go-codegen/codegen/codegen"
	"golang.org/x/tools/go/packages"
)

var (
	excludes    = []string{"implementor"}
	testTargets map[string][]*packages.Package
)

func init() {
	testTargets = make(map[string][]*packages.Package)
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes,
		Dir:       "../testtargets",
		ParseFile: codegen.NewParser("../testtargets").ParseFile,
	}
	var err error
	dirents, err := os.ReadDir("../testtargets")
	if err != nil {
		panic(err)
	}
	for _, dirent := range dirents {
		if !dirent.IsDir() || slices.Contains(excludes, dirent.Name()) {
			continue
		}
		name := dirent.Name()
		pkgs, err := packages.Load(cfg, "./"+name+"/...")
		if err != nil {
			panic(err)
		}
		testTargets[name] = pkgs
	}
}
