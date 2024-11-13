package undgen

import (
	"log/slog"
	"os"

	"golang.org/x/tools/go/packages"
)

var (
	targettypesPackages []*packages.Package
	patchtargetPackages []*packages.Package
	validatorPackages   []*packages.Package
	plaintargetPackages []*packages.Package
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
}

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
		ParseFile: ParseFileIgnoringUndgenGeneratedFiles,
		// Logf: func(format string, args ...interface{}) {
		// 	fmt.Printf("log: "+format, args...)
		// 	fmt.Println()
		// },
	}
	var err error
	targettypesPackages, err = packages.Load(cfg, "./internal/targettypes/...")
	if err != nil {
		panic(err)
	}
	patchtargetPackages, err = packages.Load(cfg, "./internal/patchtarget/...")
	if err != nil {
		panic(err)
	}
	validatorPackages, err = packages.Load(cfg, "./internal/validatortarget/...")
	if err != nil {
		panic(err)
	}
	plaintargetPackages, err = packages.Load(cfg, "./internal/plaintarget/...")
	if err != nil {
		panic(err)
	}
}
