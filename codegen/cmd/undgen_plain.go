/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"github.com/ngicks/go-codegen/codegen/generator/undgen"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

func init() {
	fset := undgenPlainCmd.Flags()
	commonFlags(fset, true)
	undgenCmd.AddCommand(undgenPlainCmd)
}

// undgenPlainCmd represents the patch command
var undgenPlainCmd = &cobra.Command{
	Use:   "plain",
	Short: "undgen-plain generates plain types and conversion methods from target types.",
	Long:  `undgen-plain generates plain types and conversion methods from target types.`, // TODO improve
	RunE: runCommand(
		"undgen plain",
		".und_plain",
		true,
		func(
			cmd *cobra.Command,
			writer *suffixwriter.Writer,
			verbose bool,
			pkgs []*packages.Package,
			args []string,
		) error {
			return undgen.GeneratePlain(writer, verbose, pkgs, undgen.ConstUnd.Imports)
		},
	),
}
