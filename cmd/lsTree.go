package cmd

import (
	"fmt"
	"path"

	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

func lsTree(repo utils.Repo, ref string, recursive bool, prefix string)  {
	sha := utils.ObjectFind(repo, ref, "tree", true)
	obj := utils.ObjectRead(repo, sha)

	tree, ok := obj.(*utils.GitTree)
	if !ok {
		return
	}

	var typ string
	var typBits string

	for _, item := range tree.Items {
		if len(item.Mode) == 5 {
			typBits = item.Mode[0:1]
		} else {
			typBits = item.Mode[0:2]
		}

		switch typBits {
			case "04": typ = "tree"
			case "10": typ = "blob"
			case "12": typ = "blob"
			case "16": typ = "commit"
			default: fmt.Printf("weird tree leaf node, mode: %v path: %v\n", item.Mode, item.Path)
		}

		if recursive && typ == "tree" {
			lsTree(repo, item.Sha, recursive, path.Join(prefix, item.Path))
		} else {
			fmt.Printf("%06s %v %v\t%v\n", item.Mode, typ, item.Sha, path.Join(prefix, item.Path))
		}
	}
}

var lsTreeCmd = &cobra.Command{
	Use:   "lsTree [-r] TREE_HASH",
	Short: "prints the content of the tree object in a list",
	Long: `use -r switch to recursively print all the object files, i.e no tree objects
	only blobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Print("Usage: lsTree [-r] TREE_HASH")
			return 
		}

		recursive, _ := cmd.Flags().GetBool("recursive")

		repo := utils.RepoFind(".", true)
		lsTree(repo, args[0], recursive, "")
	},
}

func init() {
	rootCmd.AddCommand(lsTreeCmd)

	lsTreeCmd.Flags().BoolP("recursive", "r", false, "recursively prints the blobs instead of the tree objects")
}
