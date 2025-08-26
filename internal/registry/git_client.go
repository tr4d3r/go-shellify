package registry

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/griffin/go-shellify/internal/logger"
)

// GitClient handles git repository operations
type GitClient struct {
	cacheDir string
}

// NewGitClient creates a new git client
func NewGitClient(cacheDir string) *GitClient {
	return &GitClient{
		cacheDir: cacheDir,
	}
}

// CloneRepository clones a git repository to the cache directory
func (g *GitClient) CloneRepository(url, name string) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(g.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	targetDir := filepath.Join(g.cacheDir, name)

	// Check if repository already exists
	if _, err := os.Stat(targetDir); err == nil {
		logger.Debug("Repository already exists, updating: %s", targetDir)
		return g.updateRepository(targetDir)
	}

	logger.Info("Cloning repository: %s to %s", url, targetDir)

	// Perform shallow clone for performance
	cmd := exec.Command("git", "clone", "--depth", "1", url, targetDir)
	cmd.Env = os.Environ()
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	logger.Debug("Repository cloned successfully: %s", targetDir)
	return nil
}

// updateRepository updates an existing repository
func (g *GitClient) updateRepository(repoDir string) error {
	logger.Debug("Updating repository: %s", repoDir)

	// Change to repository directory and pull latest changes
	cmd := exec.Command("git", "pull", "--depth", "1")
	cmd.Dir = repoDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w, output: %s", err, string(output))
	}

	logger.Debug("Repository updated successfully: %s", repoDir)
	return nil
}

// GetRepositoryPath returns the local path for a repository
func (g *GitClient) GetRepositoryPath(name string) string {
	return filepath.Join(g.cacheDir, name)
}

// IsRepositoryCloned checks if a repository is already cloned
func (g *GitClient) IsRepositoryCloned(name string) bool {
	repoPath := g.GetRepositoryPath(name)
	gitDir := filepath.Join(repoPath, ".git")
	
	if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
		return true
	}
	
	return false
}

// RemoveRepository removes a cloned repository from cache
func (g *GitClient) RemoveRepository(name string) error {
	repoPath := g.GetRepositoryPath(name)
	
	if !g.IsRepositoryCloned(name) {
		return fmt.Errorf("repository not found: %s", name)
	}

	logger.Info("Removing repository: %s", repoPath)
	
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}

	return nil
}

// GetRepositoryInfo returns basic information about a cloned repository
func (g *GitClient) GetRepositoryInfo(name string) (*RepositoryInfo, error) {
	repoPath := g.GetRepositoryPath(name)
	
	if !g.IsRepositoryCloned(name) {
		return nil, fmt.Errorf("repository not cloned: %s", name)
	}

	info := &RepositoryInfo{
		Name: name,
		Path: repoPath,
	}

	// Get remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoPath
	
	if output, err := cmd.Output(); err == nil {
		info.RemoteURL = strings.TrimSpace(string(output))
	}

	// Get last commit info
	cmd = exec.Command("git", "log", "-1", "--format=%H|%s|%ct")
	cmd.Dir = repoPath
	
	if output, err := cmd.Output(); err == nil {
		parts := strings.Split(strings.TrimSpace(string(output)), "|")
		if len(parts) >= 3 {
			info.LastCommitHash = parts[0]
			info.LastCommitMessage = parts[1]
			
			if timestamp, err := parseUnixTimestamp(parts[2]); err == nil {
				info.LastCommitTime = timestamp
			}
		}
	}

	return info, nil
}

// RepositoryInfo contains information about a cloned repository
type RepositoryInfo struct {
	Name              string
	Path              string
	RemoteURL         string
	LastCommitHash    string
	LastCommitMessage string
	LastCommitTime    time.Time
}

// parseUnixTimestamp parses a unix timestamp string to time.Time
func parseUnixTimestamp(timestampStr string) (time.Time, error) {
	timestamp, err := parseIntFromString(timestampStr)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(timestamp, 0), nil
}

// parseIntFromString parses an integer from string
func parseIntFromString(s string) (int64, error) {
	var result int64
	var err error
	
	// Simple integer parsing without external dependencies
	for i, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid character at position %d: %c", i, r)
		}
		result = result*10 + int64(r-'0')
	}
	
	return result, err
}