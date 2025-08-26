package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Registry represents a shellify registry
type Registry struct {
	URL         string    `json:"url"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	AddedAt     time.Time `json:"added_at"`
	LastSync    time.Time `json:"last_sync,omitempty"`
}

// RegistryIndex represents the structure of a registry's index.json
type RegistryIndex struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version,omitempty"`
	Modules     map[string]Module `json:"modules,omitempty"`
}

// Module represents a module in the registry
type Module struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Path        string `json:"path,omitempty"`
	Shell       string `json:"shell,omitempty"`
}

// Client manages registry operations
type Client struct {
	configDir string
	registries []Registry
	gitClient *GitClient
}

// NewClient creates a new registry client
func NewClient() (*Client, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".go-shellify")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	cacheDir := filepath.Join(configDir, "cache")
	gitClient := NewGitClient(cacheDir)

	client := &Client{
		configDir: configDir,
		gitClient: gitClient,
	}

	if err := client.loadRegistries(); err != nil {
		return nil, fmt.Errorf("failed to load registries: %w", err)
	}

	return client, nil
}

// AddRegistry adds a new registry after verification and cloning
func (c *Client) AddRegistry(url, name string) error {
	// Check if registry already exists
	for _, reg := range c.registries {
		if reg.URL == url {
			return fmt.Errorf("registry already exists: %s", url)
		}
		if reg.Name == name {
			return fmt.Errorf("registry name already exists: %s", name)
		}
	}

	// Clone the repository
	if err := c.gitClient.CloneRepository(url, name); err != nil {
		return fmt.Errorf("failed to clone registry: %w", err)
	}

	// Verify the cloned registry has valid structure
	if err := c.verifyLocalRegistry(name); err != nil {
		// Clean up the failed clone
		c.gitClient.RemoveRepository(name)
		return fmt.Errorf("registry structure validation failed: %w", err)
	}

	// Add registry to configuration
	registry := Registry{
		URL:      url,
		Name:     name,
		AddedAt:  time.Now(),
		LastSync: time.Now(),
	}

	c.registries = append(c.registries, registry)
	return c.saveRegistries()
}

// RemoveRegistry removes a registry by name or URL
func (c *Client) RemoveRegistry(identifier string) error {
	for i, reg := range c.registries {
		if reg.Name == identifier || reg.URL == identifier {
			// Remove the git repository from cache
			if err := c.gitClient.RemoveRepository(reg.Name); err != nil {
				// Log error but continue with registry removal
				fmt.Printf("Warning: failed to remove cached repository: %v\n", err)
			}
			
			// Remove from configuration
			c.registries = append(c.registries[:i], c.registries[i+1:]...)
			return c.saveRegistries()
		}
	}
	return fmt.Errorf("registry not found: %s", identifier)
}

// ListRegistries returns all registered registries
func (c *Client) ListRegistries() []Registry {
	return c.registries
}

// verifyLocalRegistry checks if a locally cloned registry has valid structure
func (c *Client) verifyLocalRegistry(name string) error {
	repoPath := c.gitClient.GetRepositoryPath(name)
	indexFile := filepath.Join(repoPath, "index.json")

	// Check if index.json exists
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return fmt.Errorf("registry index.json not found")
	}

	// Try to parse the index.json
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return fmt.Errorf("failed to read registry index: %w", err)
	}

	var index RegistryIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("invalid registry index JSON: %w", err)
	}

	// Basic validation - registry should have a name
	if index.Name == "" {
		return fmt.Errorf("registry index must have a name field")
	}

	return nil
}

// GetRegistryIndex loads and parses the index for a given registry by name
func (c *Client) GetRegistryIndex(registryName string) (*RegistryIndex, error) {
	// Find the registry
	var registry *Registry
	for _, reg := range c.registries {
		if reg.Name == registryName {
			registry = &reg
			break
		}
	}

	if registry == nil {
		return nil, fmt.Errorf("registry not found: %s", registryName)
	}

	// Get the local path and read index.json
	repoPath := c.gitClient.GetRepositoryPath(registry.Name)
	indexFile := filepath.Join(repoPath, "index.json")

	data, err := os.ReadFile(indexFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry index: %w", err)
	}

	var index RegistryIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to decode registry index: %w", err)
	}

	return &index, nil
}

// SyncRegistry updates a registry by pulling latest changes
func (c *Client) SyncRegistry(name string) error {
	// Find the registry
	var registryIndex int = -1
	for i, reg := range c.registries {
		if reg.Name == name {
			registryIndex = i
			break
		}
	}

	if registryIndex == -1 {
		return fmt.Errorf("registry not found: %s", name)
	}

	// Check if repository is cloned
	if !c.gitClient.IsRepositoryCloned(name) {
		// Repository not cloned, clone it
		registry := c.registries[registryIndex]
		if err := c.gitClient.CloneRepository(registry.URL, name); err != nil {
			return fmt.Errorf("failed to clone registry during sync: %w", err)
		}
	} else {
		// Update existing repository
		repoPath := c.gitClient.GetRepositoryPath(name)
		if err := c.gitClient.updateRepository(repoPath); err != nil {
			return fmt.Errorf("failed to update registry: %w", err)
		}
	}

	// Verify the registry structure after sync
	if err := c.verifyLocalRegistry(name); err != nil {
		return fmt.Errorf("registry validation failed after sync: %w", err)
	}

	// Update last sync time
	c.registries[registryIndex].LastSync = time.Now()
	return c.saveRegistries()
}

// loadRegistries loads registries from config file
func (c *Client) loadRegistries() error {
	registriesFile := filepath.Join(c.configDir, "registries.json")
	
	if _, err := os.Stat(registriesFile); os.IsNotExist(err) {
		// File doesn't exist, start with empty registries
		c.registries = []Registry{}
		return nil
	}

	data, err := os.ReadFile(registriesFile)
	if err != nil {
		return fmt.Errorf("failed to read registries file: %w", err)
	}

	if err := json.Unmarshal(data, &c.registries); err != nil {
		return fmt.Errorf("failed to parse registries file: %w", err)
	}

	return nil
}

// saveRegistries saves registries to config file
func (c *Client) saveRegistries() error {
	registriesFile := filepath.Join(c.configDir, "registries.json")
	
	data, err := json.MarshalIndent(c.registries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registries: %w", err)
	}

	if err := os.WriteFile(registriesFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write registries file: %w", err)
	}

	return nil
}