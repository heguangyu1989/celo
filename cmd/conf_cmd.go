package cmd

import (
	"fmt"
	"github.com/heguangyu1989/celo/pkg/config"
	"github.com/heguangyu1989/celo/pkg/p"
	"github.com/spf13/cobra"
)

func GetGenDefaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen-default",
		Short: "write default config to file",
		RunE:  runGenDefaultCmd,
	}
	cmd.Flags().String("dst", "celo.yaml", "destination file")
	return cmd
}

func runGenDefaultCmd(cmd *cobra.Command, args []string) error {
	dst, err := cmd.Flags().GetString("dst")
	if err != nil {
		return err
	}
	err = config.SaveConfig(dst)
	if err != nil {
		p.Error(fmt.Sprintf("write default config to %s failed : %v", dst, err))
		return err
	} else {
		p.Success(fmt.Sprintf("write default config to %s success", dst))
	}
	return nil
}
