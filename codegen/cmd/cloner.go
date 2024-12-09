/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"log/slog"

	"github.com/ngicks/go-codegen/codegen/generator/cloner"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

func init() {
	fset := clonerCmd.Flags()
	commonFlags(fset, true)
	rootCmd.AddCommand(clonerCmd)
}

// clonerCmd represents the cloner command
var clonerCmd = &cobra.Command{
	Use:   "cloner [flags] types...",
	Short: "cloner generates clone methods on target types.",
	Long: `cloner generates clone methods on target types. 

cloner command generates 2 kinds of clone methods

1) Clone() for non-generic types
2) CloneFunc() for generic types.

CloneFunc requires clone function for each type parameters.
You can use github.com/ngicks/go-codegen/pkg/cloneruntime for some help.
`,
	RunE: runCommand(
		"cloner",
		".clone",
		true,
		func(
			cmd *cobra.Command,
			writer *suffixwriter.Writer,
			verbose bool,
			pkgs []*packages.Package,
			args []string,
		) error {
			cfg := &cloner.Config{}
			if verbose {
				cfg.Logger = slog.Default()
			}
			return cfg.Generate(cmd.Context(), writer, pkgs, []imports.TargetImport{})
		},
	),
}
