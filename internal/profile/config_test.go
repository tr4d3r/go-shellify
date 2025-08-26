package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.Version != ConfigVersion {
		t.Errorf("Expected version %s, got %s", ConfigVersion, config.Version)
	}
	
	if !config.Shell.AutoDetect {
		t.Error("Expected auto_detect to be true by default")
	}
	
	if config.Shell.Type != "" {
		t.Errorf("Expected empty shell type by default, got %s", config.Shell.Type)
	}
	
	if config.Generation.IntegrationMode != "source" {
		t.Errorf("Expected integration_mode to be 'source', got %s", config.Generation.IntegrationMode)
	}
	
	if !config.Generation.BackupExisting {
		t.Error("Expected backup_existing to be true by default")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *ProfileConfig
		expectErr bool
	}{
		{
			name: "valid config",
			config: &ProfileConfig{
				Version: "1.0.0",
				Generation: struct {
					Verbose         bool   `json:"verbose"`
					BackupExisting  bool   `json:"backup_existing"`
					IntegrationMode string `json:"integration_mode"`
				}{
					IntegrationMode: "source",
				},
				Output: struct {
					Directory string `json:"directory"`
					Filename  string `json:"filename"`
				}{
					Directory: "/tmp/test",
					Filename:  "test",
				},
			},
			expectErr: false,
		},
		{
			name: "invalid integration mode",
			config: &ProfileConfig{
				Generation: struct {
					Verbose         bool   `json:"verbose"`
					BackupExisting  bool   `json:"backup_existing"`
					IntegrationMode string `json:"integration_mode"`
				}{
					IntegrationMode: "invalid",
				},
			},
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error, got %v", err)
			}
		})
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "go-shellify-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	configPath := filepath.Join(tempDir, "config.json")
	
	// Create and save a config
	config := DefaultConfig()
	config.Shell.Type = "zsh"
	config.Modules.Enabled = []string{"git", "node"}
	
	err = config.SaveToPath(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Load the config back
	loadedConfig, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify the loaded config matches
	if loadedConfig.Shell.Type != "zsh" {
		t.Errorf("Expected shell type 'zsh', got '%s'", loadedConfig.Shell.Type)
	}
	
	if len(loadedConfig.Modules.Enabled) != 2 {
		t.Errorf("Expected 2 enabled modules, got %d", len(loadedConfig.Modules.Enabled))
	}
	
	if loadedConfig.Modules.Enabled[0] != "git" || loadedConfig.Modules.Enabled[1] != "node" {
		t.Errorf("Expected modules [git, node], got %v", loadedConfig.Modules.Enabled)
	}
}

func TestLoadNonexistentConfig(t *testing.T) {
	_, err := LoadFromPath("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Expected error loading nonexistent config, got nil")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tempDir, err := os.MkdirTemp("", "go-shellify-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	configPath := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configPath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}
	
	_, err = LoadFromPath(configPath)
	if err == nil {
		t.Error("Expected error loading invalid JSON, got nil")
	}
}

func TestModuleManagement(t *testing.T) {
	config := DefaultConfig()
	
	// Test adding modules
	config.AddModule("git")
	config.AddModule("node")
	config.AddModule("git") // Should not duplicate
	
	if len(config.Modules.Enabled) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(config.Modules.Enabled))
	}
	
	if !config.IsModuleEnabled("git") {
		t.Error("Expected git module to be enabled")
	}
	
	if !config.IsModuleEnabled("node") {
		t.Error("Expected node module to be enabled")
	}
	
	// Test removing module
	config.RemoveModule("git")
	if config.IsModuleEnabled("git") {
		t.Error("Expected git module to be disabled after removal")
	}
	
	if len(config.Modules.Enabled) != 1 {
		t.Errorf("Expected 1 module after removal, got %d", len(config.Modules.Enabled))
	}
	
	// Test wildcard
	config.Modules.Enabled = []string{"*"}
	if !config.IsModuleEnabled("any-module") {
		t.Error("Expected any module to be enabled with wildcard")
	}
}

func TestRegistryManagement(t *testing.T) {
	config := DefaultConfig()
	
	// Test adding registries
	config.AddRegistry("origin")
	config.AddRegistry("company")
	config.AddRegistry("origin") // Should not duplicate
	
	if len(config.Modules.Registries) != 2 {
		t.Errorf("Expected 2 registries, got %d", len(config.Modules.Registries))
	}
	
	// Test removing registry
	config.RemoveRegistry("origin")
	if len(config.Modules.Registries) != 1 {
		t.Errorf("Expected 1 registry after removal, got %d", len(config.Modules.Registries))
	}
	
	if config.Modules.Registries[0] != "company" {
		t.Errorf("Expected remaining registry to be 'company', got '%s'", config.Modules.Registries[0])
	}
}

func TestConfigPaths(t *testing.T) {
	configPath, err := GetConfigPath()
	if err != nil {
		t.Errorf("GetConfigPath() returned error: %v", err)
	}
	
	if configPath == "" {
		t.Error("GetConfigPath() returned empty path")
	}
	
	if !filepath.IsAbs(configPath) {
		t.Error("GetConfigPath() should return absolute path")
	}
	
	configDir, err := GetConfigDir()
	if err != nil {
		t.Errorf("GetConfigDir() returned error: %v", err)
	}
	
	if configDir == "" {
		t.Error("GetConfigDir() returned empty path")
	}
	
	if !filepath.IsAbs(configDir) {
		t.Error("GetConfigDir() should return absolute path")
	}
}

func TestJSONMarshaling(t *testing.T) {
	config := DefaultConfig()
	config.Shell.Type = "zsh"
	config.Modules.Enabled = []string{"git", "node"}
	
	// Test marshaling to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	
	// Test unmarshaling from JSON
	var unmarshaledConfig ProfileConfig
	err = json.Unmarshal(data, &unmarshaledConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}
	
	// Verify the unmarshaled config matches
	if unmarshaledConfig.Shell.Type != "zsh" {
		t.Errorf("Expected shell type 'zsh', got '%s'", unmarshaledConfig.Shell.Type)
	}
	
	if len(unmarshaledConfig.Modules.Enabled) != 2 {
		t.Errorf("Expected 2 enabled modules, got %d", len(unmarshaledConfig.Modules.Enabled))
	}
}