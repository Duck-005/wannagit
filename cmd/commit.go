package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

func expandUserHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	expandedPath := filepath.Join(homeDir, path[2:])
	return expandedPath
}

func gitconfigRead() []string {
	var configFiles []string

	if runtime.GOOS == "windows" {
		configFiles = append(configFiles, `C:\ProgramData\Git\config`)
	} else {
		configFiles = append(configFiles, "/etc/gitconfig")
	}

	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		configFiles = append(configFiles, expandUserHome("~/.gitconfig"))
	} else {
		configFiles = append(configFiles, expandUserHome(filepath.Join(xdgConfigHome, "git/config")))
	}

	repo := utils.RepoFind(".", true)
	if repo.Conf == "" {
		configFiles = append(configFiles, filepath.Join(repo.Worktree, ".git", "config"))
	}

	return configFiles
}

func gitconfigUserGet() string {
	files := gitconfigRead()

	var configPath string
	for _, f := range files {
		if _, err := os.Stat(f); err == nil {
			configPath = f
		}
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("error in reading config file: %v", configPath)
		return "" 
	}

	var config map[string] map[string]string
		
	isValid := json.Valid(raw)
	if isValid {
		json.Unmarshal(raw, &config)
	} else {
		fmt.Printf("Invalid config file: %v", configPath)
	}

	if user, ok :=config["user"]; ok {
		if name, ok := user["name"]; ok {
			if email, ok := user["email"]; ok {
				return fmt.Sprintf("%v <%v>", name, email)
			}
		}
	}

	return ""
}

type treeEntry struct {
	gitIndexEntry utils.GitIndexEntry
	sha string
	basename string
}

func treeFromIndex(repo utils.Repo, index utils.GitIndex) string {
	contents := make(map[string] []treeEntry)

	for _, entry := range index.Entries {
		dirname := filepath.Dir(entry.Name)
		
		key := dirname
		for key != "" && key != "." && key != string(filepath.Separator){
			if _, ok := contents[key]; !ok {
				contents[key] = []treeEntry{}
			}

			parent := filepath.Dir(key)
			if parent == key { // stop if Dir() no longer changes
				break
			}
			key = parent
		}

		contents[dirname] = append(contents[dirname], treeEntry{gitIndexEntry: entry})
	}

	sortedPaths := make([]string, 0, len(contents))
	for k := range contents {
    	sortedPaths = append(sortedPaths, k)
	}

	sort.Slice(sortedPaths, func(i, j int) bool {
		return len(sortedPaths[i]) > len(sortedPaths[j]) // > gives reverse order
	})

	var sha string
	for _, path := range sortedPaths {
		tree := utils.GitTree{}

		for _, entry := range contents[path] {
			var leaf utils.GitTreeLeaf
			if entry.gitIndexEntry.Name != "" {
				leafMode := fmt.Sprintf("%02o%04o", entry.gitIndexEntry.ModeType, entry.gitIndexEntry.ModePerms)
				leaf = utils.GitTreeLeaf{
					Mode: leafMode,
					Path: filepath.Base(entry.gitIndexEntry.Name),
					Sha: entry.gitIndexEntry.SHA,
				}
			} else {
				leaf = utils.GitTreeLeaf{
					Mode: "040000",
					Path: entry.basename,
					Sha: entry.sha,
				}
			}

			tree.Items = append(tree.Items, leaf)
		}
		sha = utils.ObjectWrite(&tree, repo)

		if path != "" {
			parent := filepath.Dir(path)
			base := filepath.Base(path)

			contents[parent] = append(contents[parent], treeEntry{basename: base, sha: sha})
		}
	}

	return sha
}

func commitCreate(repo utils.Repo, tree string, parent string, author string, timestamp time.Time, message string ) string {
	commit := utils.GitCommit{
		Data: make(map[string][]string),
	}
	commit.Data["tree"] = []string{tree}
	if parent != "" {
		commit.Data["parent"] = []string{parent}
	}

	message = strings.TrimSpace(message) + "\n"

	ts := time.Now().UTC() 
	loc, _ := time.LoadLocation("Asia/Kolkata")
	localTs := ts.In(loc)
	_, offset := localTs.Zone()

	hours := int(math.Floor(float64(offset / 3600)))
	minutes := int(math.Floor(float64((offset % 3600) / 60)))
	
	sign := "+"
	if offset < 0 {
		sign = "-"
		hours = -hours // make hours positive for formatting
		minutes = -minutes
	}

	// format as ±HHMM
	tz := fmt.Sprintf("%s%02d%02d", sign, hours, minutes)
	epoch := ts.Unix()
	author = fmt.Sprintf("%s %d %s", author, epoch, tz)

	commit.Data["author"] = []string{author}
	commit.Data["committer"] = []string{author}
	commit.Data[""] = []string{message}

	return utils.ObjectWrite(&commit, repo)
}

var commitCmd = &cobra.Command{
	Use:   "commit -m MESSAGE",
	Short: "record changes to the repository",
	Long: `create a new commit containing the current contents of the index and the given log message describing the changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := utils.RepoFind(".", true)

		index, err := utils.IndexRead(repo)
		utils.ErrorHandler("error in reading index", err)
		
		tree := treeFromIndex(repo, *index)

		message, _ := cmd.Flags().GetString("message")
		commit := commitCreate(
			repo, 
			tree,
			utils.ObjectFind(repo, "HEAD", "commit", true),
			// gitconfigUserGet(),
			"aaa",
			time.Now(),
			message,
		)
		fmt.Printf("created commit: %v", commit)

		head, _ := os.ReadFile(filepath.Join(repo.Gitdir, "HEAD"))
		
		var isActiveBranch bool
		if strings.HasPrefix(string(head), "ref: refs/heads/") {
			isActiveBranch = true
		} else {
			isActiveBranch = false
		}

		if isActiveBranch {
			branchPath, err := utils.RepoFile(repo, false, filepath.Join("refs", "heads", string(head[16:len(head)-1])))
			utils.ErrorHandler("error in reading refs/heads/BRANCH", err)

			fd, err := os.OpenFile(branchPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				utils.ErrorHandler("error updating HEAD", err)
			}
			defer fd.Close()

			fd.Write([]byte(commit + "\n"))
			
		} else {
			headPath, err := utils.RepoFile(repo, false, "HEAD")
			utils.ErrorHandler("error in reading HEAD", err)

			fd, err := os.OpenFile(headPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				utils.ErrorHandler("error updating HEAD", err)
			}
			defer fd.Close()

			fd.Write([]byte("\n")) 
		}
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)

	commitCmd.Flags().StringP("message", "m", "", "message to associate a commit with")
}
