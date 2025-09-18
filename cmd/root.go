package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "celo",
	Short: "Efficiency, at speed.",
}

func init() {
	rootCmd.AddCommand(GetBuildInfoCmd())
	rootCmd.AddCommand(GetMD5Cmd())
}

func Execute() error {
	return rootCmd.Execute()
}
