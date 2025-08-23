package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	Registries []Registry `json:"registries"`
	CacheDir   string     `json:"cache_dir"`
	Shell      string     `json:"shell"`
	Platform   string     `json:"platform"`
}

// Registry represents a configured registry
type Registry struct {
	URL      string    `json:"url"`
	Name     string    `json:"name"`
	LastSync time.Time `json:"last_sync"`
}

var (
	// DefaultConfigDir is the default configuration directory
	DefaultConfigDir = filepath.Join(os.Getenv("HOME"), ".go-shellify")
	
	// DefaultConfigFile is the default configuration file path
	DefaultConfigFile = filepath.Join(DefaultConfigDir, "config.json")
	
	// DefaultCacheDir is the default cache directory
	DefaultCacheDir = filepath.Join(DefaultConfigDir, "cache")
)

// Manager handles configuration operations
type Manager struct {
	configPath string
	config     *Config
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	if configPath == "" {
		configPath = DefaultConfigFile
	}
	
	return &Manager{
		configPath: configPath,
	}
}

// Load reads the configuration from disk
func (m *Manager) Load() error {
	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Create default config
		m.config = m.defaultConfig()
		return m.Save()
	}
	
	// Read existing config
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	m.config = &config
	return nil
}

// Save writes the configuration to disk
func (m *Manager) Save() error {
	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Marshal config to JSON
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// Get returns the current configuration
func (m *Manager) Get() *Config {
	if m.config == nil {
		m.config = m.defaultConfig()
	}
	return m.config
}

// AddRegistry adds a new registry to the configuration
func (m *Manager) AddRegistry(url, name string) error {
	if m.config == nil {
		if err := m.Load(); err != nil {
			return err
		}
	}
	
	// Check if registry already exists
	for _, r := range m.config.Registries {
		if r.URL == url {
			return fmt.Errorf("registry already exists: %s", url)
		}
	}
	
	// Add new registry
	m.config.Registries = append(m.config.Registries, Registry{
		URL:      url,
		Name:     name,
		LastSync: time.Time{},
	})
	
	return m.Save()
}

// RemoveRegistry removes a registry from the configuration
func (m *Manager) RemoveRegistry(url string) error {
	if m.config == nil {
		if err := m.Load(); err != nil {
			return err
		}
	}
	
	// Find and remove registry
	var updated []Registry
	found := false
	for _, r := range m.config.Registries {
		if r.URL != url {
			updated = append(updated, r)
		} else {
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("registry not found: %s", url)
	}
	
	m.config.Registries = updated
	return m.Save()
}

// ListRegistries returns all configured registries
func (m *Manager) ListRegistries() ([]Registry, error) {
	if m.config == nil {
		if err := m.Load(); err != nil {
			return nil, err
		}
	}
	
	return m.config.Registries, nil
}

// UpdateRegistrySync updates the last sync time for a registry
func (m *Manager) UpdateRegistrySync(url string) error {
	if m.config == nil {
		if err := m.Load(); err != nil {
			return err
		}
	}
	
	// Find and update registry
	for i, r := range m.config.Registries {
		if r.URL == url {
			m.config.Registries[i].LastSync = time.Now()
			return m.Save()
		}
	}
	
	return fmt.Errorf("registry not found: %s", url)
}

// defaultConfig returns the default configuration
func (m *Manager) defaultConfig() *Config {
	return &Config{
		Registries: []Registry{},
		CacheDir:   DefaultCacheDir,
		Shell:      "auto",
		Platform:   "auto",
	}
}

// GetCacheDir returns the cache directory path
func (m *Manager) GetCacheDir() string {
	if m.config == nil {
		m.Load()
	}
	
	if m.config.CacheDir == "" {
		return DefaultCacheDir
	}
	
	// Expand home directory if needed
	if m.config.CacheDir[:2] == "~/" {
		return filepath.Join(os.Getenv("HOME"), m.config.CacheDir[2:])
	}
	
	return m.config.CacheDir
}