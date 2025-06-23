package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func objectHash(repo Repo, file *os.File, format string) string {
	data, _ := io.ReadAll(file)
	
	var obj GitObject
	switch format {
		case "commit": obj = &GitCommit{}
		case "tree": obj = &GitTree{}
		case "tag": obj = &GitTag{}
		case "blob": obj = &GitBlob{}

		default: fmt.Printf("Unknown type format %v", format)
	}

	obj.Deserialize(string(data))

	return ObjectWrite(obj, repo)
}

var hashObjectCmd = &cobra.Command{
	Use:   "hashObject [-w] [-t TYPE] FILE",
	Short: "create hash-object",
	Long: `create the hash-object for a particular file and 
	write it to the git directory optionally`,
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("type")
		write, _ := cmd.Flags().GetBool("write")

		var repo Repo
		if write {
			repo = RepoFind(".", true)
		} else {
			repo = Repo{}
		}

		file, err := os.Open(args[0])
		ErrorHandler(fmt.Sprintf("Invalid path %v", args[0]), err)

		sha := objectHash(repo, file, format)
		fmt.Print(sha)
	},
}

func init() {
	rootCmd.AddCommand(hashObjectCmd)
	hashObjectCmd.PersistentFlags().String("t", "", "gives the type for the object specified")

	hashObjectCmd.Flags().StringP("type", "t", "blob", "gives the type of object")
	hashObjectCmd.Flags().BoolP("write", "w", false, "writes the object to wannagit directory")
}
