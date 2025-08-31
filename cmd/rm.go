package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

func rm(repo utils.Repo, paths []string, skipMissing bool, del bool) {
	index, err := utils.IndexRead(repo)
	utils.ErrorHandler("error in reading index file", err)

	worktree := repo.Worktree + string(os.PathSeparator)

	abspaths := make(map[string]struct{})
	for _, path := range paths {
		abspath, _ := filepath.Abs(path)

		if strings.HasPrefix(abspath, worktree) {
			abspaths[abspath] = struct{}{}
		} else {
			fmt.Printf("cannot remove paths outside of worktree: %v %v", abspath, worktree)
			return
		}
	}

	var keptEntries []utils.GitIndexEntry // entries to write back to the index
	var remove []string // list of removed paths, which is used to physically remove paths from filesystem

	for _, e := range index.Entries {
		fullPath := filepath.Join(repo.Worktree, e.Name)

		if _, ok := abspaths[fullPath]; ok {
			remove = append(remove, fullPath)
			delete(abspaths, fullPath)
		} else {
			keptEntries = append(keptEntries, e)
		}
	}

	if len(abspaths) > 0 && !skipMissing {
		fmt.Print("cannot remove paths not in the index: ", abspaths)
		return 
	}

	if del {
		for _, path := range remove {
			os.Remove(path)
		}
	}

	index.Entries = keptEntries
	utils.IndexWrite(repo, *index)
}

var rmCmd = &cobra.Command{
	Use:   "rm <FILE_PATHS>",
	Short: "remove files from the working tree and from the index",
	Long: `remove files from the working tree and from the index`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := utils.RepoFind(".", true)
		rm(repo, args, false, true)
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
