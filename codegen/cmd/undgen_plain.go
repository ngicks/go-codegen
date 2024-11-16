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

// undgenPlainCmd represents the patch command
var undgenPlainCmd = &cobra.Command{
	Use:   "plain",
	Short: "undgen-plain generates plain types and conversion methods from target types.",
	Long:  `undgen-plain generates plain types and conversion methods from target types.`, // TODO improve
	RunE: func(cmd *cobra.Command, args []string) error {
		fset := cmd.Flags()

		dir, pkg, verbose, ignoreGenerated, dry, err := undCommonOpts(fset, true)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Printf("running: undgen plain\n\n\n")
		}
		ctx := cmd.Context()

		if verbose {
			fmt.Printf("loading: %#v\n", pkg)
		}
		targetPkgs, err := loadPkgs(ctx, dir, pkg, true, verbose, ignoreGenerated)
		if err != nil {
			return err
		}

		const suffix = ".und_plain"
		writer, deferred := createWriter(dir, suffix, "plain", verbose, dry)
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

		return undgen.GeneratePlain(writer, verbose, targetPkgs, undgen.ConstUnd.Imports)
	},
}

func init() {
	fset := undgenPlainCmd.Flags()
	undCommonFlags(fset, true)
	undgenCmd.AddCommand(undgenPlainCmd)
}
