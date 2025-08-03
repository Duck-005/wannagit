package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/Duck-005/wannagit/utils"
)

func checkoutTree(repo utils.Repo, tree *utils.GitTree, path string) {
	for _, item := range tree.Items {
		obj := utils.ObjectRead(repo, item.Sha)
		dest := filepath.Join(path, item.Path)

		if obj.Format() == "tree" {
			err := os.MkdirAll(dest, 0755)
			utils.ErrorHandler("error creating directory", err)

			tree := obj.(*utils.GitTree)
			checkoutTree(repo, tree, dest)

		} else if obj.Format() == "blob" {
			err := os.WriteFile(dest, []byte(obj.Serialize()), os.ModePerm)
			utils.ErrorHandler("error writing to file", err)
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

		repo := utils.RepoFind(".", true)
		obj := utils.ObjectRead(repo, args[0])

		if obj.Format() == "commit" {
			commit := obj.(*utils.GitCommit)
			obj = utils.ObjectRead(repo, commit.GetData()["tree"][0])
		}

		tree, ok := obj.(*utils.GitTree)
		if !ok {
			fmt.Print("Not a tree object\n")
			return
		}

		target := args[1]
		stat, err := os.Stat(target)

		if err == nil && stat.IsDir() {
			var dir *os.File 
			dir, err = os.Open(target)
			utils.ErrorHandler("couldn't open directory", err)

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
				utils.ErrorHandler("couldn't create directory", err)
				return
			}
		} else {
			utils.ErrorHandler("Not a directory", err)
			return
		}

		path, err := filepath.EvalSymlinks(target)
		utils.ErrorHandler("couldn't evaluate symlinks", err)
		
		checkoutTree(repo, tree, path)
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
