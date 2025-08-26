package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ProfileConfig represents the user's profile configuration
type ProfileConfig struct {
	Version string `json:"version"`
	Shell   struct {
		AutoDetect bool   `json:"auto_detect"`
		Type       string `json:"type"`
	} `json:"shell"`
	Output struct {
		Directory string `json:"directory"`
		Filename  string `json:"filename"`
	} `json:"output"`
	Modules struct {
		Enabled    []string `json:"enabled"`
		Registries []string `json:"registries"`
	} `json:"modules"`
	Generation struct {
		Verbose         bool   `json:"verbose"`
		BackupExisting  bool   `json:"backup_existing"`
		IntegrationMode string `json:"integration_mode"` // "source" or "manual"
	} `json:"generation"`
}

const (
	ConfigVersion = "1.0.0"
	ConfigDir     = ".go-shellify"
	ConfigFile    = "config.json"
)

// DefaultConfig returns a new ProfileConfig with default values
func DefaultConfig() *ProfileConfig {
	homeDir, _ := os.UserHomeDir()
	
	return &ProfileConfig{
		Version: ConfigVersion,
		Shell: struct {
			AutoDetect bool   `json:"auto_detect"`
			Type       string `json:"type"`
		}{
			AutoDetect: true,
			Type:       "",
		},
		Output: struct {
			Directory string `json:"directory"`
			Filename  string `json:"filename"`
		}{
			Directory: filepath.Join(homeDir, ConfigDir, "generated"),
			Filename:  "go-shellify",
		},
		Modules: struct {
			Enabled    []string `json:"enabled"`
			Registries []string `json:"registries"`
		}{
			Enabled:    []string{},
			Registries: []string{},
		},
		Generation: struct {
			Verbose         bool   `json:"verbose"`
			BackupExisting  bool   `json:"backup_existing"`
			IntegrationMode string `json:"integration_mode"`
		}{
			Verbose:         false,
			BackupExisting:  true,
			IntegrationMode: "source",
		},
	}
}

// GetConfigPath returns the path to the user's profile configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	
	return filepath.Join(homeDir, ConfigDir, ConfigFile), nil
}

// GetConfigDir returns the path to the user's profile configuration directory
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	
	return filepath.Join(homeDir, ConfigDir), nil
}

// Load loads the profile configuration from the default location
func Load() (*ProfileConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("getting config path: %w", err)
	}
	
	return LoadFromPath(configPath)
}

// LoadFromPath loads the profile configuration from a specific file path
func LoadFromPath(path string) (*ProfileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file not found at %s - run 'go-shellify profile init' first", path)
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	
	var config ProfileConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	
	// Validate and migrate if needed
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// Save saves the profile configuration to the default location
func (c *ProfileConfig) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("getting config path: %w", err)
	}
	
	return c.SaveToPath(configPath)
}

// SaveToPath saves the profile configuration to a specific file path
func (c *ProfileConfig) SaveToPath(path string) error {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	
	// Marshal with pretty formatting
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	
	return nil
}

// validate ensures the configuration is valid
func (c *ProfileConfig) validate() error {
	if c.Version == "" {
		c.Version = ConfigVersion
	}
	
	// Validate integration mode
	if c.Generation.IntegrationMode != "source" && c.Generation.IntegrationMode != "manual" {
		return fmt.Errorf("invalid integration_mode '%s', must be 'source' or 'manual'", c.Generation.IntegrationMode)
	}
	
	// Ensure output directory is set
	if c.Output.Directory == "" {
		homeDir, _ := os.UserHomeDir()
		c.Output.Directory = filepath.Join(homeDir, ConfigDir, "generated")
	}
	
	// Ensure filename is set
	if c.Output.Filename == "" {
		c.Output.Filename = "go-shellify"
	}
	
	return nil
}

// AddModule adds a module to the enabled list if not already present
func (c *ProfileConfig) AddModule(moduleName string) {
	for _, existing := range c.Modules.Enabled {
		if existing == moduleName {
			return // Already enabled
		}
	}
	c.Modules.Enabled = append(c.Modules.Enabled, moduleName)
}

// RemoveModule removes a module from the enabled list
func (c *ProfileConfig) RemoveModule(moduleName string) {
	for i, existing := range c.Modules.Enabled {
		if existing == moduleName {
			c.Modules.Enabled = append(c.Modules.Enabled[:i], c.Modules.Enabled[i+1:]...)
			return
		}
	}
}

// IsModuleEnabled checks if a module is enabled
func (c *ProfileConfig) IsModuleEnabled(moduleName string) bool {
	for _, enabled := range c.Modules.Enabled {
		if enabled == moduleName || enabled == "*" {
			return true
		}
	}
	return false
}

// AddRegistry adds a registry to the list if not already present
func (c *ProfileConfig) AddRegistry(registryName string) {
	for _, existing := range c.Modules.Registries {
		if existing == registryName {
			return // Already added
		}
	}
	c.Modules.Registries = append(c.Modules.Registries, registryName)
}

// RemoveRegistry removes a registry from the list
func (c *ProfileConfig) RemoveRegistry(registryName string) {
	for i, existing := range c.Modules.Registries {
		if existing == registryName {
			c.Modules.Registries = append(c.Modules.Registries[:i], c.Modules.Registries[i+1:]...)
			return
		}
	}
}

// Exists checks if a profile configuration file exists
func Exists() bool {
	configPath, err := GetConfigPath()
	if err != nil {
		return false
	}
	
	_, err = os.Stat(configPath)
	return err == nil
}