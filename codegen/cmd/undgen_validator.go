/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"github.com/ngicks/go-codegen/codegen/generator/undgen"
	"github.com/ngicks/go-codegen/codegen/pkg/suffixwriter"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

func init() {
	fset := undgenValidatorCmd.Flags()
	commonFlags(undgenValidatorCmd, fset, true)
	undgenCmd.AddCommand(undgenValidatorCmd)
}

// undgenPatchCmd represents the patch command
var undgenValidatorCmd = &cobra.Command{
	Use:   "validator [flags]",
	Short: "undgen-validator generates validator method on target types.",
	Long:  `undgen-validator generates validator method on target types.`,
	RunE: runCommand(
		"undgen validator",
		".und_validator",
		true,
		func(
			cmd *cobra.Command,
			writer *suffixwriter.Writer,
			verbose bool,
			pkgs []*packages.Package,
			args []string,
		) error {
			return undgen.GenerateValidator(writer, verbose, pkgs, undgen.ConstUnd.Imports)
		},
	),
}
