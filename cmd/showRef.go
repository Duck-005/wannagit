package cmd

import (
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"github.com/spf13/cobra"
)

func resolveRef(repo Repo, ref string) string {
	path, err := RepoFile(repo, false, ref)
	ErrorHandler("", err)

	stat, err := os.Stat(path)
	if err != nil || !stat.Mode().IsRegular() {
		return ""
	}

	dataSlice, err := os.ReadFile(path)
	ErrorHandler("couldn't read ref file", err)

	data := strings.TrimSpace(string(dataSlice))

	if strings.HasPrefix(data, "ref: ") {
		return resolveRef(repo, data[5:])
	} 

	return data
}

func listRef(repo Repo, path string) map[string]any {
	var err error
	var basePath string

	if path == "" {
		basePath, err = RepoDir(repo, false, "refs")
		ErrorHandler("", err)
		path = "refs"
	} else {
		basePath, err = RepoDir(repo, false, path)
		ErrorHandler("", err)
	}

	refMap := make(map[string]any)

	entries, err := os.ReadDir(basePath)
	ErrorHandler("", err)

	for _, entry := range entries {
		relativePath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			refMap[entry.Name()] = listRef(repo, relativePath)
		} else {
			refMap[entry.Name()] = resolveRef(repo, relativePath)
		}
	}

	return refMap
}

func showRef(repo Repo, refs map[string]any, withHash bool, prefix string) {
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
		repo := RepoFind(".", true)
		refs := listRef(repo, "")

		showRef(repo, refs, true, "refs")
	},
}

func init() {
	rootCmd.AddCommand(showRefCmd)
}
