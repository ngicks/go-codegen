package tests

import (
	"os"
	"slices"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/ngicks/go-codegen/codegen/pkg/astutil"
	"github.com/ngicks/und"
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
	"golang.org/x/tools/go/packages"
)

var (
	// Advertise Equal functions for basic types, keep complex nested types
	compareOptionInt = gocmp.Comparer(func(i, j option.Option[int]) bool {
		return option.Equal(i, j)
	})
	compareOptionString = gocmp.Comparer(func(i, j option.Option[string]) bool {
		return option.Equal(i, j)
	})
	compareUndInt = gocmp.Comparer(func(i, j und.Und[int]) bool {
		return und.Equal(i, j)
	})
	compareUndString = gocmp.Comparer(func(i, j und.Und[string]) bool {
		return und.Equal(i, j)
	})
	compareElasticString = gocmp.Comparer(func(i, j elastic.Elastic[string]) bool {
		return elastic.Equal(i, j)
	})
	// Specific comparators for complex nested types from error messages
	compareUndOptionArray = gocmp.Comparer(func(i, j und.Und[[2]option.Option[string]]) bool {
		return und.Equal(i, j)
	})
	compareOptionOptions = gocmp.Comparer(func(i, j option.Option[option.Options[string]]) bool {
		return i.EqualFunc(j, func(a, b option.Options[string]) bool {
			return option.EqualOptions(a, b)
		})
	})
	compareSliceUndString = gocmp.Comparer(func(i, j sliceund.Und[string]) bool {
		return sliceund.Equal(i, j)
	})
	compareSliceElasticString = gocmp.Comparer(func(i, j sliceelastic.Elastic[string]) bool {
		return sliceelastic.Equal(i, j)
	})
	compareOptionOptionArray = gocmp.Comparer(func(i, j option.Option[[2]option.Option[string]]) bool {
		return i.EqualFunc(j, func(a, b [2]option.Option[string]) bool {
			return option.Equal(a[0], b[0]) && option.Equal(a[1], b[1])
		})
	})
	compareOptionStringArray3 = gocmp.Comparer(func(i, j option.Option[[3]string]) bool {
		return option.Equal(i, j)
	})
	compareOptionStringSlice = gocmp.Comparer(func(i, j []option.Option[string]) bool {
		return option.EqualOptions(i, j)
	})
	compareOptionOptionStringSlice = gocmp.Comparer(func(i, j option.Option[[]option.Option[string]]) bool {
		return i.EqualFunc(j, func(i, j []option.Option[string]) bool {
			return option.EqualOptions(i, j)
		})
	})
	compareUndStringSlice = gocmp.Comparer(func(i, j und.Und[[]string]) bool {
		return i.EqualFunc(j, func(i, j []string) bool { return slices.Equal(i, j) })
	})
	compareOptionEmpty = gocmp.Comparer(func(i, j option.Option[conversion.Empty]) bool {
		return i.EqualFunc(j, func(a, b conversion.Empty) bool {
			return len(a) == len(b)
		})
	})
	compareOptionStruct = gocmp.Comparer(func(i, j option.Option[*struct{}]) bool {
		return i.EqualFunc(j, func(a, b *struct{}) bool {
			return (a == nil && b == nil) || (a != nil && b != nil)
		})
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
		ParseFile: astutil.NewParser("../testtargets").ParseFile,
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
