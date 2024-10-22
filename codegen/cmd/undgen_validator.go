/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-codegen/codegen/undgen"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

// undgenPatchCmd represents the patch command
var undgenValidatorCmd = &cobra.Command{
	Use:   "validator [flags]",
	Short: "undgen-validator generates validator method on target types.",
	Long:  `undgen-validator generates validator method on target types.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fset := cmd.Flags()

		dir, err := fset.GetString("dir")
		if err != nil {
			return err
		}
		if dir != "" {
			dir, err = filepath.Abs(dir)
			if err != nil {
				return err
			}
		}

		pkg, err := fset.GetStringArray("pkg")
		if err != nil {
			return err
		}

		verbose, err := fset.GetBool("verbose")
		if err != nil {
			return err
		}

		test, err := fset.GetBool("test")
		if err != nil {
			return err
		}

		ctx := cmd.Context()

		cfg := &packages.Config{
			Mode: packages.NeedName |
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

		targetPkg, err := packages.Load(cfg, pkg...)
		if err != nil {
			return err
		}

		if len(targetPkg) == 0 {
			return fmt.Errorf("package not loaded: wrong import pattern?")
		}

		writerOpts := []suffixwriter.Option{
			suffixwriter.WithCwd(dir),
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
			fmt.Printf("\n\n")
			fmt.Printf("matched packages: len(pkgs) == %d\n", len(targetPkg))
			for _, pkg := range targetPkg {
				fmt.Printf("path=%s\n", pkg.PkgPath)
			}
			writerOpts = append(
				writerOpts,
				suffixwriter.WithLogf(
					func(format string, args ...any) {
						fmt.Printf(format, args...)
					},
				),
			)
		}
		const suffix = ".und_validator"
		writer := suffixwriter.New(suffix, writerOpts...)
		if test {
			testWriter := suffixwriter.NewTestWriter(suffix, writerOpts...)
			writer = testWriter.Writer
			defer func() {
				results := testWriter.Results()
				for _, k := range slices.Sorted(maps.Keys(results)) {
					result := results[k]
					fmt.Printf("%q:\n%s\n\n", k, result)
				}
			}()
		}
		err = hiter.TryForEach(
			func(s string) {
				if verbose {
					fmt.Printf("removed %q\n", s)
				}
			},
			pkgsutil.RemoveSuffixedFiles(targetPkg, dir, suffix),
		)
		if err != nil {
			return err
		}
		return undgen.GenerateValidator(writer, verbose, targetPkg, undgen.ConstUnd.Imports)
	},
}

func init() {
	fset := undgenValidatorCmd.Flags()
	fset.StringP("dir", "d", "", "directory under which target package is located. If empty cwd will be used.")
	fset.StringArrayP("pkg", "p", nil, "target package name. relative to dir. only single package will be used so if should not be ./...")
	fset.BoolP("verbose", "v", false, "verbose logs")
	fset.Bool("test", false, "only writes test")
	_ = undgenValidatorCmd.MarkFlagRequired("pkg")
	undgenCmd.AddCommand(undgenValidatorCmd)
}
