package cmd

import (
	"github.com/heguangyu1989/celo/pkg/config"
	"github.com/heguangyu1989/celo/pkg/utils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "celo",
	Short: "Efficiency, at speed.",
}

func init() {
	rootCmd.PersistentFlags().String("config", config.DefaultPath(), "config file path")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		if utils.FileExists(configFile) {
			return config.LoadConfig(configFile)
		}
		return nil
	}

	rootCmd.AddCommand(GetBuildInfoCmd())
	rootCmd.AddCommand(GetMD5Cmd())
	rootCmd.AddCommand(GetGenDefaultCmd())
	rootCmd.AddCommand(GetMergeCommand())
	rootCmd.AddCommand(GetEnvListCmd())
}

func Execute() error {
	return rootCmd.Execute()
}
