package utils

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"strconv"

	// "github.com/Duck-005/wannagit/cmd"
)



func ObjectRead(repo Repo, sha string) GitObject {
	path, err := RepoFile(repo, false, "objects", sha[:2], sha[2:])
	ErrorHandler("couldn't fetch object file", err)

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