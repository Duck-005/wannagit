package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"slices"
)

func IndexRead(repo Repo) (*GitIndex, error) {
	indexFile, _ := RepoFile(repo, false, "index")

	data, err := os.ReadFile(indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &GitIndex{}, nil 
		}
		return nil, err
	}

	header := data[:12]
	signature := string(header[:4])
	if signature != "DIRC" {
		return nil, fmt.Errorf("invalid index signature: %s", signature)
	}

	version := binary.BigEndian.Uint32(header[4:8])
	if version != 2 {
		return nil, fmt.Errorf("unsupported index version: %d", version)
	}

	count := int(binary.BigEndian.Uint32(header[8:12]))
	content := data[12:]
	idx := 0
	entries := []GitIndexEntry{}

	for i := 0; i < count; i++ {
		ctime_s := binary.BigEndian.Uint32(content[idx : idx+4])
		ctime_ns := binary.BigEndian.Uint32(content[idx+4 : idx+8])
		mtime_s := binary.BigEndian.Uint32(content[idx+8 : idx+12])
		mtime_ns := binary.BigEndian.Uint32(content[idx+12 : idx+16])
		dev := binary.BigEndian.Uint32(content[idx+16 : idx+20])
		ino := binary.BigEndian.Uint32(content[idx+20 : idx+24])

		unused := binary.BigEndian.Uint16(content[idx+24 : idx+26])
		if unused != 0 {
			return nil, fmt.Errorf("unused field not 0")
		}

		mode := binary.BigEndian.Uint16(content[idx+26 : idx+28])
		modeType := mode >> 12
		valid := slices.Contains([]int{0b1000, 0b1010, 0b1110}, int(modeType))
		if !valid {
			return &GitIndex{}, fmt.Errorf("invalid index file format: invalid modeType: %04b", modeType)
		}
		modePerms := mode & 0x01FF

		uid := binary.BigEndian.Uint32(content[idx+28 : idx+32])
		gid := binary.BigEndian.Uint32(content[idx+32 : idx+36])
		fsize := binary.BigEndian.Uint32(content[idx+36 : idx+40])

		sha := fmt.Sprintf("%040x", content[idx+40:idx+60])

		flags := binary.BigEndian.Uint16(content[idx+60 : idx+62])

		flagAssumeValid := (flags & 0b1000000000000000) != 0
		flagExtended := (flags & 0b0100000000000000) != 0
		if flagExtended {
			return nil, fmt.Errorf("extended flags not supported")
		}
		flagStage := (flags & 0b0011000000000000) >> 12
		nameLength := int(flags & 0x0FFF)

		idx += 62

		var rawName []byte
		if nameLength < 0xFFF {
			if idx+nameLength >= len(content) || content[idx+nameLength] != 0x00 {
				return nil, fmt.Errorf("missing null terminator for name")
			}
			rawName = content[idx : idx+nameLength]
			idx += nameLength + 1
		} else {
			nullIdx := bytes.IndexByte(content[idx+0xFFF:], 0x00)
			if nullIdx == -1 {
				return nil, fmt.Errorf("unterminated name")
			}
			rawName = content[idx : idx+0xFFF+nullIdx]
			fmt.Printf("Notice: Name is 0xFFF or more bytes long.\n")
			idx += 0xFFF + nullIdx + 1
		}

		name := string(rawName)

		idx = 8 * int(math.Ceil(float64(idx)/8))

		entry := GitIndexEntry{
			Ctime:       [2]uint32{ctime_s, ctime_ns},
			Mtime:       [2]uint32{mtime_s, mtime_ns},
			Dev:         dev,
			Ino:         ino,
			ModeType:    modeType,
			ModePerms:   modePerms,
			UID:         uid,
			GID:         gid,
			Size:        fsize,
			SHA:         sha,
			AssumeValid: flagAssumeValid,
			Stage:       flagStage,
			Name:        name,
		}

		entries = append(entries, entry)
	}

	return &GitIndex{
		Version: version,
		Entries: entries,
	}, nil
}