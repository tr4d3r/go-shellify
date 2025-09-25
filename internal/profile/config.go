package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProfileConfig represents the user's profile configuration
type ProfileConfig struct {
	Version string `json:"version" yaml:"version"`
	Shell   struct {
		AutoDetect bool   `json:"auto_detect" yaml:"auto_detect"`
		Type       string `json:"type" yaml:"type"`
	} `json:"shell" yaml:"shell"`
	Output struct {
		Directory string `json:"directory" yaml:"directory"`
		Filename  string `json:"filename" yaml:"filename"`
	} `json:"output" yaml:"output"`
	Modules struct {
		Enabled    []string `json:"enabled" yaml:"enabled"`
		Registries []string `json:"registries" yaml:"registries"`
	} `json:"modules" yaml:"modules"`
	Generation struct {
		Verbose         bool   `json:"verbose" yaml:"verbose"`
		BackupExisting  bool   `json:"backup_existing" yaml:"backup_existing"`
		IntegrationMode string `json:"integration_mode" yaml:"integration_mode"` // "source" or "manual"
	} `json:"generation" yaml:"generation"`
}

const (
	ConfigVersion = "1.0.0"
	ConfigDir     = ".go-shellify"
	ConfigFile    = "config.json"
)

// ConfigFormat represents the supported configuration file formats
type ConfigFormat string

const (
	FormatJSON ConfigFormat = "json"
	FormatYAML ConfigFormat = "yaml"
)

// DefaultConfig returns a new ProfileConfig with default values
func DefaultConfig() *ProfileConfig {
	homeDir, _ := os.UserHomeDir()
	
	return &ProfileConfig{
		Version: ConfigVersion,
		Shell: struct {
			AutoDetect bool   `json:"auto_detect" yaml:"auto_detect"`
			Type       string `json:"type" yaml:"type"`
		}{
			AutoDetect: true,
			Type:       "",
		},
		Output: struct {
			Directory string `json:"directory" yaml:"directory"`
			Filename  string `json:"filename" yaml:"filename"`
		}{
			Directory: filepath.Join(homeDir, ConfigDir, "generated"),
			Filename:  "go-shellify",
		},
		Modules: struct {
			Enabled    []string `json:"enabled" yaml:"enabled"`
			Registries []string `json:"registries" yaml:"registries"`
		}{
			Enabled:    []string{},
			Registries: []string{},
		},
		Generation: struct {
			Verbose         bool   `json:"verbose" yaml:"verbose"`
			BackupExisting  bool   `json:"backup_existing" yaml:"backup_existing"`
			IntegrationMode string `json:"integration_mode" yaml:"integration_mode"`
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
	format := detectFormat(path)

	switch format {
	case FormatYAML:
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("parsing YAML config file: %w", err)
		}
	case FormatJSON:
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("parsing JSON config file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format for file: %s", path)
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

	format := detectFormat(path)
	var data []byte
	var err error

	switch format {
	case FormatYAML:
		data, err = yaml.Marshal(c)
		if err != nil {
			return fmt.Errorf("marshaling YAML config: %w", err)
		}
	case FormatJSON:
		data, err = json.MarshalIndent(c, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config format for file: %s", path)
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

// detectFormat determines the config format based on file extension
func detectFormat(filename string) ConfigFormat {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		return FormatYAML
	case ".json":
		return FormatJSON
	default:
		return FormatJSON // default fallback
	}
}

// GetConfigPathWithFormat returns the path to the user's profile configuration file with specified format
func GetConfigPathWithFormat(format ConfigFormat) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	var filename string
	switch format {
	case FormatYAML:
		filename = "config.yaml"
	case FormatJSON:
		filename = "config.json"
	default:
		filename = "config.json"
	}

	return filepath.Join(homeDir, ConfigDir, filename), nil
}

// ExistsWithFormat checks if a profile configuration file exists for the specified format
func ExistsWithFormat(format ConfigFormat) bool {
	configPath, err := GetConfigPathWithFormat(format)
	if err != nil {
		return false
	}

	_, err = os.Stat(configPath)
	return err == nil
}

// FindExistingConfig searches for existing configuration files and returns the first one found
func FindExistingConfig() (string, ConfigFormat, error) {
	// Check for YAML first, then JSON
	formats := []ConfigFormat{FormatYAML, FormatJSON}

	for _, format := range formats {
		if ExistsWithFormat(format) {
			path, err := GetConfigPathWithFormat(format)
			if err != nil {
				continue
			}
			return path, format, nil
		}
	}

	// No existing config found, return default JSON path
	path, err := GetConfigPath()
	if err != nil {
		return "", FormatJSON, fmt.Errorf("getting default config path: %w", err)
	}

	return path, FormatJSON, nil
}

// Exists checks if a profile configuration file exists (checks both formats)
func Exists() bool {
	_, _, err := FindExistingConfig()
	return err == nil && (ExistsWithFormat(FormatJSON) || ExistsWithFormat(FormatYAML))
}