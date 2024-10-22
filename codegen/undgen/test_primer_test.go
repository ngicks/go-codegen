package undgen

import (
	"golang.org/x/tools/go/packages"
)

var (
	testdataPackages    []*packages.Package
	patchtargetPackages []*packages.Package
	validatorPackages   []*packages.Package
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
	testdataPackages, err = packages.Load(cfg, "./testdata/targettypes/...")
	if err != nil {
		panic(err)
	}
	patchtargetPackages, err = packages.Load(cfg, "./testdata/patchtarget/...")
	if err != nil {
		panic(err)
	}
	validatorPackages, err = packages.Load(cfg, "./testdata/validatortarget/...")
	if err != nil {
		panic(err)
	}
}
