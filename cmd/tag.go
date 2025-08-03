package cmd

import (
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag [NAME] [OBJECT] [-a]",
	Short: "add a reference in refs/tags/",
	Long: `add the `,
	Run: func(cmd *cobra.Command, args []string) {
		
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
}
