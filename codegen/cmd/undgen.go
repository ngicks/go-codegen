/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// undgenCmd represents the undgen command
var undgenCmd = &cobra.Command{
	Use:   "undgen",
	Short: "[DEPRECATED] undgen generates code for types that contain those defined in github.com/ngicks/und. see subcommands",
	Long: `⚠️  DEPRECATED: This command is deprecated and will be removed in a future version.
Please use the new two-phase workflow instead:
  1. codegen automark ./... -g und:patch -g und:plain -g und:validator
  2. codegen autoimpl ./...

For more information, run: codegen automark --help

---

undgen holds subcommands that generates types and methods on them based on types that contain those defined in github.com/ngicks/und.
`,
}

func init() {
	rootCmd.AddCommand(undgenCmd)
}
