/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/ngicks/go-codegen/codegen/astmeta"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/tools/go/packages"
)

// undgenCmd represents the undgen command
var undgenCmd = &cobra.Command{
	Use:   "undgen",
	Short: "undgen generates code for types that contain those defined in github.com/ngicks/und. see subcommands",
	Long: `undgen holds subcommands that generates types and methods on them based on types that contain those defined in github.com/ngicks/und.
`,
}

func init() {
	rootCmd.AddCommand(undgenCmd)
}

func undCommonFlags(fset *pflag.FlagSet, multiplePkg bool) {
	fset.StringP("dir", "d", "", "directory under which target package is located. If empty cwd will be used.")
	if multiplePkg {
		fset.StringArrayP("pkg", "p", nil, "target package name. relative to dir. must start with ./")
	} else {
		fset.StringP("pkg", "p", "", "target package name. relative to dir. specifying 2 or more packages is not allowed")
	}
	fset.BoolP("verbose", "v", false, "verbose logs")
	fset.Bool(
		"ignore-generated",
		false,
		"if set, the type checker ignores ast nodes with comment //codegen:generated. "+
			"Useful for internal debugging. "+
			"You do not need this option.")
	fset.Bool("dry", false, "enables dry run mode. any files will be remove nor generated.")
	_ = undgenPatchCmd.MarkFlagRequired("pkg")
}

func undCommonOpts(fset *pflag.FlagSet, multiplePkg bool) (dir string, pkg []string, verbose bool, ignoreGenerated bool, dry bool, err error) {
	dir, err = fset.GetString("dir")
	if err != nil {
		return
	}
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			err = fmt.Errorf("Getwd: %w", err)
			return
		}
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return
	}

	if multiplePkg {
		pkg, err = fset.GetStringArray("pkg")
		if err != nil {
			return
		}
	} else {
		var single string
		single, err = fset.GetString("pkg")
		if err != nil {
			return
		}
		pkg = []string{single}
	}

	verbose, err = fset.GetBool("verbose")
	if err != nil {
		return
	}

	ignoreGenerated, err = fset.GetBool("ignore-generated")
	if err != nil {
		return
	}

	dry, err = fset.GetBool("dry")
	if err != nil {
		return
	}

	return
}

func loadPkgs(
	ctx context.Context,
	dir string,
	pkg []string,
	multiplePkg bool,
	verbose bool,
	ignoreGenerated bool,
) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes,
		Context: ctx,
		Dir:     dir,
	}
	if verbose {
		cfg.Logf = func(format string, args ...interface{}) {
			fmt.Printf(format, args...)
		}
	}
	if ignoreGenerated {
		cfg.ParseFile = astmeta.NewParser(cfg.Dir).ParseFile
	}

	targetPkgs, err := packages.Load(cfg, pkg...)
	if err != nil {
		return targetPkgs, err
	}
	if err := pkgsutil.CheckLoadError(targetPkgs); err != nil {
		return targetPkgs, err
	}

	if verbose {
		fmt.Print("\n\n")
	}

	if verbose {
		fmt.Printf("matched packages: len(pkgs) == %d\n", len(targetPkgs))
		for i, pkg := range targetPkgs {
			fmt.Printf("\t%d: %s\n", i, pkg.PkgPath)
		}
	}

	if len(targetPkgs) == 0 {
		return targetPkgs, fmt.Errorf("package not loaded: wrong import pattern?")
	}
	if !multiplePkg && len(targetPkgs) >= 2 {
		return targetPkgs, fmt.Errorf("loaded more than a package: must be single.")
	}

	return targetPkgs, nil
}

func createWriter(dir string, suffix string, subcommand string, verbose bool, dry bool) (writer *suffixwriter.Writer, deferred func()) {
	writerOpts := []suffixwriter.Option{
		suffixwriter.WithCwd(dir),
		suffixwriter.WithPrefix([]byte(
			generationNotice +
				"// to regenerate the code, refer to help by invoking\n" +
				"// go run github.com/ngicks/go-codegen/codegen " + subcommand + " --help\n",
		)),
	}
	if verbose {
		slog.SetDefault(
			slog.New(
				slog.NewTextHandler(
					os.Stdout,
					&slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				),
			),
		)

		writerOpts = append(
			writerOpts,
			suffixwriter.WithLogf(
				func(format string, args ...any) {
					fmt.Printf(format, args...)
				},
			),
		)
	}

	writer = suffixwriter.New(suffix, writerOpts...)
	deferred = func() {}

	if dry {
		testWriter := suffixwriter.NewTestWriter(suffix, writerOpts...)
		writer = testWriter.Writer
		deferred = func() {
			results := testWriter.Results()
			fmt.Printf("generated result:\n")
			for _, k := range slices.Sorted(maps.Keys(results)) {
				result := results[k]
				fmt.Printf("%q:\n\n%s\n\n\nj", k, result)
			}
		}
	}

	return
}
