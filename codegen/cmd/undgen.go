/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// undgenCmd represents the undgen command
var undgenCmd = &cobra.Command{
	Use:   "undgen",
	Short: "undgen generates code for types that contain those defined in github.com/ngicks/und. see subcommands",
	Long: `undgen holds subcommands that generates types and methods on them based on types that contain those defined in github.com/ngicks/und.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("undgen called")
	},
}

func init() {
	rootCmd.AddCommand(undgenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// undgenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// undgenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
