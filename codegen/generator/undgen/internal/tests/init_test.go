package tests

import (
	"os"
	"slices"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/ngicks/go-codegen/codegen/codegen"
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
	"golang.org/x/tools/go/packages"
)

var (
	compareOptionStringSlice = gocmp.Comparer(func(i, j []option.Option[string]) bool {
		return option.Options[string](i).Equal(option.Options[string](j))
	})
	compareOptionOptionStringSlice = gocmp.Comparer(func(i, j option.Option[[]option.Option[string]]) bool {
		return i.EqualFunc(j, func(i, j []option.Option[string]) bool {
			return option.Options[string](i).Equal(option.Options[string](j))
		})
	})
	compareUndStringSlice = gocmp.Comparer(func(i, j und.Und[[]string]) bool {
		return i.EqualFunc(j, func(i, j []string) bool { return slices.Equal(i, j) })
	})
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

func ptr[T any](t T) *T {
	return &t
}
