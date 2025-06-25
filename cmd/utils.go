package cmd

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	data map[string][]string
}

func (b *GitCommit) Serialize() string {
	return string(SerializeKVLM(b.data))
}

func (b *GitCommit) Deserialize(data string) {
	b.data = ParseKVLM([]byte(data))
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

func ErrorHandler(customMsg string, err error) {
	if err != nil {
		fmt.Printf(customMsg + "\nerror: ", err)
	}
}

func ObjectRead(repo Repo, sha string) GitObject {
	path, _ := RepoFile(repo, false, "objects", sha[:2], sha[2:])

	if stat, _ := os.Stat(path); !stat.Mode().IsRegular() {
		fmt.Printf("Not a valid object file: %v", sha)
	} 

	file, err := os.Open(path)
	if err != nil {
		ErrorHandler("couldn't open object file", err)
		return nil
	}
	defer file.Close()

	zlibReader, err := zlib.NewReader(file)
	if err != nil {
		ErrorHandler("could'nt create zlib reader", err)
		return nil
	}
	defer zlibReader.Close() 

	var decompressed bytes.Buffer
	_, err = io.Copy(&decompressed, zlibReader)
	if err != nil {
		ErrorHandler("failed to decompress object", err)
		return nil
	}

	rawSlice := decompressed.Bytes()

	spaceIdx := bytes.IndexByte(rawSlice, ' ')
	nullIdx := bytes.IndexByte(rawSlice, 0)

	if spaceIdx == -1 || nullIdx == -1 {
		fmt.Printf("Malformed object %v: missing header format\n", sha)
		return nil
	}

	format := string(rawSlice[0:spaceIdx])
	size, _ := strconv.Atoi(string(rawSlice[spaceIdx+1:nullIdx]))

	if size != len(rawSlice) - nullIdx - 1 {
		fmt.Printf("malformed object %v: bad length", sha)
		return nil
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

	result := []byte(obj.Format() + " " + strconv.Itoa(len(data)) + "\x00" + string(data))

	hash := sha1.Sum(result)
	sha := fmt.Sprintf("%x", hash[:])

	if repo.gitdir == "" {
		return sha
	}

	path, _ := RepoFile(repo, true, "objects", sha[0:2], sha[2:]) 
	if _, err := os.Stat(path); err != nil {
		var buf bytes.Buffer
		writer := zlib.NewWriter(&buf)
		writer.Write(result)
		writer.Close()

		err = os.WriteFile(path, buf.Bytes(), os.ModePerm)
		ErrorHandler("could'nt write object to file", err)
	}

	return sha
}

type KVLM map[string][]string

func ParseKVLM(raw []byte) KVLM {
	return parseKVLM(raw, 0, make(KVLM))
}

func parseKVLM(raw []byte, start int, dict KVLM) KVLM {
	spaceIdx := bytes.IndexByte(raw[start:], ' ')
	newLineIdx := bytes.IndexByte(raw[start:], '\n')

	if spaceIdx == -1 || newLineIdx < spaceIdx {
		// Base case: no more key-value pairs, only commit message
		if newLineIdx != 0 {
			panic("expected newline at start of commit message")
		}
		dict[""] = []string{string(raw[start+1:])}
		return dict
	}

	spaceIdx += start
	newLineIdx += start
	key := string(raw[start:spaceIdx])

	// Find end of value (handling continuation lines)
	end := spaceIdx
	for {
		nextNewLine := bytes.IndexByte(raw[end+1:], '\n')
		if nextNewLine == -1 {
			panic("unterminated header value")
		}
		nextNewLine += end + 1
		if nextNewLine+1 >= len(raw) || raw[nextNewLine+1] != ' ' {
			end = nextNewLine
			break
		}
		end = nextNewLine
	}

	valBytes := raw[spaceIdx+1 : end]
	valBytes = bytes.ReplaceAll(valBytes, []byte("\n "), []byte("\n"))
	value := string(valBytes)

	if existing, ok := dict[key]; ok {
		dict[key] = append(existing, value)
	} else {
		dict[key] = []string{value}
	}

	return parseKVLM(raw, end+1, dict)
}

func SerializeKVLM(dict KVLM) []byte {
	var buf bytes.Buffer

	for key, values := range dict {
		if key == "" {
			continue
		}

		for _, v := range values {
			escaped := strings.ReplaceAll(v, "\n", "\n")
			buf.WriteString(fmt.Sprintf("%s %s\n", key, escaped))
		}
	}

	buf.WriteByte('\n')
	if msg, ok := dict[""]; ok {
		buf.WriteString(strings.Join(msg, "\n"))
	}

	return buf.Bytes()
}
