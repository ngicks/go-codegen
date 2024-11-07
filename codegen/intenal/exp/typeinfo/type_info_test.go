package typeinfo

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"github.com/ngicks/go-iterator-helper/hiter"
	"golang.org/x/tools/go/packages"
)

var (
	targetPackages []*packages.Package
)

func init() {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedExportFile |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes |
			packages.NeedModule |
			packages.NeedEmbedFiles |
			packages.NeedEmbedPatterns,
		// Logf: func(format string, args ...interface{}) {
		// 	fmt.Printf("log: "+format, args...)
		// 	fmt.Println()
		// },
	}
	var err error
	targetPackages, err = packages.Load(cfg, "./target/...")
	if err != nil {
		panic(err)
	}
}

func Test_target(t *testing.T) {
	var fooObj types.Object
PKG:
	for _, pkg := range targetPackages {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				if genDecl.Tok != token.TYPE {
					continue
				}
				for _, spec := range genDecl.Specs {
					ts := spec.(*ast.TypeSpec)
					if ts.Name != nil && ts.Name.Name == "Foo" {
						fooObj = pkg.TypesInfo.Defs[ts.Name]
						break PKG
					}
				}
			}
		}
	}

	mset := types.NewMethodSet(fooObj.Type())
	for i, sel := range hiter.AtterAll(mset) {
		t.Logf("%d: %s", i, sel.Obj().Name())
		// type_info_test.go:70: 0: MethodOnNonPointer
	}

	mset = types.NewMethodSet(types.NewPointer(fooObj.Type().(*types.Named)))
	for i, sel := range hiter.AtterAll(mset) {
		t.Logf("%d: %s", i, sel.Obj().Name())
		// type_info_test.go:75: 0: MethodOnNonPointer
		// type_info_test.go:75: 1: MethodOnPointer
	}
}
