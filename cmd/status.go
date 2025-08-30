package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

func cmdStatusBranch(repo utils.Repo) {
	// get the active branch
	head, _ := os.ReadFile(filepath.Join(repo.Gitdir, "HEAD"))

	var isActiveBranch bool
	if strings.HasPrefix(string(head), "ref: refs/heads/") {
		isActiveBranch = true
	} else {
		isActiveBranch = false
	}

	if isActiveBranch {
		fmt.Printf("On branch %v\n", string(head[16:len(head)-1]))
	} else {
		fmt.Printf("HEAD detached at %v\n", utils.ObjectFind(repo, "HEAD", "commit", true))
	}
}

func treeToMap(repo utils.Repo, ref string, prefix string) (map[string]string, error) {
	result := make(map[string]string)
	treeSha := utils.ObjectFind(repo, ref, "tree", true)
	obj := utils.ObjectRead(repo, treeSha)

	tree, ok := obj.(*utils.GitTree)
	if !ok {
		return nil, fmt.Errorf("reference hash is not a tree object")
	}

	for _, leaf := range tree.Items {
		fullPath := path.Join(prefix, leaf.Path)
		modeInt, _ := strconv.ParseInt(leaf.Mode, 8, 32)
		isSubtree := modeInt == 040000

		if isSubtree {
			subDict, err := treeToMap(repo, leaf.Sha, fullPath)
			if err != nil {
				return nil, err
			}
			for k, v := range subDict {
				result[k] = v
			}
		} else {
			result[fullPath] = leaf.Sha
		}
	}

	return result, nil
}

func cmdStatusHeadIndex(repo utils.Repo, index utils.GitIndex) {
	fmt.Println("changes to be committed:")

	head, err := treeToMap(repo, "HEAD", "")
	utils.ErrorHandler("", err)

	for _, entry := range index.Entries {
		if sha, ok := head[entry.Name]; ok{
			if sha != entry.SHA {
				fmt.Printf("  modified:%v\n", entry.Name)
			}
			delete(head, entry.Name)
		} else {
			fmt.Printf("  added:  %v\n", entry.Name)
		}
	}

	for entry := range head {
		fmt.Printf("  deleted:  %v\n", entry)
	}
}

func cmdStatusIndexWorktree(repo utils.Repo, index utils.GitIndex) ([]string, error){
	fmt.Println("changes not staged for commit:")

	ignore, err := gitignoreRead(repo)
	utils.ErrorHandler("error in reading gitignore file", err)

	gitignorePrefix := repo.Gitdir + string(os.PathSeparator)
	var allFiles []string

	_ = filepath.WalkDir(repo.Worktree, func (path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if path == repo.Gitdir || strings.HasPrefix(path, gitignorePrefix) || strings.HasPrefix(path, ".git") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			relPath, err := filepath.Rel(repo.Worktree, path)
			if err != nil {
				return err
			}
			allFiles = append(allFiles, strings.ReplaceAll(relPath, "\\", "/"))
		}
		return nil
	})

	for _, entry := range index.Entries {
		fullPath := path.Join(repo.Worktree, entry.Name)
		if stat, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("  deleted:  %v\n", entry.Name)
		} else {
			mtimeNs := entry.Mtime[0] * 10^9 + entry.Mtime[1]
			if int64(stat.ModTime().Nanosecond()) != int64(mtimeNs) {
				file, _ := os.Open(fullPath)
				newSha := objectHash(utils.Repo{}, file, "blob")
				defer file.Close()

				if newSha != entry.SHA {
					fmt.Printf("  modified:  %v\n", entry.Name)
				}
			}
		}

		for i, f := range allFiles {
			if f == entry.Name {
				allFiles = append(allFiles[:i], allFiles[i+1:]...)
				break
			}
		}
	}

	fmt.Println()
	fmt.Println("untracked files:")

	for _, f := range allFiles {
		if isIgnored, err := checkIgnore(ignore, f); !isIgnored && err == nil {
			fmt.Println(" ", f)
		}
	}

	return allFiles, nil
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "gives the status of the current ",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		repo := utils.RepoFind(".", true)
		index, err := utils.IndexRead(repo)
		if err != nil {
			fmt.Print(err)
			return
		}
		
		cmdStatusBranch(repo)
		cmdStatusHeadIndex(repo, *index)
		fmt.Println()
		cmdStatusIndexWorktree(repo, *index)
	},
} 

func init() {
	rootCmd.AddCommand(statusCmd)
}
