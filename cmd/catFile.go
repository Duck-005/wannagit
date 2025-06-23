package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func ObjectFind(repo Repo, name string, objectType string, follow bool) string {
	return name
}
	
var catFileCmd = &cobra.Command{
	Use:   "catFile TYPE OBJECT_HASH",
	Short: "prints the raw uncompressed object data to stdout",
	Long: `prints the raw uncompressed object data to stdout without the wannagit header`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := RepoFind(".", true)
		obj := ObjectRead(repo, ObjectFind(repo, args[1], args[0], true))

		fmt.Printf(obj.Serialize())
	},
}

func init() {
	rootCmd.AddCommand(catFileCmd)
}
