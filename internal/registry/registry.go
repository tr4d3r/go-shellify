package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	client := &Client{
		configDir: configDir,
	}

	if err := client.loadRegistries(); err != nil {
		return nil, fmt.Errorf("failed to load registries: %w", err)
	}

	return client, nil
}

// AddRegistry adds a new registry after verification
func (c *Client) AddRegistry(url, name string) error {
	// Verify registry is accessible and valid
	if err := c.VerifyRegistry(url); err != nil {
		return fmt.Errorf("registry verification failed: %w", err)
	}

	// Check if registry already exists
	for _, reg := range c.registries {
		if reg.URL == url {
			return fmt.Errorf("registry already exists: %s", url)
		}
		if reg.Name == name {
			return fmt.Errorf("registry name already exists: %s", name)
		}
	}

	// Add registry
	registry := Registry{
		URL:     url,
		Name:    name,
		AddedAt: time.Now(),
	}

	c.registries = append(c.registries, registry)
	return c.saveRegistries()
}

// RemoveRegistry removes a registry by name or URL
func (c *Client) RemoveRegistry(identifier string) error {
	for i, reg := range c.registries {
		if reg.Name == identifier || reg.URL == identifier {
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

// VerifyRegistry checks if a registry URL is accessible and has valid structure
func (c *Client) VerifyRegistry(url string) error {
	// Ensure URL ends with index.json or add it
	indexURL := url
	if !strings.HasSuffix(url, "index.json") {
		if !strings.HasSuffix(url, "/") {
			indexURL += "/"
		}
		indexURL += "index.json"
	}

	// Try to fetch the index.json
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(indexURL)
	if err != nil {
		return fmt.Errorf("failed to fetch registry index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry index returned status %d", resp.StatusCode)
	}

	// Try to parse as JSON
	var index RegistryIndex
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read registry index: %w", err)
	}

	if err := json.Unmarshal(body, &index); err != nil {
		return fmt.Errorf("invalid registry index JSON: %w", err)
	}

	return nil
}

// GetRegistryIndex fetches and parses the index for a given registry
func (c *Client) GetRegistryIndex(registryURL string) (*RegistryIndex, error) {
	indexURL := registryURL
	if !strings.HasSuffix(registryURL, "index.json") {
		if !strings.HasSuffix(registryURL, "/") {
			indexURL += "/"
		}
		indexURL += "index.json"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(indexURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry index returned status %d", resp.StatusCode)
	}

	var index RegistryIndex
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, fmt.Errorf("failed to decode registry index: %w", err)
	}

	return &index, nil
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