package cmd

import (
	"errors"
	"github.com/heguangyu1989/celo/internal/merge"
	"github.com/spf13/cobra"
)

func GetMergeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge",
		Short: "create a merge request",
		RunE:  runMergeCommand,
	}
	cmd.Flags().String("src", "", "source branch")
	cmd.Flags().String("dst", "", "branch to merge to")
	cmd.Flags().String("title", "", "merge title")
	cmd.Flags().StringArray("tags", []string{}, "tags to merge")
	return cmd
}

func runMergeCommand(cmd *cobra.Command, args []string) error {
	srcBranch, _ := cmd.Flags().GetString("src")
	dstBranch, _ := cmd.Flags().GetString("dst")
	title, _ := cmd.Flags().GetString("title")
	tags, _ := cmd.Flags().GetStringArray("tags")
	if srcBranch == "" || dstBranch == "" || title == "" {
		return errors.New("src, dst title and tags must be set")
	}

	return merge.Merge(srcBranch, dstBranch, title, tags)
}
