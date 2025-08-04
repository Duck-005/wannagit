package cmd

import (
	"os"
	"fmt"
	"path/filepath"

	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

func listRef(repo utils.Repo, path string) map[string]any {
	var err error
	var basePath string

	if path == "" {
		basePath, err = utils.RepoDir(repo, false, "refs")
		utils.ErrorHandler("", err)
		path = "refs"
	} else {
		basePath, err = utils.RepoDir(repo, false, path)
		utils.ErrorHandler("", err)
	}

	refMap := make(map[string]any)

	entries, err := os.ReadDir(basePath)
	utils.ErrorHandler("", err)

	for _, entry := range entries {
		relativePath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			refMap[entry.Name()] = listRef(repo, relativePath)
		} else {
			refMap[entry.Name()] = utils.ResolveRef(repo, relativePath)
		}
	}

	return refMap
}

func showRef(repo utils.Repo, refs map[string]any, withHash bool, prefix string) {
	if prefix != "" {
		prefix += "/"
	}

	for k, v := range refs {
		switch val := v.(type) {
		case string:
			if withHash {
				fmt.Printf("%s %s%s\n", val, prefix, k)
			} else {
				fmt.Printf("%s%s\n", prefix, k)
			}
		case map[string]any:
			showRef(repo, val, withHash, prefix+k)
		default:
			fmt.Printf("Invalid ref at %s%s\n", prefix, k)
		}
	}
}

var showRefCmd = &cobra.Command{
	Use:   "showRef",
	Short: "shows all the references to commit files in the repository",
	Long: `shows all the references to commit files in the repository recursively`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := utils.RepoFind(".", true)
		refs := listRef(repo, "")

		showRef(repo, refs, true, "refs")
	},
}

func init() {
	rootCmd.AddCommand(showRefCmd)
}
