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
var undgenPatchCmd = &cobra.Command{
	Use:   "patch [flags] types...",
	Short: "undgen-patch generates patcher types based on target types.",
	Long: `undgen-patch generates patcher types base on target types defined in a target package.

A patch is basically same type as target but name is suffixed with Patch and all fields are wrapped in sliceund.Und[T].
If each field that is already a und type, namely one of und.Und[T], sliceund.Und[T], elastic.Elastic[T], sliceelastic.Elastic[T].
option.Option[T] will be widened to be sliceund.Und[T].

All generated code will be written along the source code the target type is defined.
Generated files are suffixed with und_patch before file extension, i.e. <original_source_filename>.und_patch.go.
`,
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

		pkg, err := fset.GetString("pkg")
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

		types := fset.Args()
		if len(types) == 0 {
			return fmt.Errorf("no types are specified")
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

		targetPkg, err := packages.Load(cfg, pkg)
		if err != nil {
			return err
		}

		if len(targetPkg) == 0 {
			return fmt.Errorf("package not loaded: wrong import pattern?")
		}
		if len(targetPkg) > 1 {
			return fmt.Errorf("2 or more packages are loaded: must be only one")
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
			fmt.Printf("generating for: %#v\n", types)
			writerOpts = append(
				writerOpts,
				suffixwriter.WithLogf(
					func(format string, args ...any) {
						fmt.Printf(format, args...)
					},
				),
			)
		}
		suffix := ".und_patch"
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
			pkgsutil.RemoveSuffixedFiles([]*packages.Package{targetPkg[0]}, dir, suffix),
		)
		if err != nil {
			return err
		}
		return undgen.GeneratePatcher(writer, verbose, targetPkg[0], undgen.ConstUnd.Imports, types...)
	},
}

func init() {
	fset := undgenPatchCmd.Flags()
	fset.StringP("dir", "d", "", "directory under which target package is located. If empty cwd will be used.")
	fset.StringP("pkg", "p", "", "target package name. relative to dir. only single package will be used so if should not be ./...")
	fset.BoolP("verbose", "v", false, "verbose logs")
	fset.Bool("test", false, "only writes test")
	_ = undgenPatchCmd.MarkFlagRequired("pkg")
	undgenCmd.AddCommand(undgenPatchCmd)
}
