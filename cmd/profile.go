package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/griffin/go-shellify/internal/generator"
	"github.com/griffin/go-shellify/internal/logger"
	"github.com/griffin/go-shellify/internal/profile"
	"github.com/griffin/go-shellify/internal/registry"
	"github.com/griffin/go-shellify/internal/shell"
	"github.com/spf13/cobra"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage shell profile generation",
	Long: `The profile command manages shell profile generation and configuration.

Use this to initialize profile configurations, generate shell scripts from modules,
and manage your shell environment setup.`,
}

// profileInitCmd represents the profile init command
var profileInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new profile configuration",
	Long: `Creates a new profile configuration file with default settings.

This command will create a configuration file at ~/.go-shellify/config.json
(or config.yaml if --format yaml is specified) with sensible defaults based
on your current system and shell.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		force, _ := cmd.Flags().GetBool("force")

		return initProfile(format, force)
	},
}

// profileGenerateCmd represents the profile generate command
var profileGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate shell script from profile configuration",
	Long: `Generates a shell script based on your profile configuration and enabled modules.

The generated script will include:
- Environment variables from enabled modules
- Shell aliases and functions
- Smart PATH management
- Module-specific configurations

The output location is determined by your profile configuration's output settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath, _ := cmd.Flags().GetString("output")
		verbose, _ := cmd.Flags().GetBool("verbose")

		return generateProfile(outputPath, verbose)
	},
}

