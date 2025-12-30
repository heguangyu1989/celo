package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func GetBuildInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "show build info",
		RunE:  runBuildInfoCmd,
	}
	return cmd
}

func runBuildInfoCmd(cmd *cobra.Command, args []string) error {
	info, _ := debug.ReadBuildInfo()
	fmt.Println("======BUILD INFO======")
	fmt.Println(info)
	return nil
}
