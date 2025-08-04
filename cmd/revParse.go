package cmd

import (
	"fmt"

	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

var revParseCmd = &cobra.Command{
	Use:   "revParse [--type] [TYPE] REFERENCE",
	Short: "Parse revision (or other objects) identifiers",
	Long: `Parse revision (or other objects) identifiers`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Print("Usage: revParse --type TYPE REFERENCE")
			return 
		}

		format, _ := cmd.Flags().GetString("type")
		
		repo := utils.RepoFind(".", true)
		sha := utils.ObjectFind(repo, args[0], format, true)
		
		if sha == "" {
			fmt.Print("None")
		} else {
			fmt.Print(sha)
		}
	},
}

func init() {
	rootCmd.AddCommand(revParseCmd)

	revParseCmd.Flags().StringP("type", "t", "blob", "give the type of object")
}
