package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/Duck-005/wannagit/utils"
)

func createRepo(repo utils.Repo) {
	
	if stat, err := os.Stat(repo.Worktree); err == nil {
		if !stat.IsDir() {
			fmt.Print("Not a directory\n")
		} else if stat.Size() != 0 {
			fmt.Printf("Directory is not empty: %v\n", err)
		}
	} else {
		os.MkdirAll(repo.Worktree, os.ModePerm)
	}

	errorHandler := func(err error) {
		if err != nil {
			fmt.Printf("could'nt create repository: %v\n", err)
		}
	}

	_, err := utils.RepoDir(repo, true, "branches")
	errorHandler(err)

	_, err = utils.RepoDir(repo, true, "objects")
	errorHandler(err)

	_, err = utils.RepoDir(repo, true, "refs", "tags")
	errorHandler(err)

	_, err = utils.RepoDir(repo, true, "refs", "heads")
	errorHandler(err)
	
	repoFile, _ := utils.RepoFile(repo, false, "description")
	f, err := os.OpenFile(repoFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	errorHandler(err)
	io.WriteString(f, "Unnamed repository; edit this file 'description' to name the repository.\n")
	f.Close()

	repoFile, _ = utils.RepoFile(repo, false, "HEAD")
	f, err = os.OpenFile(repoFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	errorHandler(err)
	io.WriteString(f, "ref: refs/heads/main\n")
	f.Close()

	configData := defaultConfig()
	
	repoFile, _ = utils.RepoFile(repo, false, "config")
	f, err = os.OpenFile(repoFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	io.WriteString(f, configData)
	errorHandler(err)
	
	f.Close()
}

func defaultConfig() string{
	config := map[string] map[string]string{
		"core": {
			"repositoryformatversion": "0",
			"filemode": "false",
			"bare": "false",
		},
	}

	configData, _ := json.Marshal(config)
	return string(configData)
}

var initCmd = &cobra.Command{
	Use:   "init <path>",
	Short: "Initialize a new wannagit repo",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		repo := utils.Repo{}

		if len(args) == 0 {
			repo.Worktree, _ = filepath.EvalSymlinks(".")
		} else {
			repo.Worktree, _ = filepath.EvalSymlinks(args[0])
		}

		repo.Gitdir = filepath.Join(repo.Worktree, ".wannagit")
		repo.Conf = filepath.Join(repo.Worktree, repo.Gitdir, "config")

		createRepo(repo)

		fmt.Printf("initializing the repository at %v", repo.Worktree)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
