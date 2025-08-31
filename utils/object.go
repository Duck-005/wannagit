package utils

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func ObjectRead(repo Repo, sha string) GitObject {
	path, err := RepoFile(repo, false, "objects", sha[:2], sha[2:])
	if err != nil {
		ErrorHandler("couldn't fetch object file", err)
		return nil
	}

	if stat, _ := os.Stat(path); !stat.Mode().IsRegular() {
		fmt.Printf("Not a valid object file: %v", sha)
		return nil
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
		fmt.Printf("Malformed object %v: missing header format", sha)
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

		default: 
			fmt.Printf("Unknown type format %v for object %v", format, sha)
			return nil
	}
	
	obj.Deserialize(string(rawSlice[nullIdx+1:]))
	return obj
}

func ObjectWrite(obj GitObject, repo Repo) string {
	data := obj.Serialize()

	result := []byte(obj.Format() + " " + strconv.Itoa(len(data)) + "\x00" + string(data))
	hash := sha1.Sum(result)
	sha := fmt.Sprintf("%x", hash[:])

	if (repo == Repo{}) {
		return sha
	}

	if repo.Gitdir == "" {
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

func objectResolve(repo Repo, name string) []string{
	// resolves HEAD refs, short, long hashes, tags, branches, remote branches.
	hashRE := "^[0-9A-Fa-f]{4,40}$"
	var candidates []string 

	if strings.TrimSpace(name) == "" {
		return []string{""}
	}

	if name == "HEAD" {
		return []string{ResolveRef(repo, "HEAD")}
	}

	if matched, _ := regexp.MatchString(hashRE, name); matched {
		name = strings.ToLower(name)
		prefix := name[0:2]

		path, err := RepoDir(repo, false, "objects", prefix)
		ErrorHandler("couldn't resolve tag object file", err)

		if path != "" {
			rem := name[2:]

			entries, err := os.ReadDir(path)
			ErrorHandler("couldn't read the ref directory", err)

			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), rem) {
					candidates = append(candidates, prefix + entry.Name())
				}
			}
		}

		asTag := ResolveRef(repo, "refs/tags/" + name)
		if asTag != "" {
			candidates = append(candidates, asTag)
		}

		asBranch := ResolveRef(repo, "refs/heads/" + name)
		if asBranch != "" {
			candidates = append(candidates, asBranch)
		}

		asRemoteBranch := ResolveRef(repo, "refs/remotes/" + name)
		if asRemoteBranch != "" {
			candidates = append(candidates, asRemoteBranch)
		}

		return candidates
	}

	return []string{""}
}

func ObjectFind(repo Repo, name string, objectType string, follow bool) string {
	shas := objectResolve(repo, name)

	if shas[0] == "" {
		panic("No such reference: " + shas[0])
	}

	if len(shas) > 1 {
		fmt.Printf("Ambiguous reference %v: candidates are:\n", name)
		for _, sha := range shas {
			fmt.Printf("%s", sha + "\n")
		}
		panic("")
	}
	
	sha := shas[0]

	if objectType == "" {
		return sha
	}

	for {
		obj := ObjectRead(repo, sha)

		if obj.Format() == objectType {
			return sha
		}

		if !follow {
			return ""
		}

		if obj.Format() == "tag" {
			tag := obj.(*GitTag)
			sha = tag.GetData()["object"][0]
		} else if obj.Format() == "commit" && objectType == "tree"{
			commit := obj.(*GitCommit)
			sha = commit.GetData()["tree"][0]
		} else {
			return ""
		}
	}
}