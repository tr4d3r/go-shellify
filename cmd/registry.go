package cmd

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/griffin/go-shellify/internal/errors"
	"github.com/griffin/go-shellify/internal/logger"
	"github.com/griffin/go-shellify/internal/registry"
	"github.com/spf13/cobra"
)

// registryCmd represents the registry command
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage shellify registries",
	Long: `Manage shellify module registries.

A registry is a git repository containing shell module definitions.
Use this command to add, list, remove, and validate registries.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help when no subcommand is provided
		cmd.Help()
	},
}

// registryAddCmd represents the registry add command
var registryAddCmd = &cobra.Command{
	Use:   "add <url> [name]",
	Short: "Add a new registry",
	Long: `Add a new shellify registry from a git repository URL.

The URL will be validated to ensure it points to a valid and accessible git repository.
If no name is provided, one will be generated from the repository URL.

Examples:
  go-shellify registry add https://github.com/user/shellify-registry
  go-shellify registry add https://github.com/user/registry my-registry
  go-shellify registry add git@github.com:user/registry.git`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		var name string
		
		if len(args) > 1 {
			name = args[1]
		} else {
			// Generate name from URL
			name = generateRegistryName(url)
		}
		
		logger.Info("Adding registry: %s (name: %s)", url, name)
		
		// Validate URL format and accessibility
		logger.Debug("Validating registry URL...")
		validator := registry.NewURLValidator()
		if err := validator.ValidateURL(url); err != nil {
			logger.Error("URL validation failed: %v", err)
			return errors.Wrap(err, errors.ErrTypeValidation, "Invalid registry URL").
				WithContext("url", url).
				WithContext("name", name)
		}
		logger.Debug("URL validation passed")
		
		// Create registry client and add registry
		logger.Debug("Creating registry client and cloning repository...")
		client, err := registry.NewClient()
		if err != nil {
			logger.Error("Failed to create registry client: %v", err)
			return errors.Wrap(err, errors.ErrTypeConfig, "Failed to create registry client").
				WithContext("url", url).
				WithContext("name", name)
		}
		
		if err := client.AddRegistry(url, name); err != nil {
			logger.Error("Failed to add registry: %v", err)
			return errors.Wrap(err, errors.ErrTypeConfig, "Failed to add registry").
				WithContext("url", url).
				WithContext("name", name)
		}
		
		logger.Info("Registry '%s' added successfully", name)
		fmt.Printf("Registry '%s' has been added successfully.\n", name)
		fmt.Printf("URL: %s\n", url)
		
		return nil
	},
}

// registryListCmd represents the registry list command
var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registries",
	Long:  `List all configured shellify registries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Listing configured registries")
		
		client, err := registry.NewClient()
		if err != nil {
			return errors.Wrap(err, errors.ErrTypeConfig, "Failed to create registry client")
		}
		
		registries := client.ListRegistries()
		
		if len(registries) == 0 {
			fmt.Println("No registries configured")
			fmt.Println("Use 'go-shellify registry add <url>' to add a registry")
			return nil
		}
		
		fmt.Println("Configured registries:")
		for _, reg := range registries {
			fmt.Printf("  - %s (%s)\n", reg.Name, reg.URL)
			if reg.LastSync.IsZero() {
				fmt.Println("    Never synced")
			} else {
				fmt.Printf("    Last synced: %s\n", reg.LastSync.Format("2006-01-02 15:04:05"))
			}
		}
		
		return nil
	},
}

// registryRemoveCmd represents the registry remove command
var registryRemoveCmd = &cobra.Command{
	Use:   "remove <name-or-url>",
	Short: "Remove a registry",
	Long:  `Remove a shellify registry by name or URL.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		
		logger.Info("Removing registry: %s", identifier)
		
		client, err := registry.NewClient()
		if err != nil {
			logger.Error("Failed to create registry client: %v", err)
			return errors.Wrap(err, errors.ErrTypeConfig, "Failed to create registry client").
				WithContext("identifier", identifier)
		}
		
		if err := client.RemoveRegistry(identifier); err != nil {
			logger.Error("Failed to remove registry: %v", err)
			return errors.Wrap(err, errors.ErrTypeConfig, "Failed to remove registry").
				WithContext("identifier", identifier)
		}
		
		logger.Info("Registry '%s' removed successfully", identifier)
		fmt.Printf("Registry '%s' has been removed successfully.\n", identifier)
		
		return nil
	},
}

// registryValidateCmd represents the registry validate command
var registryValidateCmd = &cobra.Command{
	Use:   "validate <url>",
	Short: "Validate a registry",
	Long:  `Validate that a git repository is a valid shellify registry.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		// TODO: Implement registry validation functionality (subtask 1.2.3)
		fmt.Printf("Validating registry: %s\n", url)
		fmt.Println("Registry validation functionality not yet implemented")
	},
}

// generateRegistryName generates a registry name from a URL
func generateRegistryName(rawURL string) string {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		// If parsing fails, use the raw URL as base
		return sanitizeName(rawURL)
	}

	var name string

	// Handle SSH URLs (git@host:path format)
	if parsedURL.Scheme == "git" || strings.Contains(rawURL, "git@") {
		// Extract from SSH format: git@github.com:user/repo.git
		parts := strings.Split(rawURL, ":")
		if len(parts) >= 2 {
			pathPart := parts[len(parts)-1]
			name = filepath.Base(pathPart)
		}
	} else if parsedURL.Path != "" {
		// Handle HTTPS URLs - use the repository name from path
		name = filepath.Base(parsedURL.Path)
	}

	// If we couldn't extract a name, use the host
	if name == "" || name == "/" || name == "." {
		if parsedURL.Host != "" {
			name = parsedURL.Host
		} else {
			name = "registry"
		}
	}

	// Remove .git suffix and sanitize
	name = strings.TrimSuffix(name, ".git")
	return sanitizeName(name)
}

// sanitizeName ensures the name is valid for use as a registry identifier
func sanitizeName(name string) string {
	// Replace invalid characters with hyphens
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "@", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, " ", "-")
	
	// Remove multiple consecutive hyphens
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	
	// Trim hyphens from start and end
	name = strings.Trim(name, "-")
	
	// Ensure name is not empty
	if name == "" {
		name = "registry"
	}
	
	return name
}

func init() {
	rootCmd.AddCommand(registryCmd)
	
	// Add subcommands to registry
	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryRemoveCmd)
	registryCmd.AddCommand(registryValidateCmd)
}