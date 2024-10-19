/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-codegen/codegen/undgen"
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

		types := fset.Args()
		if len(types) == 0 {
			return fmt.Errorf("no types are specified")
		}

		ctx := cmd.Context()

		cfg := &packages.Config{
			Mode: packages.NeedName |
				packages.NeedSyntax |
				packages.NeedTypesInfo |
				packages.NeedTypesSizes,
			Context: ctx,
			Dir:     dir,
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

		writer := suffixwriter.New(".und_patch", suffixwriter.WithCwd(dir))
		return undgen.GeneratePatcher(writer, targetPkg[0], undgen.ConstUnd.Imports, types...)
	},
}

func init() {
	fset := undgenPatchCmd.Flags()
	fset.StringP("dir", "d", "", "directory under which target package is located. If empty cwd will be used.")
	fset.StringP("pkg", "p", "", "target package name. relative to dir. only single package will be used so if should not be ./...")
	_ = undgenPatchCmd.MarkFlagRequired("pkg")
	undgenCmd.AddCommand(undgenPatchCmd)
}
