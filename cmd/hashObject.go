package cmd

import (
	"fmt"
	"io"
	"os"
	
	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

func objectHash(repo utils.Repo, file *os.File, format string) string {
	data, err := io.ReadAll(file)
	utils.ErrorHandler("couldn't open file for creating object: ", err)
	
	var obj utils.GitObject
	switch format {
		case "commit": obj = &utils.GitCommit{}
		case "tree": obj = &utils.GitTree{}
		case "tag": obj = &utils.GitTag{}
		case "blob": obj = &utils.GitBlob{}

		default: fmt.Printf("Unknown type format %v", format)
	}

	obj.Deserialize(string(data))

	return utils.ObjectWrite(obj, repo)
}

var hashObjectCmd = &cobra.Command{
	Use:   "hashObject [-w] [-t TYPE] FILE",
	Short: "create hash-object",
	Long: `create the hash-object for a particular file and 
	write it to the git directory optionally`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Print("Usage: hashObject [-w] [-t TYPE] FILE")
			return
		}

		format, _ := cmd.Flags().GetString("type")
		write, _ := cmd.Flags().GetBool("write")

		var repo utils.Repo
		if write {
			repo = utils.RepoFind(".", true)
		} else {
			repo = utils.Repo{}
		}

		file, err := os.Open(args[0])
		utils.ErrorHandler(fmt.Sprintf("Invalid path %v", args[0]), err)

		sha := objectHash(repo, file, format)
		fmt.Print(sha)
	},
}

func init() {
	rootCmd.AddCommand(hashObjectCmd)

	hashObjectCmd.Flags().StringP("type", "t", "blob", "gives the type of object")
	hashObjectCmd.Flags().BoolP("write", "w", false, "writes the object to wannagit directory")
}
