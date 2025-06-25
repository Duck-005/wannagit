package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

func logGraphviz(repo Repo, sha string, seenSet map[string]bool) {
	if seenSet[sha] {
		return
	}
	seenSet[sha] = true

	object := ObjectRead(repo, sha)
	commit, ok := object.(*GitCommit)
	if !ok {
		return
	}
	message := string(commit.data[""][0])
	message = strings.ReplaceAll(message, "\\", "\\\\")
	message = strings.ReplaceAll(message, "\"", "\\\"")

	if strings.Contains(message, "\n") {
		message = message[:strings.Index(message, "\n")]
	}

	fmt.Printf(" c_%v [label=\"%v: %v\"]", sha, sha[0:7], message)
	
	if _, ok := commit.data["parent"]; !ok {
		// base case: the initial commit
		return
	}

	for _, parent := range commit.data["parent"] {
		fmt.Printf(" c_%v -> c_%v;", sha, parent)
		logGraphviz(repo, parent, seenSet)
	}
}

var logCmd = &cobra.Command{
	Use:   "log COMMIT_HASH",
	Short: "review logging of commit data and its metadata",
	Long: `review the different commits along with their information like the authors, time stamps etc.
	use dot -O -Tpdf log.dot to generate a pdf of the commit tree`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := RepoFind(".", true)

		fmt.Print("digraph wannagitLog{\n")
		fmt.Print("	node[shape=rect]")

		logGraphviz(repo, ObjectFind(repo, args[0], "", false), make(map[string]bool))

		fmt.Print("}")
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
