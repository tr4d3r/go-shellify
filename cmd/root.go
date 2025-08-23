package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information
	Version   = "0.1.0-dev"
	BuildTime = "unknown"
	GitCommit = "unknown"

	// Global flags
	verboseFlag bool
	configFile  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-shellify",
	Short: "A CLI client for managing shellify module registries",
	Long: `go-shellify is a command-line tool for managing and consuming shellify module registries.

It connects to git repositories containing shell module definitions and provides tools 
to discover, validate, and install shell modules (aliases, functions, environment variables) 
across bash, zsh, fish, and PowerShell.`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help when no subcommand is provided
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file (default is $HOME/.go-shellify/config.json)")

	// Version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(`{{with .Name}}{{printf "%%s version information:\n" .}}{{end}}
  Version:    %s
  Build Time: %s
  Git Commit: %s

`, Version, BuildTime, GitCommit))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	// TODO: Implement configuration loading
	// This will be implemented in subtask 1.1.2
}