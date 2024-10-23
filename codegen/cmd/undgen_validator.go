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
var undgenValidatorCmd = &cobra.Command{
	Use:   "validator [flags]",
	Short: "undgen-validator generates validator method on target types.",
	Long:  `undgen-validator generates validator method on target types.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fset := cmd.Flags()
		dir, pkg, verbose, dry, err := undCommonOpts(fset, true)
		if err != nil {
			return err
		}

		ctx := cmd.Context()

		targetPkgs, err := loadPkgs(ctx, dir, pkg, true, verbose)
		if err != nil {
			return err
		}

		const suffix = ".und_validator"
		writer, deferred := createWriter(dir, suffix, verbose, dry)
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

		return undgen.GenerateValidator(writer, verbose, targetPkgs, undgen.ConstUnd.Imports)
	},
}

func init() {
	fset := undgenValidatorCmd.Flags()
	undCommonFlags(fset, true)
	undgenCmd.AddCommand(undgenValidatorCmd)
}
