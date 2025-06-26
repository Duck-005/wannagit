package cmd

import (
	"fmt"
	"path"
	"github.com/spf13/cobra"
)

func lsTree(repo Repo, ref string, recursive bool, prefix string)  {
	sha := ObjectFind(repo, ref, "tree", false)
	obj := ObjectRead(repo, sha)

	tree, ok := obj.(*GitTree)
	if !ok {
		return
	}

	var typ string
	var typBits string

	for _, item := range tree.items {
		if len(item.mode) == 5 {
			typBits = item.mode[0:1]
		} else {
			typBits = item.mode[0:2]
		}

		switch typBits {
			case "4": typ = "tree"
			case "10": typ = "blob"
			case "12": typ = "blob"
			case "16": typ = "commit"
			default: fmt.Printf("weird tree leaf node, mode: %v path: %v\n", item.mode, item.path)
		}

		if recursive && typ == "tree" {
			lsTree(repo, item.sha, recursive, path.Join(prefix, item.path))
		} else {
			fmt.Printf("%06s %v %v\t%v\n", item.mode, typ, sha, path.Join(prefix, item.path))
		}
	}
}

var lsTreeCmd = &cobra.Command{
	Use:   "lsTree [-r] TREE_HASH",
	Short: "prints the content of the tree object in a list",
	Long: `use -r switch to recursively print all the object files, i.e no tree objects
	only blobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		recursive, _ := cmd.Flags().GetBool("recursive")

		repo := RepoFind(".", true)
		lsTree(repo, args[0], recursive, "")
	},
}

func init() {
	rootCmd.AddCommand(lsTreeCmd)

	lsTreeCmd.Flags().BoolP("recursive", "r", false, "recursively prints the blobs instead of the tree objects")
}
