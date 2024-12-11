/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"log/slog"

	"github.com/ngicks/go-codegen/codegen/generator/cloner"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

var (
	noCopyIgnore   bool
	noCopyDisallow bool
	noCopyCopy     bool

	chanIgnore   bool
	chanDisallow bool
	chanCopy     bool
)

func init() {
	fset := clonerCmd.Flags()

	commonFlags(fset, true)

	// TODO: add description

	fset.BoolVar(&noCopyIgnore, "no-copy-ignore", false, "")
	fset.BoolVar(&noCopyDisallow, "no-copy-disallow", false, "")
	fset.BoolVar(&noCopyCopy, "no-copy-copy", false, "")

	fset.BoolVar(&chanIgnore, "chan-ignore", false, "")
	fset.BoolVar(&chanDisallow, "chan-disallow", false, "")
	fset.BoolVar(&chanCopy, "chan-copy", false, "")

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
			cfg := &cloner.Config{
				MatcherConfig: &cloner.MatcherConfig{},
			}
			if verbose {
				cfg.Logger = slog.Default()
			}

			switch {
			case noCopyIgnore:
				cfg.MatcherConfig.NoCopyHandle = cloner.NoCopyHandleIgnore
			case noCopyDisallow:
				cfg.MatcherConfig.NoCopyHandle = cloner.NoCopyHandleDisallow
			case noCopyCopy:
				cfg.MatcherConfig.NoCopyHandle = cloner.NoCopyHandleCopyPointer
			}
			switch {
			case chanIgnore:
				cfg.MatcherConfig.ChannelHandle = cloner.NoCopyHandleIgnore
			case chanDisallow:
				cfg.MatcherConfig.ChannelHandle = cloner.NoCopyHandleDisallow
			case chanCopy:
				cfg.MatcherConfig.ChannelHandle = cloner.NoCopyHandleCopyPointer
			}

			return cfg.Generate(cmd.Context(), writer, pkgs)
		},
	),
}
