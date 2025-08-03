package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
)

// GitTree ------------------------------------

type GitTree struct {
	BaseGitObject
	Items []GitTreeLeaf
}

func (b *GitTree) Serialize() string {
	return string(treeSerialize(b)) 
	// returns bytes in the form of string NOT READABLE
}

func (b *GitTree) Deserialize(data string) {
	b.Items = ParseTree([]byte(data))
	b.format = "tree"
}

// tree leaf node --------------------------------

type GitTreeLeaf struct {
	Mode string
	Path string
	Sha string
}

func NewGitTreeLeaf(mode string, path string, sha string) *GitTreeLeaf {
	return &GitTreeLeaf {
		Mode: mode,
		Path: path,
		Sha: sha,
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

	rawSha := raw[nullIdx+1 : nullIdx+21]
	sha := hex.EncodeToString(rawSha[:])

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
	sort.Slice(obj.Items, func(i, j int) bool {
		return obj.Items[i].Path < obj.Items[j].Path
	})

	var ret []byte
	for _, node := range obj.Items {
		entry := fmt.Sprintf("%s %s", node.Mode, node.Path)
		ret = append(ret, []byte(entry)...)
		ret = append(ret, 0x00)

		rawSha, err := hex.DecodeString(node.Sha)
		if err != nil {
			panic("invalid SHA in tree leaf: " + node.Sha)
		}
		ret = append(ret, rawSha...)
	}
	return ret
}
