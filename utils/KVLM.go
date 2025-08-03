package utils

import (
	"fmt"
	"strings"
	"bytes"
)

// GitCommit -----------------------------------

type GitCommit struct {
	BaseGitObject
	Data map[string][]string
}

func (b *GitCommit) Serialize() string {
	return string(SerializeKVLM(b.Data))
}

func (b *GitCommit) Deserialize(data string) {
	b.Data = ParseKVLM([]byte(data))
	b.format = "commit"
} 

func (b *GitCommit) GetData() map[string][]string {
	return b.Data
}

// helper functions -------------------------------

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