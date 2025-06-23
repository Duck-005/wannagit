package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"encoding/json"
	"github.com/spf13/cobra"
)

type Repo struct {
	worktree string;
	gitdir string;
	conf string;
}

// building a path from the gitdir of repository
func repoPath(repo Repo, path ...string) string {

	return filepath.Join(repo.gitdir, filepath.Join(path...))
}

// creates the path if it does not exist
func RepoFile(repo Repo, mkdir bool, path ...string) (string, error) {

	if _, err := RepoDir(repo, mkdir, path[:len(path) - 1]...); err == nil {
		return repoPath(repo, path...), nil
	} else {
		fmt.Print(err)
	}
	
	return "", fmt.Errorf("couldn't make the repo path")
}

// makes the directory if it does'nt exist
func RepoDir(repo Repo, mkdir bool, path ...string) (string, error) {
	
	dirPath := repoPath(repo, path...)

	if stat, err := os.Stat(dirPath); err == nil {
		if stat.IsDir() {
			return dirPath, nil
		} else {
			return "", fmt.Errorf("not a directory")
		}
	} 

	if mkdir {
		os.MkdirAll(dirPath, os.ModePerm)
		return dirPath, nil
	}
	return "", fmt.Errorf("error occurred")
}

func createRepo(repo Repo) {
	
	if stat, err := os.Stat(repo.worktree); err == nil {
		if !stat.IsDir() {
			fmt.Print("Not a directory\n")
		} else if stat.Size() != 0 {
			fmt.Print("Directory is not empty: %v\n", err)
		}
	} else {
		os.MkdirAll(repo.worktree, os.ModePerm)
	}

	errorHandler := func(err error) {
		if err != nil {
			fmt.Printf("could'nt create repository: %v\n", err)
		}
	}

	_, err := RepoDir(repo, true, "branches")
	errorHandler(err)

	_, err = RepoDir(repo, true, "objects")
	errorHandler(err)

	_, err = RepoDir(repo, true, "refs", "tags")
	errorHandler(err)

	_, err = RepoDir(repo, true, "refs", "heads")
	errorHandler(err)
	
	repoFile, _ := RepoFile(repo, false, "description")
	f, err := os.OpenFile(repoFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	errorHandler(err)
	io.WriteString(f, "Unnamed repository; edit this file 'description' to name the repository.\n")
	f.Close()

	repoFile, _ = RepoFile(repo, false, "HEAD")
	f, err = os.OpenFile(repoFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	errorHandler(err)
	io.WriteString(f, "ref: refs/heads/main\n")
	f.Close()

	configData := defaultConfig()
	
	repoFile, _ = RepoFile(repo, false, "config")
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

		repo := Repo{}

		if len(args) == 0 {
			repo.worktree, _ = filepath.EvalSymlinks(".")
		} else {
			repo.worktree, _ = filepath.EvalSymlinks(args[0])
		}

		repo.gitdir = filepath.Join(repo.worktree, ".wannagit")
		repo.conf = filepath.Join(repo.worktree, repo.gitdir, "config")

		createRepo(repo)

		fmt.Printf("initializing the repository at %v", repo.worktree)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
