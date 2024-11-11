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
	targetPackages, builtinsPackages []*packages.Package
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
	loadPanicking := func(pat ...string) []*packages.Package {
		pkgs, err := packages.Load(cfg, pat...)
		if err != nil {
			panic(err)
		}
		return pkgs
	}
	targetPackages = loadPanicking("./target/...")
	builtinsPackages = loadPanicking("./builtins/...")
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

	mset = types.NewMethodSet(types.NewPointer(fooObj.Type()))
	for i, sel := range hiter.AtterAll(mset) {
		t.Logf("%d: %s", i, sel.Obj().Name())
		// type_info_test.go:75: 0: MethodOnNonPointer
		// type_info_test.go:75: 1: MethodOnPointer
	}

}

func Test_builtins(t *testing.T) {
	pkg := builtinsPackages[0]

	{
		obj := pkg.Types.Scope().Lookup("Aaaa")
		named := obj.Type().(*types.Named)
		t.Logf("name: %s", named.Obj().Name())
		t.Logf("pkgPath: %s", named.Obj().Pkg().Path())
	}
	{
		obj := pkg.Types.Scope().Lookup("BuiltIns")
		named := obj.Type().(*types.Named)
		t.Logf("name: %s", named.Obj().Name())
		t.Logf("pkgPath: %s", named.Obj().Pkg().Path())
		st := named.Underlying().(*types.Struct)
		for i := range st.NumFields() {
			f := st.Field(i)
			t.Logf("%v", f.Type())
			switch x := f.Type().(type) {
			case *types.Alias:
				t.Logf("alias: obj:%v", x.Obj())
			case *types.Array:
			case *types.Basic:
			case *types.Chan:
			case *types.Interface:
			case *types.Map:
			case *types.Named:
				t.Logf("named: obj:%v", x.Obj())
				t.Logf("named: pkg:%v", x.Obj().Pkg())
			case *types.Pointer:
			case *types.Signature:
			case *types.Slice:
			case *types.Struct:
			case *types.Tuple:
			case *types.TypeParam:
			case *types.Union:
			}
		}
	}
}
