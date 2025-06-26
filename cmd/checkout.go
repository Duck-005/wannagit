package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/spf13/cobra"
)

func checkoutTree(repo Repo, tree *GitTree, path string) {
	for _, item := range tree.items {
		obj := ObjectRead(repo, item.sha)
		dest := filepath.Join(path, item.path)

		if obj.Format() == "tree" {
			err := os.MkdirAll(dest, 0755)
			ErrorHandler("error creating directory", err)

			tree := obj.(*GitTree)
			checkoutTree(repo, tree, dest)

		} else if obj.Format() == "blob" {
			err := os.WriteFile(dest, []byte(obj.Serialize()), os.ModePerm)
			ErrorHandler("error writing to file", err)
		}
	} 
}

var checkoutCmd = &cobra.Command{
	Use:   "checkout COMMIT DIRECTORY",
	Short: "checkout a commit inside of an empty directory",
	Long: `ensure the directory is empty before running the command`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Print("usage: checkout COMMIT DIRECTORY\n")
			return
		}

		repo := RepoFind(".", true)
		obj := ObjectRead(repo, args[0])

		if obj.Format() == "commit" {
			commit := obj.(*GitCommit)
			obj = ObjectRead(repo, commit.GetData()["tree"][0])
		}

		tree, ok := obj.(*GitTree)
		if !ok {
			fmt.Print("Not a tree object\n")
			return
		}

		target := args[1]
		stat, err := os.Stat(target)

		if err == nil && stat.IsDir() {
			var dir *os.File 
			dir, err = os.Open(target)
			ErrorHandler("couldn't open directory", err)

			_, err = dir.Readdirnames(1)
			if err == nil {
				fmt.Print("Directory is NOT EMPTY\n")
				return
			} else if err.Error() != "EOF" {
				fmt.Printf("error reading directory: %v\n", err)
				return
			}
		} else if os.IsNotExist(err) {
			err := os.MkdirAll(target, 0755)
			if err != nil {
				ErrorHandler("couldn't create directory", err)
				return
			}
		} else {
			ErrorHandler("Not a directory", err)
			return
		}

		path, err := filepath.EvalSymlinks(target)
		ErrorHandler("couldn't evaluate symlinks", err)
		
		checkoutTree(repo, tree, path)
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
