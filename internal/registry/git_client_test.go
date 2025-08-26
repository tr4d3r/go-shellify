package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitClient_GetRepositoryPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-client-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	client := NewGitClient(tmpDir)
	
	tests := []struct {
		name     string
		repoName string
		expected string
	}{
		{
			name:     "simple repository name",
			repoName: "test-repo",
			expected: filepath.Join(tmpDir, "test-repo"),
		},
		{
			name:     "repository with hyphens",
			repoName: "my-test-repo",
			expected: filepath.Join(tmpDir, "my-test-repo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.GetRepositoryPath(tt.repoName)
			if result != tt.expected {
				t.Errorf("GetRepositoryPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGitClient_IsRepositoryCloned(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-client-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	client := NewGitClient(tmpDir)

	// Test non-existent repository
	if client.IsRepositoryCloned("non-existent") {
		t.Error("IsRepositoryCloned() should return false for non-existent repository")
	}

	// Create a mock git repository directory structure
	repoName := "test-repo"
	repoPath := client.GetRepositoryPath(repoName)
	gitDir := filepath.Join(repoPath, ".git")
	
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create mock git directory: %v", err)
	}

	// Test existing repository
	if !client.IsRepositoryCloned(repoName) {
		t.Error("IsRepositoryCloned() should return true for existing repository")
	}
}

func TestGitClient_RemoveRepository(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-client-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	client := NewGitClient(tmpDir)

	// Test removing non-existent repository
	err = client.RemoveRepository("non-existent")
	if err == nil {
		t.Error("RemoveRepository() should return error for non-existent repository")
	}

	// Create a mock repository
	repoName := "test-repo"
	repoPath := client.GetRepositoryPath(repoName)
	gitDir := filepath.Join(repoPath, ".git")
	
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create mock git directory: %v", err)
	}

	// Test successful removal
	if err := client.RemoveRepository(repoName); err != nil {
		t.Errorf("RemoveRepository() failed: %v", err)
	}

	// Verify repository is gone
	if client.IsRepositoryCloned(repoName) {
		t.Error("Repository should be removed after RemoveRepository()")
	}
}

func TestParseUnixTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
	}{
		{
			name:    "valid timestamp",
			input:   "1672531200", // 2023-01-01 00:00:00 UTC
			wantErr: false,
		},
		{
			name:    "zero timestamp",
			input:   "0",
			wantErr: false,
		},
		{
			name:    "invalid timestamp",
			input:   "abc123",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseUnixTimestamp(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Error("parseUnixTimestamp() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseUnixTimestamp() unexpected error: %v", err)
				return
			}

			// For valid cases, just ensure we get a valid time
			if result.IsZero() && tt.input != "0" {
				t.Errorf("parseUnixTimestamp() returned zero time for valid input: %s", tt.input)
			}
		})
	}
}