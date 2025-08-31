package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Repo struct {
	Worktree 	string
	Gitdir 		string
	Conf 		string
}

type GitObject interface {
	Serialize() string
	Deserialize(data string)
	Format() string
}

type BaseGitObject struct {
	data 		string
	format 		string
}

func (b *BaseGitObject) Format() string {
	return b.format
}

// GitBlob ----------------------------------------

type GitBlob struct {
	BaseGitObject
}

func (b *GitBlob) Serialize() string {
	b.format = "blob"
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

func (b *GitTag) GetData() map[string][]string {
	return b.Data
}

// GitIndexEntry and GitIndex ----------------------------------

type GitIndex struct {
	Version 	uint32
	Entries 	[]GitIndexEntry
}

type GitIndexEntry struct {
	Ctime            [2]uint32 // metadata modified timestamp in (seconds, nanoseconds)
	Mtime            [2]uint32 // file modified timestamp in (seconds, nanoseconds)
	Dev              uint32 // device ID containing this file
	Ino              uint32 // the file's inode number
	ModeType         uint16 // the object type, either b1000 (regular), b1010 (symlink), b1110 (gitlink)
	ModePerms        uint16 // file permissions
	UID              uint32 // user ID of the owner
	GID              uint32 // group ID of the owner
	Size             uint32 // size of this object in bytes
	SHA              string // object's SHA
	AssumeValid      bool
	Stage            uint16
	Name             string // full path of this object (name)
}

// GitIgnore and Rule --------------------------------------

type Rule struct {
	Path string
	IsIgnored bool
}

type GitIgnore struct {
	Absolute [][]Rule
	Scoped map[string][]Rule
}

// helper functions -------------------------------

func RepoFind(path string, required bool) Repo {
	path, _ = filepath.Abs(path)

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

// makes the directory if it doesn't exist
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

func ResolveRef(repo Repo, ref string) string {
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
		return ResolveRef(repo, data[5:])
	} 

	return data
}

func ErrorHandler(customMsg string, err error) {
	if err != nil {
		fmt.Printf(customMsg + "\nerror: ", err)
	}
}
