/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"log/slog"

	"github.com/ngicks/go-codegen/codegen/generator/cloner"
	"github.com/ngicks/go-codegen/codegen/pkg/suffixwriter"
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
	chanMake     bool

	funcIgnore   bool
	funcDisallow bool
	funcCopy     bool

	interfaceIgnore bool
	interfaceCopy   bool
)

func init() {
	fset := clonerCmd.Flags()

	commonFlags(clonerCmd, fset, true)

	fset.BoolVar(&noCopyIgnore, "no-copy-ignore", false, "sets global option that ignores no-copy object. Clone methods just simply leave fields zero value.")
	fset.BoolVar(&noCopyDisallow, "no-copy-disallow", false, "sets global option that disallow no-copy object. Types that contain no-copy type fields are not generation target.")
	fset.BoolVar(&noCopyCopy, "no-copy-copy", false, "sets global option that copy pointer of no-copy object. Clone methods copy no-copy object if and only if field is pointer type.")

	fset.BoolVar(&chanIgnore, "chan-ignore", false, "sets global option that ignores channel fields.")
	fset.BoolVar(&chanDisallow, "chan-disallow", false, "sets global option that disallows channel fields")
	fset.BoolVar(&chanCopy, "chan-copy", false, "sets global option that copies channel fields")
	fset.BoolVar(&chanMake, "chan-make", false, "sets global option that makes new channel. Clone methods also copy the capacity of input channels.")

	fset.BoolVar(&funcIgnore, "func-ignore", false, "sets global option that ignores func fields. func literal or named function type.")
	fset.BoolVar(&funcDisallow, "func-disallow", false, "sets global option that disallow func fields.")
	fset.BoolVar(&funcCopy, "func-copy", false, "sets global option that copies func fields")

	fset.BoolVar(&interfaceIgnore, "interface-ignore", false, "sets global option that ignores interface fields. func literal or named function type.")
	fset.BoolVar(&interfaceCopy, "interface-copy", false, "sets global option that copies interface fields")

	rootCmd.AddCommand(clonerCmd)
}

// clonerCmd represents the cloner command
var clonerCmd = &cobra.Command{
	Use:   "cloner [flags] --pkg ./",
	Short: "cloner generates clone methods on target types.",
	Long: `cloner generates clone methods on target types. 

cloner command generates 2 kinds of clone methods

1) Clone() for non-generic types
2) CloneFunc() for generic types.

CloneFunc requires clone function for each type parameters.

Example:

func (c C[T, U]) CloneFunc(cloneT func(T) T, cloneU func(U) U) C[T, U] {
	// ...
}

The cloner sub command, as other commands do, loads and parses Go source code files
by using "golang.org/x/tools/go/packages".Load
then it examines types defined in them whether if they are clone-able or not.
Multiple packages can be loaded and processed at once.
The type dependency chain is allowed to span across multiple packages and generated Clone method considers it.

The specified package path must be relative to the cwd, which can be changed by --dir option,
to limit the target packages to which the process can write generated code safely.

The clone-able is defines as
1) A struct type which has at least a field of
1-1) basic types or pointer of basic types
1-2) array, slice or map of 1-1).
1-3) channel, noCopy object (types with the Lock method, e.g. sync.Mutex or sync.Locker), func type when the configuration allows copying each of them.
1-4) a type that implements Clone or CloneFunc method
1-5) other 1) types.
2) A named type whose underlying type is array, slice or map of 1), basic types or pointer of basic type.

A field of deeply nested type, for example, []*[5]map[int]string is still considered as clone-able,
since bottom type, string, is a basic therefore clone-able type.
We call parts other than that ([]*[5]map[int]) _route_. And each element of them as _route node_
(i.e. for []*[5]map[int]string, route nodes are slice, pointer, map, in the order. The bottom type is string.)
Only disallowed _route node_ is interface literal. They are ignored silently.

The cloner sub command also allows per-field basis configuration by writing comments associated to it.
For example:

type Foo struct {
	//cloner:copyptr
	NoCopy *sync.Mutex
}

the cloner command generates

func (v Foo) Clone() Foo {
	return Foo{
		NoCopy: v.NoCopy,
	}
}

Without the comment, the cloner command ignores the type Foo since it has no clone-able fields other than that.
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
				cfg.MatcherConfig.NoCopyHandle = cloner.CopyHandleIgnore
			case noCopyDisallow:
				cfg.MatcherConfig.NoCopyHandle = cloner.CopyHandleDisallow
			case noCopyCopy:
				cfg.MatcherConfig.NoCopyHandle = cloner.CopyHandleCopyPointer
			}
			switch {
			case chanIgnore:
				cfg.MatcherConfig.ChannelHandle = cloner.CopyHandleIgnore
			case chanDisallow:
				cfg.MatcherConfig.ChannelHandle = cloner.CopyHandleDisallow
			case chanCopy:
				cfg.MatcherConfig.ChannelHandle = cloner.CopyHandleCopyPointer
			case chanMake:
				cfg.MatcherConfig.ChannelHandle = cloner.CopyHandleMake
			}
			switch {
			case funcIgnore:
				cfg.MatcherConfig.FuncHandle = cloner.CopyHandleIgnore
			case funcDisallow:
				cfg.MatcherConfig.FuncHandle = cloner.CopyHandleDisallow
			case funcCopy:
				cfg.MatcherConfig.FuncHandle = cloner.CopyHandleCopyPointer
			}

			switch {
			case interfaceIgnore:
				cfg.MatcherConfig.InterfaceHandle = cloner.CopyHandleIgnore
			case interfaceCopy:
				cfg.MatcherConfig.InterfaceHandle = cloner.CopyHandleCopyPointer
			}

			return cfg.Generate(cmd.Context(), writer, pkgs)
		},
	),
}