// profileStatusCmd represents the profile status command
var profileStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current system and profile status",
	Long: `Displays comprehensive information about your current system setup and profile configuration.

This command shows:
- Detected shell and platform information
- Profile configuration location and status
- Generated profile location and integration status
- Helpful suggestions for next steps

Use this command to debug setup issues or verify your configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showProfileStatus()
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileInitCmd)
	profileCmd.AddCommand(profileGenerateCmd)
	profileCmd.AddCommand(profileStatusCmd)

	// profile init flags
	profileInitCmd.Flags().StringP("format", "f", "json", "Configuration format (json or yaml)")
	profileInitCmd.Flags().BoolP("force", "", false, "Overwrite existing configuration")

	// profile generate flags
	profileGenerateCmd.Flags().StringP("output", "o", "", "Override output path from configuration")
	profileGenerateCmd.Flags().BoolP("verbose", "v", false, "Enable verbose generation output")
}

// initProfile initializes a new profile configuration
func initProfile(format string, force bool) error {
	logger.Info("Initializing profile configuration...")

	// Detect shell
	detectedShell, err := shell.Detect()
	if err != nil {
		logger.Warn("Could not detect shell: %v", err)
		detectedShell = "bash" // fallback
	}
	logger.Info("Detected shell: %s", detectedShell)

	// Create default config
	config := profile.DefaultConfig()

	// Set detected shell (but keep auto_detect enabled)
	if detectedShell != "" {
		config.Shell.Type = "" // Leave empty when auto_detect is true
		logger.Debug("Shell auto-detection enabled")
	}

	// Determine config path based on format
	var configPath string
	switch format {
	case "yaml", "yml":
		configPath, err = profile.GetConfigPathWithFormat(profile.FormatYAML)
	case "json":
		configPath, err = profile.GetConfigPathWithFormat(profile.FormatJSON)
	default:
		return fmt.Errorf("unsupported format: %s (use 'json' or 'yaml')", format)
	}

	if err != nil {
		return fmt.Errorf("getting config path: %w", err)
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("configuration already exists at %s (use --force to overwrite)", configPath)
	}

	// Save the configuration
	if err := config.SaveToPath(configPath); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}

	logger.Info("Profile configuration created at: %s", configPath)
	logger.Info("Edit the configuration file to customize your setup, then run 'go-shellify profile generate'")

	return nil
}

// generateProfile generates a shell script from the profile configuration
func generateProfile(outputPath string, verbose bool) error {
	logger.Info("Generating profile...")

	// Load profile configuration
	var config *profile.ProfileConfig
	var err error

	// Check if a specific config file was provided via global flag
	if configFile != "" {
		config, err = profile.LoadFromPath(configFile)
		if err != nil {
			return fmt.Errorf("loading profile configuration from %s: %w", configFile, err)
		}
	} else {
		// Try to find existing config in default locations
		config, err = profile.Load()
		if err != nil {
			return fmt.Errorf("loading profile configuration: %w", err)
		}
	}

	// Determine shell type
	var shellType string
	if config.Shell.AutoDetect {
		shellType, err = shell.Detect()
		if err != nil {
			return fmt.Errorf("detecting shell: %w", err)
		}
		logger.Debug("Auto-detected shell: %s", shellType)
	} else if config.Shell.Type != "" {
		shellType = config.Shell.Type
		logger.Debug("Using configured shell: %s", shellType)
	} else {
		return fmt.Errorf("no shell specified and auto-detection failed")
	}

	// Create generator for the shell
	gen, err := generator.New(shellType)
	if err != nil {
		return fmt.Errorf("creating generator: %w", err)
	}

	// Use output path from config if not overridden
	if outputPath == "" {
		if config.Output.Directory != "" && config.Output.Filename != "" {
			// Expand ~ in directory path
			outputDir := config.Output.Directory
			if outputDir == "~" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("getting home directory: %w", err)
				}
				outputDir = homeDir
			} else if len(outputDir) > 2 && outputDir[:2] == "~/" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("getting home directory: %w", err)
				}
				outputDir = filepath.Join(homeDir, outputDir[2:])
			}

			outputPath = filepath.Join(outputDir, config.Output.Filename)
		} else {
			return fmt.Errorf("no output path specified in configuration")
		}
	}

	// Use verbose setting from config if not overridden by flag
	if !verbose {
		verbose = config.Generation.Verbose
	}

	// Get enabled modules (for now, we'll create dummy modules since registry system is not fully integrated)
	modules := createDummyModules(config.Modules.Enabled)

	logger.Info("Generating script for shell: %s", shellType)
	logger.Info("Output path: %s", outputPath)

	if verbose {
		logger.Info("Verbose mode enabled")
		logger.Info("Enabled modules: %v", config.Modules.Enabled)
	}

	// Generate the script
	outputFile, err := gen.Generate(modules, filepath.Dir(outputPath), verbose)
	if err != nil {
		return fmt.Errorf("generating script: %w", err)
	}

	// If the generated file doesn't match our desired path, rename it
	if outputFile != outputPath {
		if err := os.Rename(outputFile, outputPath); err != nil {
			return fmt.Errorf("moving generated file to desired location: %w", err)
		}
	}

	logger.Info("Profile generated successfully: %s", outputPath)

	if shellType == "zsh" || shellType == "bash" {
		logger.Info("To use this profile, add the following to your .%src:", shellType)
		logger.Info("  source %s", outputPath)
	}

	return nil
}

// createDummyModules creates dummy modules for testing until registry integration is complete
func createDummyModules(enabledModules []string) []registry.Module {
	var modules []registry.Module

	for _, moduleName := range enabledModules {
		switch moduleName {
		case "git-shortcuts":
			modules = append(modules, registry.Module{
				Name:        "git-shortcuts",
				Description: "Useful Git shortcuts and aliases",
				Aliases: []registry.Alias{
					{Name: "gs", Command: "git status"},
					{Name: "ga", Command: "git add"},
					{Name: "gc", Command: "git commit"},
					{Name: "gp", Command: "git push"},
					{Name: "gl", Command: "git log --oneline"},
				},
			})
		case "docker-basics":
			modules = append(modules, registry.Module{
				Name:        "docker-basics",
				Description: "Basic Docker commands and shortcuts",
				Aliases: []registry.Alias{
					{Name: "dps", Command: "docker ps"},
					{Name: "di", Command: "docker images"},
					{Name: "drm", Command: "docker rm"},
					{Name: "drmi", Command: "docker rmi"},
				},
				Environment: []registry.Environment{
					{Name: "DOCKER_DEFAULT_PLATFORM", Value: "linux/amd64", Export: true},
				},
			})
		case "productivity-tools":
			modules = append(modules, registry.Module{
				Name:        "productivity-tools",
				Description: "Productivity enhancement tools",
				Functions: []registry.Function{
					{
						Name:        "mkcd",
						Description: "Create directory and change to it",
						Commands:    []string{"mkdir -p \"$1\"", "cd \"$1\""},
					},
				},
				PathEntries: []registry.PathEntry{
					{Path: "/usr/local/bin", Prepend: false},
				},
			})
		case "core-secrets":
			modules = append(modules, registry.Module{
				Name:        "core-secrets",
				Description: "Cross-platform OS secret manager integration",
				Functions: []registry.Function{
					{
						Name:        "secret_set",
						Description: "Store a secret in OS-native secret manager",
						Commands: []string{
							"local service=\"$1\"",
							"local account=\"$2\"",
							"local secret=\"$3\"",
							"",
							"if [ -z \"$service\" ] || [ -z \"$account\" ] || [ -z \"$secret\" ]; then",
							"    echo \"Usage: secret_set <service> <account> <secret>\" >&2",
							"    return 1",
							"fi",
							"",
							"case \"$(uname -s)\" in",
							"    Darwin)",
							"        security add-generic-password -U -a \"$account\" -s \"go-shellify.$service\" -w \"$secret\" 2>/dev/null",
							"        if [ $? -eq 0 ]; then",
							"            echo \"✅ Secret stored for $service/$account\"",
							"        else",
							"            echo \"❌ Failed to store secret for $service/$account\" >&2",
							"            return 1",
							"        fi",
							"        ;;",
							"    *)",
							"        echo \"❌ Unsupported platform: $(uname -s)\" >&2",
							"        return 1",
							"        ;;",
							"esac",
						},
					},
					{
						Name:        "secret_get",
						Description: "Retrieve a secret from OS-native secret manager",
						Commands: []string{
							"local service=\"$1\"",
							"local account=\"$2\"",
							"",
							"if [ -z \"$service\" ] || [ -z \"$account\" ]; then",
							"    echo \"Usage: secret_get <service> <account>\" >&2",
							"    return 1",
							"fi",
							"",
							"case \"$(uname -s)\" in",
							"    Darwin)",
							"        security find-generic-password -a \"$account\" -s \"go-shellify.$service\" -w 2>/dev/null",
							"        ;;",
							"    *)",
							"        echo \"❌ Unsupported platform: $(uname -s)\" >&2",
							"        return 1",
							"        ;;",
							"esac",
						},
					},
				},
				Aliases: []registry.Alias{
					{Name: "secrets", Command: "echo \"🔐 core-secrets module loaded. Try: secret_set github token <your-token>\""},
				},
			})
		default:
			// Create a generic module for unknown names
			modules = append(modules, registry.Module{
				Name:        moduleName,
				Description: fmt.Sprintf("Module: %s", moduleName),
			})
		}
	}

	return modules
}

// showProfileStatus displays comprehensive system and profile status information
func showProfileStatus() error {
	fmt.Println("🔍 go-shellify Profile Status")
	fmt.Println("=" + strings.Repeat("=", 40))
	fmt.Println()

	// System Information
	if err := showSystemInfo(); err != nil {
		logger.Warn("Could not gather system information: %v", err)
	}

	// Profile Configuration
	if err := showProfileConfig(); err != nil {
		logger.Warn("Could not load profile configuration: %v", err)
	}

	// Shell Integration Status
	if err := showShellIntegration(); err != nil {
		logger.Warn("Could not check shell integration: %v", err)
	}

	// Suggestions
	showSuggestions()

	return nil
}

// showSystemInfo displays detected system information
func showSystemInfo() error {
	fmt.Println("📋 System Information:")

	// Detect shell
	detectedShell, err := shell.Detect()
	if err != nil {
		fmt.Printf("  Shell:          ❌ Could not detect (%v)\n", err)
	} else {
		fmt.Printf("  Shell:          %s (auto-detected)\n", detectedShell)
	}

	// Platform information
	fmt.Printf("  Platform:       %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// Home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("  Home:           ❌ Could not detect (%v)\n", err)
	} else {
		fmt.Printf("  Home:           %s\n", homeDir)
	}

	fmt.Println()
	return nil
}

// showProfileConfig displays profile configuration status
func showProfileConfig() error {
	fmt.Println("⚙️  Profile Configuration:")

	// Try to find existing config
	configPath, format, err := profile.FindExistingConfig()
	if err != nil {
		fmt.Printf("  Config File:    ❌ No configuration found\n")
		fmt.Printf("                     Run 'go-shellify profile init' to create one\n")
		fmt.Println()
		return nil
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("  Config File:    ❌ %s (not found)\n", configPath)
		fmt.Printf("                     Run 'go-shellify profile init' to create one\n")
		fmt.Println()
		return nil
	}

	fmt.Printf("  Config File:    ✅ %s (%s format)\n", configPath, format)

	// Load and display config details
	config, err := profile.LoadFromPath(configPath)
	if err != nil {
		fmt.Printf("  Config Status:  ❌ Could not load (%v)\n", err)
		fmt.Println()
		return err
	}

	// Output path
	if config.Output.Directory != "" && config.Output.Filename != "" {
		outputDir := config.Output.Directory
		if outputDir == "~" {
			homeDir, _ := os.UserHomeDir()
			outputDir = homeDir
		} else if len(outputDir) > 2 && outputDir[:2] == "~/" {
			homeDir, _ := os.UserHomeDir()
			outputDir = filepath.Join(homeDir, outputDir[2:])
		}
		outputPath := filepath.Join(outputDir, config.Output.Filename)
		fmt.Printf("  Output Path:    %s\n", outputPath)
	} else {
		fmt.Printf("  Output Path:    ❌ Not configured\n")
	}

	// Shell configuration
	if config.Shell.AutoDetect {
		fmt.Printf("  Shell Config:   Auto-detect enabled\n")
		if config.Shell.Type != "" {
			fmt.Printf("                     Override: %s\n", config.Shell.Type)
		}
	} else if config.Shell.Type != "" {
		fmt.Printf("  Shell Config:   Fixed to %s\n", config.Shell.Type)
	} else {
		fmt.Printf("  Shell Config:   ❌ No shell configured\n")
	}

	// Enabled modules
	if len(config.Modules.Enabled) > 0 {
		fmt.Printf("  Enabled Modules: %s\n", strings.Join(config.Modules.Enabled, ", "))
	} else {
		fmt.Printf("  Enabled Modules: ❌ None enabled\n")
	}

	fmt.Println()
	return nil
}

// showShellIntegration checks if generated profile is integrated with shell
func showShellIntegration() error {
	fmt.Println("🔗 Shell Integration:")

	// Load config to get output path
	var config *profile.ProfileConfig
	var err error

	if configFile != "" {
		config, err = profile.LoadFromPath(configFile)
	} else {
		config, err = profile.Load()
	}

	if err != nil {
		fmt.Printf("  Generated Profile: ❌ No configuration found\n")
		fmt.Println()
		return nil
	}

	// Calculate output path
	if config.Output.Directory == "" || config.Output.Filename == "" {
		fmt.Printf("  Generated Profile: ❌ Output path not configured\n")
		fmt.Println()
		return nil
	}

	outputDir := config.Output.Directory
	if outputDir == "~" {
		homeDir, _ := os.UserHomeDir()
		outputDir = homeDir
	} else if len(outputDir) > 2 && outputDir[:2] == "~/" {
		homeDir, _ := os.UserHomeDir()
		outputDir = filepath.Join(homeDir, outputDir[2:])
	}
	outputPath := filepath.Join(outputDir, config.Output.Filename)

	// Check if generated file exists
	if fileInfo, err := os.Stat(outputPath); err == nil {
		fmt.Printf("  Generated Profile: ✅ %s\n", outputPath)
		fmt.Printf("                        Size: %.1fKB, Modified: %s\n",
			float64(fileInfo.Size())/1024,
			fileInfo.ModTime().Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("  Generated Profile: ❌ %s (not found)\n", outputPath)
		fmt.Printf("                        Run 'go-shellify profile generate' to create\n")
		fmt.Println()
		return nil
	}

	// Check shell integration
	shellType, err := shell.Detect()
	if err != nil {
		fmt.Printf("  Shell Integration: ❌ Could not detect shell\n")
		fmt.Println()
		return nil
	}

	shellConfigPath, err := shell.GetConfigPath(shellType)
	if err != nil {
		fmt.Printf("  Shell Integration: ❌ Could not determine shell config path\n")
		fmt.Println()
		return nil
	}

	// Check if shell config sources our generated profile
	if content, err := os.ReadFile(shellConfigPath); err == nil {
		contentStr := string(content)
		// Check for multiple possible references to the profile
		found := strings.Contains(contentStr, outputPath) ||
			strings.Contains(contentStr, config.Output.Filename) ||
			strings.Contains(contentStr, "~/"+config.Output.Filename)

		if found {
			fmt.Printf("  Shell Integration: ✅ Sourced in %s\n", shellConfigPath)
		} else {
			fmt.Printf("  Shell Integration: ❌ Not sourced in %s\n", shellConfigPath)
		}
	} else {
		fmt.Printf("  Shell Config:      ❌ %s (not found or not readable)\n", shellConfigPath)
	}

	fmt.Println()
	return nil
}

// showSuggestions provides helpful next steps based on current status
func showSuggestions() {
	fmt.Println("💡 Suggestions:")

	// Check if config exists
	if !profile.Exists() {
		fmt.Println("  • Run 'go-shellify profile init' to create a profile configuration")
		return
	}

	// Load config to provide specific suggestions
	config, err := profile.Load()
	if err != nil {
		fmt.Println("  • Fix configuration file errors and try again")
		return
	}

	suggestions := []string{}

	// Check if modules are enabled
	if len(config.Modules.Enabled) == 0 {
		suggestions = append(suggestions, "Edit your config file to enable some modules")
	}

	// Check if profile needs generation
	if config.Output.Directory != "" && config.Output.Filename != "" {
		outputDir := config.Output.Directory
		if outputDir == "~" {
			homeDir, _ := os.UserHomeDir()
			outputDir = homeDir
		}
		outputPath := filepath.Join(outputDir, config.Output.Filename)

		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			suggestions = append(suggestions, fmt.Sprintf("Run 'go-shellify profile generate' to create %s", outputPath))
		}

		// Check shell integration
		shellType, err := shell.Detect()
		if err == nil {
			shellConfigPath, err := shell.GetConfigPath(shellType)
			if err == nil {
				if content, err := os.ReadFile(shellConfigPath); err == nil {
					contentStr := string(content)
					// Check for multiple possible references to the profile
					found := strings.Contains(contentStr, outputPath) ||
						strings.Contains(contentStr, config.Output.Filename) ||
						strings.Contains(contentStr, "~/"+config.Output.Filename)

					if !found {
						suggestions = append(suggestions, fmt.Sprintf("Add 'source %s' to your %s", outputPath, shellConfigPath))
					}
				}
			}
		}
	}

	// Show suggestions
	if len(suggestions) == 0 {
		fmt.Println("  ✅ Everything looks good! Your profile setup is complete.")
	} else {
		for _, suggestion := range suggestions {
			fmt.Printf("  • %s\n", suggestion)
		}
	}

	fmt.Println()
}