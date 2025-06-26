package cmd

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// GitTree ------------------------------------

type GitTree struct {
	BaseGitObject
	items []GitTreeLeaf
}

func (b *GitTree) Serialize() string {
	return treeSerialize(b)
}

func (b *GitTree) Deserialize(data string) {
	b.items = ParseTree([]byte(data))
	b.format = "tree"
}

// tree leaf node --------------------------------

type GitTreeLeaf struct {
	mode string
	path string
	sha string
}

func NewGitTreeLeaf(mode string, path string, sha string) *GitTreeLeaf {
	return &GitTreeLeaf {
		mode: mode,
		path: path,
		sha: sha,
	}
}

// helper functions ----------------------------

func treeParseLeaf(raw []byte, start int) (position int, node GitTreeLeaf){
	spaceIdx := bytes.IndexByte(raw, ' ')

	if spaceIdx-start != 5 || spaceIdx-start != 6 {
		fmt.Printf("invalid tree node\n")
	}

	mode := string(raw[start:spaceIdx])
	if len(mode) == 5 {
		mode = "0" + mode
	}

	nullIdx := bytes.IndexByte(raw, '\x00')
	path := string(raw[spaceIdx+1:nullIdx])

	sha := hex.EncodeToString(raw[nullIdx+1 : nullIdx+21])

	return nullIdx+21, *NewGitTreeLeaf(mode, path, sha)
}

func ParseTree(raw []byte) []GitTreeLeaf {
	pos := 0
	max := len(raw)

	var ret []GitTreeLeaf
	var data GitTreeLeaf

	for {
		if pos < max {
			pos, data = treeParseLeaf(raw, pos)
			ret = append(ret, data)
		} else {
			break
		}
	}

	return ret
}

func convertPath(mode string) string {
	if strings.HasPrefix(mode, "10") {
		return mode
	} else {
		return mode + "\\"
	}
}

func treeSerialize(obj *GitTree) string {
	sort.Slice(obj.items, func(i, j int) bool {
		mode_i := convertPath(obj.items[i].mode)
		mode_j := convertPath(obj.items[j].mode)
		return mode_i < mode_j
	})

	var ret string
	for _, node := range obj.items {
		ret += node.mode
		ret += " "
		ret += node.path
		ret += "\x00"
		sha := fmt.Sprintf("%x", node.sha)
		ret += sha
	}
	return ret
}
