package cmd

import (
	"fmt"
	"time"

	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

var lsFilesCmd = &cobra.Command{
	Use:   "lsFiles [-v|--verbose]",
	Short: "lists out all the stage files",
	Long: `lists out all the files in the staging area`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := utils.RepoFind(".", true)

		index, _ := utils.IndexRead(repo)
		if index.Version == 0 {
			return 
		}
		
		isVerbose, _ := cmd.Flags().GetBool("verbose")
		if isVerbose {
			fmt.Printf("Index file format v%v, containing %v entries", index.Version, len(index.Entries))
		}

		for _, entry := range index.Entries {
			fmt.Println(entry.Name)
			if isVerbose {
				entryType := map[uint16]string{
					0b1000: "regular file",
					0b1010: "symlink",
					0b1110: "git link",
				}[entry.ModeType]

				fmt.Printf("	%v with perms: %o", entryType, entry.ModePerms)
				fmt.Printf("	on blob: %v", entry.SHA)
				fmt.Printf("	created: %v, modified: %v", 
					time.Unix(int64(entry.Ctime[0]), int64(entry.Ctime[1])),
					time.Unix(int64(entry.Mtime[0]), int64(entry.Mtime[1])),
				)
				fmt.Printf("	device: %v, inode: %v", entry.Dev, entry.Ino)
				// fmt.Printf("	flags: stage=%v assumeValid=%v", entry.FlagStage, entry.FlagAssumeValid)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(lsFilesCmd)

	lsFilesCmd.Flags().BoolP("verbose", "v", false, "list out all the info about the files in the staging area")
}
