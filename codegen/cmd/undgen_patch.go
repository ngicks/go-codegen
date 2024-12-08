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
	fset := undgenPatchCmd.Flags()
	commonFlags(fset, false)
	undgenCmd.AddCommand(undgenPatchCmd)
}

// undgenPatchCmd represents the patch command
var undgenPatchCmd = &cobra.Command{
	Use:   "patch [flags] types...",
	Short: "undgen-patch generates patcher types based on target types.",
	Long: `undgen-patch generates patcher types base on target types defined in a target package.

The generation target types are specified as cli argument. e.g.

codegen undgen patch --pkg ./path/to/package TypeA TypeB TypeC (...and so on)

or you can even 

codegen undgen patch --pkg ./path/to/package ...

to generate for all types found in the package.

A patch is basically same type as target but name is suffixed with Patch and all fields are wrapped in sliceund.Und[T].
If each field that is already a und type, namely one of und.Und[T], sliceund.Und[T], elastic.Elastic[T], sliceelastic.Elastic[T].
option.Option[T] will be widened to be sliceund.Und[T].

All generated code will be written along the source code in which the target type is defined.
Generated files are suffixed with und_patch before file extension, i.e. <original_source_filename>.und_patch.go.
`,
	RunE: runCommand(
		"undgen patch",
		".und_patch",
		false,
		func(
			cmd *cobra.Command,
			writer *suffixwriter.Writer,
			verbose bool,
			pkgs []*packages.Package,
			args []string,
		) error {
			return undgen.GeneratePatcher(writer, verbose, pkgs[0], undgen.ConstUnd.Imports, args...)
		},
	),
}
