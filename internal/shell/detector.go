package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Detect automatically detects the current shell
func Detect() (string, error) {
	// Check environment variables first
	if shell := os.Getenv("SHELL"); shell != "" {
		return detectFromPath(shell), nil
	}

	// Windows-specific detection
	if runtime.GOOS == "windows" {
		if psModulePath := os.Getenv("PSModulePath"); psModulePath != "" {
			return string(PowerShell), nil
		}
		return string(Cmd), nil
	}

	// Unix-like systems
	if parent := getParentProcess(); parent != "" {
		return detectFromPath(parent), nil
	}

	// Fallback to common shells based on OS
	switch runtime.GOOS {
	case "darwin":
		return string(Zsh), nil // macOS default since Catalina
	case "linux", "freebsd", "openbsd", "netbsd":
		return string(Bash), nil
	default:
		return string(Bash), nil
	}
}

// detectFromPath extracts shell name from a path
func detectFromPath(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".exe")

	switch base {
	case "bash":
		return string(Bash)
	case "zsh":
		return string(Zsh)
	case "fish":
		return string(Fish)
	case "powershell", "pwsh":
		return string(PowerShell)
	case "cmd":
		return string(Cmd)
	default:
		// Try to match common patterns
		lower := strings.ToLower(base)
		if strings.Contains(lower, "bash") {
			return string(Bash)
		}
		if strings.Contains(lower, "zsh") {
			return string(Zsh)
		}
		if strings.Contains(lower, "powershell") || strings.Contains(lower, "pwsh") {
			return string(PowerShell)
		}
		return string(Bash) // Default fallback
	}
}

// getParentProcess attempts to get the parent process name
func getParentProcess() string {
	// This is a simplified implementation
	// In a production system, you might want to use more sophisticated process detection

	if runtime.GOOS == "windows" {
		return ""
	}

	// Try to read from /proc/self/stat on Linux
	if data, err := os.ReadFile("/proc/self/stat"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 3 {
			// This is a simplified parsing - in reality, you'd want more robust parsing
			return fields[1]
		}
	}

	return ""
}

// GetConfigPath returns the appropriate config path for the shell
func GetConfigPath(shellType string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	switch ShellType(shellType) {
	case Bash:
		// Check for .bashrc, .bash_profile, .profile in order
		candidates := []string{
			filepath.Join(homeDir, ".bashrc"),
			filepath.Join(homeDir, ".bash_profile"),
			filepath.Join(homeDir, ".profile"),
		}
		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
		return filepath.Join(homeDir, ".bashrc"), nil

	case Zsh:
		return filepath.Join(homeDir, ".zshrc"), nil

	case Fish:
		configDir := filepath.Join(homeDir, ".config", "fish")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", fmt.Errorf("creating fish config directory: %w", err)
		}
		return filepath.Join(configDir, "config.fish"), nil

	case PowerShell:
		if runtime.GOOS == "windows" {
			// Windows PowerShell profile
			return filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"), nil
		}
		// PowerShell Core on Unix
		configDir := filepath.Join(homeDir, ".config", "powershell")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", fmt.Errorf("creating PowerShell config directory: %w", err)
		}
		return filepath.Join(configDir, "Microsoft.PowerShell_profile.ps1"), nil

	case Cmd:
		// Windows Command Prompt doesn't have a standard config file
		return "", fmt.Errorf("cmd shell doesn't support configuration files")

	default:
		return "", fmt.Errorf("unsupported shell type: %s", shellType)
	}
}