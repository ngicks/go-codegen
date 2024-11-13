/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-codegen/codegen/undgen"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/spf13/cobra"
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
		dir, pkg, verbose, dry, err := undCommonOpts(fset, false)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Printf("running: undgen patch\n\n\n")
		}
		typeNames := fset.Args()

		ctx := cmd.Context()

		targetPkgs, err := loadPkgs(ctx, dir, pkg, false, verbose)
		if err != nil {
			return err
		}

		const suffix = ".und_patch"
		writer, deferred := createWriter(dir, suffix, "patch", verbose, dry)
		defer deferred()

		err = hiter.TryForEach(
			func(s string) {
				if verbose || dry {
					if dry {
						fmt.Printf("\t[DRY]: removed %q\n", s)
					} else {
						fmt.Printf("\tremoved %q\n", s)
					}
				}
			},
			pkgsutil.RemoveSuffixedFiles(targetPkgs, dir, suffix, dry),
		)
		if err != nil {
			return err
		}
		return undgen.GeneratePatcher(writer, verbose, targetPkgs[0], undgen.ConstUnd.Imports, typeNames...)
	},
}

func init() {
	fset := undgenPatchCmd.Flags()
	undCommonFlags(fset, false)
	undgenCmd.AddCommand(undgenPatchCmd)
}
