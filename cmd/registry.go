package cmd

import (
	"fmt"

	"github.com/griffin/go-shellify/internal/errors"
	"github.com/griffin/go-shellify/internal/logger"
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
	Use:   "add <url>",
	Short: "Add a new registry",
	Long:  `Add a new shellify registry from a git repository URL.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		logger.Info("Adding registry: %s", url)
		
		// TODO: Implement registry add functionality (subtask 1.2.1)
		// For now, just demonstrate error handling
		return errors.New(errors.ErrTypeSystem, "Registry add functionality not yet implemented").
			WithContext("url", url)
	},
}

// registryListCmd represents the registry list command
var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registries",
	Long:  `List all configured shellify registries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Listing configured registries")
		
		registries, err := ConfigManager.ListRegistries()
		if err != nil {
			return errors.Wrap(err, errors.ErrTypeConfig, "Failed to list registries")
		}
		
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
	Use:   "remove <url>",
	Short: "Remove a registry",
	Long:  `Remove a shellify registry.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		// TODO: Implement registry remove functionality (subtask 1.2.4)
		fmt.Printf("Removing registry: %s\n", url)
		fmt.Println("Registry remove functionality not yet implemented")
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

func init() {
	rootCmd.AddCommand(registryCmd)
	
	// Add subcommands to registry
	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryRemoveCmd)
	registryCmd.AddCommand(registryValidateCmd)
}