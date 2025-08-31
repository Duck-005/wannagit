package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

func getCTime(path string) (int, int, error) {
    info, err := os.Stat(path)
    if err != nil {
        return 0, 0, err
    }
    sec, nsec := utils.ExtractCTime(info)
    return sec, nsec, nil
}

func add(repo utils.Repo, paths []string, del bool, skipMissing bool) {
	rm(repo, paths, true, false)

	worktree := repo.Worktree + string(os.PathSeparator)

	type pair struct {
		abspath string
		relPath string
	}
	var cleanPaths []pair
	for _, path := range paths {
		abspath, _ := filepath.Abs(path)
		if !strings.HasPrefix(abspath, worktree) {
			if stat, err := os.Stat(abspath); err != nil && stat.Mode().IsRegular() {
				fmt.Printf("not a file, or outside the worktree: %v", paths)
				return 
			}
		}
		relPath, _ := filepath.Rel(repo.Worktree, abspath)
		cleanPaths = append(cleanPaths, pair{abspath: abspath, relPath: relPath})
	}

	index, err := utils.IndexRead(repo)
	utils.ErrorHandler("error reading index", err)

	for _, path := range cleanPaths {
		fd, _ := os.Open(path.abspath)
		sha := objectHash(repo, fd, "blob")

		stat, err := os.Stat(path.abspath)
		if err != nil {
			fmt.Printf("error reading file: %v\n", stat.Name())
		}

		ctimeS, ctimeNs, _ := getCTime(path.abspath)
		
		mtimeS := uint32(stat.ModTime().Unix())
		mtimeNs := uint32(stat.ModTime().Nanosecond())

		devIno, _ := utils.GetDevIno(path.abspath)
		GidUid := utils.GetGidUid(path.abspath)

		entry := utils.GitIndexEntry {
			Ctime: [2]uint32{uint32(ctimeS), uint32(ctimeNs)},
			Mtime: [2]uint32{mtimeS, mtimeNs},
			Dev: uint32(devIno.Dev),
			Ino: uint32(devIno.Ino),
			ModeType: 0b1000,
			ModePerms: 0o644,
			UID: GidUid.Uid,
			GID: GidUid.Gid,
			Size: uint32(stat.Size()),
			SHA: sha,
			AssumeValid: false,
			Stage: 0,
			Name: path.relPath,
		}

		index.Entries = append(index.Entries, entry)
	}

	utils.IndexWrite(repo, *index)
}

var addCmd = &cobra.Command{
	Use:   "add <FILE_PATHS>",
	Short: "add file contents to the index",
	Long: `this command updates the index using the current content found in the working tree, to prepare the content
       staged for the next commit.`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := utils.RepoFind(".", true)

		add(repo, args, true, false)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
