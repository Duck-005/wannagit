package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Duck-005/wannagit/utils"
	"github.com/spf13/cobra"
)

// TODO: FIX THIS SHIT

func gitignoreParse1(data string) *utils.Rule {
	data = strings.TrimSpace(data)

	if data == "" || strings.Compare(string(data[0]), "#") == 0 {
		return nil

	} else if string(data[0]) == "!"{
		return &utils.Rule{Path: data[1:], IsIgnored: false}

	} else if string(data[0]) == "\\" {
		return &utils.Rule{Path: data[1:], IsIgnored: true}

	} else {
		return &utils.Rule{Path: data, IsIgnored: true}	
	}
}

func gitignoreParse(lines []string) []utils.Rule {
	var ret []utils.Rule
	for _, line := range lines {
		rule  := gitignoreParse1(line)
		if rule != nil {
			ret = append(ret, *rule)
		}
	}
	return ret
}

func gitignoreRead(repo utils.Repo) (*utils.GitIgnore, error){
	ret := &utils.GitIgnore{
		Absolute: [][]utils.Rule{}, 
		Scoped: make(map[string][]utils.Rule),
	}
	
	readLines := func (filePath string) ([]string, error) {
		if _, err := os.Stat(filePath); err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, fmt.Errorf("error accessing file %v: %w", filePath, err)
		}

		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("couldn't open %v: %w ", filePath, err)
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
	
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("couldn't read %v: %w", filePath, err)
		}

		return lines, nil
	}

	repoFile := filepath.Join(repo.Gitdir, "info", "exclude")
	lines, err := readLines(repoFile)
	if err != nil {
		return nil, err
	} else if lines != nil {
		ret.Absolute = append(ret.Absolute, gitignoreParse(lines))
	}
	
	var configHome string
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		configHome = xdg
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home user directory: %w", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}

	globalFile := filepath.Join(configHome, "git", "ignore")
	if  lines, err := readLines(globalFile); err != nil {
		return nil, err
	} else if lines != nil {
		parsed := gitignoreParse(lines)
		ret.Absolute = append(ret.Absolute, parsed)
	}
	
	index, err := utils.IndexRead(repo)
	if err != nil {
		return nil, fmt.Errorf("no index file found: %w", err)
	}

	for _, entry := range index.Entries {
		if entry.Name == ".gitignore" || strings.HasSuffix(entry.Name, "/.gitignore") {
			dirName := filepath.Dir(entry.Name)

			contents := utils.ObjectRead(repo, entry.SHA)
			text := contents.Serialize()

			lines := strings.FieldsFunc(text, func(r rune) bool {
				return r == '\n' || r == '\r'
			})

			parsed := gitignoreParse(lines)
			if len(parsed) > 0 {
				ret.Scoped[dirName] = parsed
			}
		}
	}
	return ret, nil
}

func checkIgnore1(rules []utils.Rule, relPath string) bool {
	var last bool
    matched := false

	matchGitignorePattern := func (pattern, relPath string) bool {
		// Handle directory patterns like "build/"
		if strings.HasSuffix(pattern, "/") {
			return relPath == strings.TrimSuffix(pattern, "/") ||
				strings.HasPrefix(relPath, pattern)
		}
	
		// Pattern with slash → match against whole relative path
		if strings.Contains(pattern, string(os.PathSeparator)) {
			ok, _ := filepath.Match(pattern, relPath)
			return ok
		}
	
		// Pattern without slash → match against basename only
		ok, _ := filepath.Match(pattern, filepath.Base(relPath))
		return ok
	}
	
    for _, rule := range rules {
        if matchGitignorePattern(rule.Path, relPath) {
            last = rule.IsIgnored
            matched = true
        }
    }

    if matched {
        return last
    }
    return false
}

func checkIgnoreScoped(rules map[string][]utils.Rule, path string) bool {
	parent := filepath.Dir(path)
	
	for {
		if rs, ok := rules[parent]; ok {
			if res := checkIgnore1(rs, path); res {
				return true
			}
		}
		if parent == "."{
			break
		}
		parent = filepath.Dir(parent)
	}
	return false
}

func checkIgnoreAbsolute(rules [][]utils.Rule, path string) bool {
	result := false
	for _, ruleset := range rules {
		if res := checkIgnore1(ruleset, path); res {
			result = true
		}
	}
	return result
}

func checkIgnore(rules *utils.GitIgnore, path string) (bool, error) {
	if filepath.IsAbs(path) {
		return false, fmt.Errorf("this function requires path to be relative to the repository's root")
	}

	if checkIgnoreScoped(rules.Scoped, path) {
		return true, nil
	}

	return checkIgnoreAbsolute(rules.Absolute, path), nil
}

var checkIgnoreCmd = &cobra.Command{
	Use:   "checkIgnore",
	Short: "check path(s) against ignore rules",
	Long: `check path(s) against ignore rules`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := utils.RepoFind(".", true)

		rules, err := gitignoreRead(repo)
		if err != nil {
			fmt.Print(err)
			return
		}

		for _, path := range args {
			isIgnored, err := checkIgnore(rules, path)
			if err != nil {
				fmt.Print(err)
			}

			if isIgnored {
				fmt.Println(path)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(checkIgnoreCmd)
}
