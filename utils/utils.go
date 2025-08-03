package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

type Repo struct {
	Worktree string;
	Gitdir string;
	Conf string;
}

type GitObject interface {
	Serialize() string
	Deserialize(data string)
	Format() string
}

type BaseGitObject struct {
	data string
	format string
}

func (b *BaseGitObject) Format() string {
	return b.format
}

// GitBlob ----------------------------------------

type GitBlob struct {
	BaseGitObject
}

func (b *GitBlob) Serialize() string {
	return b.data
}

func (b *GitBlob) Deserialize(data string) {
	b.data = data
	b.format = "blob"
}

// GitTag ----------------------------------------

type GitTag struct {
	GitCommit
}

func (b *GitTag) Deserialize(data string) {
	b.Data = ParseKVLM([]byte(data))
	b.format = "tag"
}

//--------------------------------------------------

func RepoFind(path string, required bool) Repo {
	path, _ = filepath.EvalSymlinks(path)

	if stat, _ := os.Stat(filepath.Join(path, ".wannagit")); stat.IsDir() {
		return Repo {
			Worktree: path,
			Gitdir: filepath.Join(path, ".wannagit"),
			Conf: filepath.Join(path, ".wannagit", "config"),
		}
	}

	parent, _ := filepath.EvalSymlinks(filepath.Join(path, ".."))

	if parent == path {
		// if parent == path then parent is root
		// /.. --> / still root
		if required {
			fmt.Print("No git Directory\n")
		} else {
			return Repo {}
		}
	}

	// recursively go back to find the .git folder
	return RepoFind(parent, required)
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

// building a path from the Gitdir of repository
func repoPath(repo Repo, path ...string) string {

	return filepath.Join(repo.Gitdir, filepath.Join(path...))
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

func ErrorHandler(customMsg string, err error) {
	if err != nil {
		fmt.Printf(customMsg + "\nerror: ", err)
	}
}
