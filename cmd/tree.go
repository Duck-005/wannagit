package cmd

import (
	"bytes"
	"fmt"
	"sort"
)

// GitTree ------------------------------------

type GitTree struct {
	BaseGitObject
	items []GitTreeLeaf
}

func (b *GitTree) Serialize() string {
	return string(treeSerialize(b)) 
	// returns bytes in the form of string NOT READABLE
}

func (b *GitTree) Deserialize(data string) {
	b.items = ParseTree([]byte(data))
	b.format = "tree"
}

// tree leaf node --------------------------------

type GitTreeLeaf struct {
	mode string
	path string
	sha []byte
}

func NewGitTreeLeaf(mode string, path string, sha []byte) *GitTreeLeaf {
	return &GitTreeLeaf {
		mode: mode,
		path: path,
		sha: sha,
	}
}

// helper functions ----------------------------

func treeParseLeaf(raw []byte, start int) (position int, node GitTreeLeaf){
	spaceIdx := bytes.IndexByte(raw[start:], ' ')
	if spaceIdx == -1 {
		panic("invalid tree: no space found")
	}
	spaceIdx += start

	if spaceIdx-start != 5 && spaceIdx-start != 6 {
		fmt.Printf("invalid tree node\n")
	}

	mode := string(raw[start:spaceIdx])
	
	nullIdx := bytes.IndexByte(raw[spaceIdx+1:], 0x00)
	if nullIdx == -1 {
		panic("invalid tree: no null byte found")
	}
	nullIdx += spaceIdx + 1

	path := string(raw[spaceIdx+1:nullIdx])

	var sha []byte
	copy(sha[:], raw[nullIdx+1:nullIdx+21])

	return nullIdx+21, *NewGitTreeLeaf(mode, path, sha)
}

func ParseTree(raw []byte) []GitTreeLeaf {
	pos := 0
	max := len(raw)

	var ret []GitTreeLeaf
	var node GitTreeLeaf

	for pos < max {
		pos, node = treeParseLeaf(raw, pos)
		ret = append(ret, node)
	}

	return ret
}

func treeSerialize(obj *GitTree) []byte {
	sort.Slice(obj.items, func(i, j int) bool {
		return obj.items[i].path < obj.items[j].path
	})

	var ret []byte
	for _, node := range obj.items {
		entry := fmt.Sprintf("%s %s", node.mode, node.path)
		ret = append(ret, []byte(entry)...)
		ret = append(ret, 0x00)
		ret = append(ret, node.sha[:]...)
	}
	return ret
}
