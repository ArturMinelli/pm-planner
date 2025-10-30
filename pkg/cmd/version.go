package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = ""
	BuiltAt   = ""
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show pm version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pm %s", Version)
		if Commit != "" {
			fmt.Printf(" (%s)", Commit)
		}
		if BuiltAt != "" {
			fmt.Printf(" built %s", BuiltAt)
		}
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
