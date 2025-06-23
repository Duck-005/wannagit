package cmd

import (
	"bytes"
	"compress/zlib"
	"io"
	"strconv"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
)

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

// GitCommit -----------------------------------

type GitCommit struct {
	BaseGitObject
}

func (b *GitCommit) Serialize() string {
	return b.data
}

func (b *GitCommit) Deserialize(data string) {
	b.data = data
	b.format = "commit"
} 

// GitTree ------------------------------------

type GitTree struct {
	BaseGitObject
}

func (b *GitTree) Serialize() string {
	return b.data
}

func (b *GitTree) Deserialize(data string) {
	b.data = data
	b.format = "tree"
}

// GitTag ----------------------------------------

type GitTag struct {
	BaseGitObject
}

func (b *GitTag) Serialize() string {
	return b.data
}

func (b *GitTag) Deserialize(data string) {
	b.data = data
	b.format = "tag"
}

// GitBLob ----------------------------------------

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

// -------------------------------------------------

func RepoFind(path string, required bool) Repo {
	path, _ = filepath.EvalSymlinks(path)

	if stat, _ := os.Stat(filepath.Join(path, ".wannagit")); stat.IsDir() {
		return Repo {
			worktree: path,
			gitdir: filepath.Join(path, ".wannagit"),
			conf: filepath.Join(path, ".wannagit", "config"),
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

func errorHandler(customMsg string, err error) {
	if err != nil {
		fmt.Printf(customMsg + ": ", err)
	}
}

func ObjectRead(repo Repo, sha string) GitObject {
	path, _ := RepoFile(repo, false, "objects", sha[:2], sha[2:])

	if stat, _ := os.Stat(path); !stat.Mode().IsRegular() {
		fmt.Printf("Not a valid object file: %v", sha)
	} 

	file, err := os.Open(path)
	errorHandler("could'nt open object file", err)
	defer file.Close()

	zlibReader, err := zlib.NewReader(file)
	errorHandler("could'nt create zlib reader", err)
	defer zlibReader.Close() 

	var decompressed bytes.Buffer
	io.Copy(&decompressed, zlibReader)

	rawSlice := decompressed.Bytes()

	spaceIdx := bytes.IndexByte(rawSlice, ' ')
	nullIdx := bytes.IndexByte(rawSlice, 0)

	format := string(rawSlice[0:spaceIdx])
	size, _ := strconv.Atoi(string(rawSlice[spaceIdx+1:nullIdx]))

	if size != len(rawSlice) - nullIdx - 1 {
		fmt.Printf("malformed object %v: bad length", sha)
	}
	
	var obj GitObject

	switch format {
		case "commit": obj = &GitCommit{}
		case "tree": obj = &GitTree{}
		case "tag": obj = &GitTag{}
		case "blob": obj = &GitBlob{}

		default: fmt.Printf("Unknown type format %v for object %v", format, sha)
	}
	obj.Deserialize(string(rawSlice[nullIdx+1:]))
	return obj
}

func ObjectWrite(obj GitObject, repo Repo) string {
	data := obj.Serialize()

	result := []byte(obj.Format() + " " + strconv.Itoa(len(data)) + "\x00" + data)

	hash := sha1.Sum(result)
	sha := fmt.Sprintf("%x", hash[:])

	if repo.gitdir == "" {
		return sha
	}

	path, _ := RepoFile(repo, true, sha[0:2], sha[2:]) 
	if _, err := os.Stat(path); err != nil {
		err = os.WriteFile(path, result, os.ModePerm)
		errorHandler("could'nt write object to file", err)
	}

	return sha
}
