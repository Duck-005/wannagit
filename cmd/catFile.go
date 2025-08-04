package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/Duck-005/wannagit/utils"
)
	
var catFileCmd = &cobra.Command{
	Use:   "catFile TYPE OBJECT_HASH",
	Short: "prints the raw uncompressed object data to stdout",
	Long: `prints the raw uncompressed object data to stdout without the wannagit header`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Printf("Usage: catFile TYPE OBJECT_HASH")
			return 
		}

		repo := utils.RepoFind(".", true)
		obj := utils.ObjectRead(repo, utils.ObjectFind(repo, args[1], args[0], true))

		if obj == nil {
			return
		}

		fmt.Printf(obj.Serialize())
	},
}

func init() {
	rootCmd.AddCommand(catFileCmd)
}
