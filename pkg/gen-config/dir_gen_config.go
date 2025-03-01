package gen_config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type GitRemote struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type GitRepo struct {
	Directory string      `yaml:"directory"`
	Remotes   []GitRemote `yaml:"remotes"`
}

type Config struct {
	Repositories []GitRepo `yaml:"repositories"`
}

func main() {
	// Example parameters (replace with actual values)
	targetDir := "./your/target/directory"
	configFile := "./config.yaml"
	depth := 2 // Example depth

	repos, err := findGitRepos(targetDir, depth)
	if err != nil {
		fmt.Printf("Error finding git repos: %v\n", err)
		return
	}

	if err := saveToConfig(repos, configFile); err != nil {
		fmt.Printf("Error saving to config: %v\n", err)
	}
}

func findGitRepos(targetDir string, maxDepth int) ([]GitRepo, error) {
	var repos []GitRepo
	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the depth of the current path
		relPath, err := filepath.Rel(targetDir, path)
		if err != nil {
			return err
		}
		depth := len(filepath.SplitList(relPath))

		if depth > maxDepth {
			return filepath.SkipDir
		}

		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			remotes, err := getGitRemotes(repoPath)
			if err != nil {
				return fmt.Errorf("failed to get remotes for %s: %w", repoPath, err)
			}
			repos = append(repos, GitRepo{Directory: repoPath, Remotes: remotes})
			return filepath.SkipDir // Skip further walking inside .git directory
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return repos, nil
}

func getGitRemotes(repoPath string) ([]GitRemote, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseRemotes(string(output)), nil
}

func parseRemotes(output string) []GitRemote {
	lines := splitLines(output)
	remoteMap := make(map[string]string)

	for _, line := range lines {
		fields := filepath.SplitList(line)
		if len(fields) > 2 {
			name := fields[0]
			url := fields[1]
			remoteMap[name] = url
		}
	}

	var remotes []GitRemote
	for name, url := range remoteMap {
		remotes = append(remotes, GitRemote{Name: name, URL: url})
	}
	return remotes
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range filepath.SplitList(s) {
		lines = append(lines, line)
	}
	return lines
}

func saveToConfig(repos []GitRepo, configFile string) error {
	config := Config{Repositories: repos}
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := ioutil.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func createTestDirectories(baseDir string) error {
	repoPaths := []string{
		filepath.Join(baseDir, "repo1/.git"),
		filepath.Join(baseDir, "nested/repo2/.git"),
		filepath.Join(baseDir, "nested/deep/repo3/.git"),
	}

	for _, path := range repoPaths {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create test directory %s: %w", path, err)
		}
	}

	return nil
}
