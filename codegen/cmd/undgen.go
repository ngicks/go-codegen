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
	Short: "undgen generates code for types that contain those defined in github.com/ngicks/und. see subcommands",
	Long: `undgen holds subcommands that generates types and methods on them based on types that contain those defined in github.com/ngicks/und.
`,
}

func init() {
	rootCmd.AddCommand(undgenCmd)
}
