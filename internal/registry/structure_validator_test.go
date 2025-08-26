package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStructureValidator_ValidateSemanticVersion(t *testing.T) {
	validator := &StructureValidator{}

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "valid version 1.0.0",
			version: "1.0.0",
			wantErr: false,
		},
		{
			name:    "valid version with pre-release",
			version: "1.0.0-beta",
			wantErr: false,
		},
		{
			name:    "valid version with build metadata",
			version: "1.0.0+20230101",
			wantErr: false,
		},
		{
			name:    "valid complex version",
			version: "2.1.3-beta.1+build.123",
			wantErr: false,
		},
		{
			name:    "invalid version - missing patch",
			version: "1.0",
			wantErr: true,
		},
		{
			name:    "invalid version - non-numeric",
			version: "1.a.0",
			wantErr: true,
		},
		{
			name:    "invalid version - empty",
			version: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateSemanticVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSemanticVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStructureValidator_ValidateRegistryName(t *testing.T) {
	validator := &StructureValidator{}

	tests := []struct {
		name         string
		registryName string
		wantErr      bool
	}{
		{
			name:         "valid registry name",
			registryName: "my-awesome-registry",
			wantErr:      false,
		},
		{
			name:         "valid registry with numbers",
			registryName: "registry-v2",
			wantErr:      false,
		},
		{
			name:         "invalid registry - uppercase",
			registryName: "MyRegistry",
			wantErr:      true,
		},
		{
			name:         "invalid registry - spaces",
			registryName: "my registry",
			wantErr:      true,
		},
		{
			name:         "invalid registry - too short",
			registryName: "ab",
			wantErr:      true,
		},
		{
			name:         "invalid registry - special chars",
			registryName: "my_registry@test",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateRegistryName(tt.registryName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRegistryName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStructureValidator_ValidateModuleName(t *testing.T) {
	validator := &StructureValidator{}

	tests := []struct {
		name       string
		moduleName string
		wantErr    bool
	}{
		{
			name:       "valid module name",
			moduleName: "git-helpers",
			wantErr:    false,
		},
		{
			name:       "valid simple name",
			moduleName: "docker",
			wantErr:    false,
		},
		{
			name:       "invalid module - uppercase",
			moduleName: "GitHelpers",
			wantErr:    true,
		},
		{
			name:       "invalid module - spaces",
			moduleName: "git helpers",
			wantErr:    true,
		},
		{
			name:       "invalid module - too short",
			moduleName: "a",
			wantErr:    true,
		},
		{
			name:       "invalid module - underscores",
			moduleName: "git_helpers",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateModuleName(tt.moduleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModuleName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStructureValidator_ValidateShell(t *testing.T) {
	validator := &StructureValidator{}

	tests := []struct {
		name    string
		shell   string
		wantErr bool
	}{
		{
			name:    "valid shell - bash",
			shell:   "bash",
			wantErr: false,
		},
		{
			name:    "valid shell - zsh",
			shell:   "zsh",
			wantErr: false,
		},
		{
			name:    "valid shell - fish",
			shell:   "fish",
			wantErr: false,
		},
		{
			name:    "valid shell - powershell",
			shell:   "powershell",
			wantErr: false,
		},
		{
			name:    "valid shell - mixed case",
			shell:   "BASH",
			wantErr: false,
		},
		{
			name:    "invalid shell - unsupported",
			shell:   "tcsh",
			wantErr: true,
		},
		{
			name:    "invalid shell - empty",
			shell:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateShell(tt.shell)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateShell() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStructureValidator_ValidateModuleType(t *testing.T) {
	validator := &StructureValidator{}

	tests := []struct {
		name       string
		moduleType string
		wantErr    bool
	}{
		{
			name:       "valid type - aliases",
			moduleType: "aliases",
			wantErr:    false,
		},
		{
			name:       "valid type - functions",
			moduleType: "functions",
			wantErr:    false,
		},
		{
			name:       "valid type - mixed case",
			moduleType: "SCRIPTS",
			wantErr:    false,
		},
		{
			name:       "invalid type - unsupported",
			moduleType: "unknown",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateModuleType(tt.moduleType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModuleType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStructureValidator_ValidateStructure(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "registry-structure-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		setupFunc func(string) error
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid registry structure",
			setupFunc: func(dir string) error {
				return createValidRegistry(dir)
			},
			wantErr: false,
		},
		{
			name: "missing index.json",
			setupFunc: func(dir string) error {
				// Don't create index.json
				return nil
			},
			wantErr: true,
			errMsg:  "index.json not found",
		},
		{
			name: "invalid index.json - missing name",
			setupFunc: func(dir string) error {
				index := map[string]interface{}{
					"description": "Test registry",
					"version":     "1.0.0",
					"modules":     map[string]interface{}{},
				}
				return writeJSON(filepath.Join(dir, "index.json"), index)
			},
			wantErr: true,
			errMsg:  "name field is required",
		},
		{
			name: "invalid version format",
			setupFunc: func(dir string) error {
				index := map[string]interface{}{
					"name":        "test-registry",
					"description": "Test registry",
					"version":     "1.0", // Invalid version
					"modules":     map[string]interface{}{},
				}
				return writeJSON(filepath.Join(dir, "index.json"), index)
			},
			wantErr: true,
			errMsg:  "version format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}

			if err := tt.setupFunc(testDir); err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			validator := NewStructureValidator(testDir)
			err := validator.ValidateStructure()

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateStructure() expected error but got none")
					return
				}
				if tt.errMsg != "" && !containsString(err.Error(), tt.errMsg) {
					t.Errorf("ValidateStructure() error = %v, expected to contain %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateStructure() unexpected error = %v", err)
				}
			}
		})
	}
}

// Helper functions for tests

func createValidRegistry(dir string) error {
	// Create index.json
	index := map[string]interface{}{
		"name":        "test-registry",
		"description": "A test registry for validation",
		"version":     "1.0.0",
		"modules": map[string]interface{}{
			"git-helpers": map[string]interface{}{
				"name":        "git-helpers",
				"description": "Git helper functions",
				"version":     "1.0.0",
				"path":        "modules/git-helpers",
				"shell":       "bash",
			},
		},
	}
	if err := writeJSON(filepath.Join(dir, "index.json"), index); err != nil {
		return err
	}

	// Create module directory and module.json
	moduleDir := filepath.Join(dir, "modules", "git-helpers")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return err
	}

	moduleConfig := map[string]interface{}{
		"name":        "git-helpers",
		"description": "Git helper functions",
		"type":        "functions",
		"shell":       "bash",
	}
	return writeJSON(filepath.Join(moduleDir, "module.json"), moduleConfig)
}

func writeJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// (containsString and findInString helpers removed; use strings.Contains instead)