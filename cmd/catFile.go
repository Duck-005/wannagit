package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/Duck-005/wannagit/utils"
)

func ObjectFind(repo utils.Repo, name string, objectType string, follow bool) string {
	if objectType == "" {
		objectType = "blob"
	}
	return name
}
	
var catFileCmd = &cobra.Command{
	Use:   "catFile TYPE OBJECT_HASH",
	Short: "prints the raw uncompressed object data to stdout",
	Long: `prints the raw uncompressed object data to stdout without the wannagit header`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := utils.RepoFind(".", true)
		obj := utils.ObjectRead(repo, ObjectFind(repo, args[1], args[0], true))

		if obj == nil {
			return
		}

		fmt.Printf(obj.Serialize())
	},
}

func init() {
	rootCmd.AddCommand(catFileCmd)
}
