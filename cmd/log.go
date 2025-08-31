package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

func logGraphviz(repo utils.Repo, sha string, seenSet map[string]bool, log string) (string, error) {
	var err error
	if seenSet[sha] {
		return log, err
	}
	seenSet[sha] = true

	object := utils.ObjectRead(repo, sha)
	commit, ok := object.(*utils.GitCommit)
	if !ok {
		return "", fmt.Errorf("error reading commit object: %v", sha)
	}
	message := string(commit.Data[""][0])
	message = strings.ReplaceAll(message, "\\", "\\\\")
	message = strings.ReplaceAll(message, "\"", "\\\"")

	if strings.Contains(message, "\n") {
		message = message[:strings.Index(message, "\n")]
	}

	log += 	fmt.Sprintf(" c_%v [label=\"%v: %v\"]", sha, sha[0:7], message)

	if _, ok := commit.Data["parent"]; !ok {
		// base case: the initial commit
		return log, err
	}

	for _, parent := range commit.Data["parent"] {
		log += fmt.Sprintf(" c_%v -> c_%v;", sha, parent)
		l, err := logGraphviz(repo, parent, seenSet, log)
		if err != nil {
			return "", err
		}
		log = l
	}
	return log, err
}

var logCmd = &cobra.Command{
	Use:   "log COMMIT_HASH",
	Short: "review logging of commit data and its metadata",
	Long: `review the different commits along with their information like the authors, time stamps etc.
	use dot -O -Tpdf log.dot to generate a pdf of the commit tree`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Print("Usage: log COMMIT_HASH")
			return 
		}

		repo := utils.RepoFind(".", true)

		log := ""
		log += "digraph wannagitLog{node[shape=rect]"
		l, err := logGraphviz(repo, utils.ObjectFind(repo, args[0], "", false), make(map[string]bool), log)
		if err != nil {
			utils.ErrorHandler("error in getting log data", err)
			return
		}
		l += "}"

		os.WriteFile("log.dot", []byte(l), os.ModePerm)
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
