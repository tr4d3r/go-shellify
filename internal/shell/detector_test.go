package shell

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestDetectFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/bin/bash", "bash"},
		{"/usr/local/bin/zsh", "zsh"},
		{"/usr/bin/fish", "fish"},
		{"/usr/local/bin/powershell", "powershell"},
		{"C:\\Windows\\System32\\cmd.exe", "bash"}, // fallback on non-Windows systems
		{"/some/path/with-bash-in-name", "bash"},
		{"/usr/bin/unknown", "bash"}, // fallback
		{"powershell.exe", "powershell"},
		{"pwsh", "powershell"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("detectFromPath(%s) = %s, expected %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		shellType string
		expected  bool
	}{
		{"bash", true},
		{"zsh", true},
		{"fish", true},
		{"powershell", true},
		{"cmd", false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.shellType, func(t *testing.T) {
			result := IsSupported(tt.shellType)
			if result != tt.expected {
				t.Errorf("IsSupported(%s) = %v, expected %v", tt.shellType, result, tt.expected)
			}
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		shellType string
		expected  string
	}{
		{"bash", ".sh"},
		{"zsh", ".sh"},
		{"fish", ".fish"},
		{"powershell", ".ps1"},
		{"unknown", ".sh"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.shellType, func(t *testing.T) {
			result := GetFileExtension(tt.shellType)
			if result != tt.expected {
				t.Errorf("GetFileExtension(%s) = %s, expected %s", tt.shellType, result, tt.expected)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	// Save original environment
	originalShell := os.Getenv("SHELL")
	originalPSModulePath := os.Getenv("PSModulePath")
	
	defer func() {
		// Restore original environment
		if originalShell != "" {
			os.Setenv("SHELL", originalShell)
		} else {
			os.Unsetenv("SHELL")
		}
		if originalPSModulePath != "" {
			os.Setenv("PSModulePath", originalPSModulePath)
		} else {
			os.Unsetenv("PSModulePath")
		}
	}()

	t.Run("with SHELL environment variable", func(t *testing.T) {
		os.Setenv("SHELL", "/bin/zsh")
		
		shell, err := Detect()
		if err != nil {
			t.Errorf("Detect() returned error: %v", err)
		}
		if shell != "zsh" {
			t.Errorf("Detect() = %s, expected zsh", shell)
		}
	})

	t.Run("fallback based on OS", func(t *testing.T) {
		os.Unsetenv("SHELL")
		os.Unsetenv("PSModulePath")
		
		shell, err := Detect()
		if err != nil {
			t.Errorf("Detect() returned error: %v", err)
		}
		
		// Verify expected defaults based on OS
		switch runtime.GOOS {
		case "darwin":
			if shell != "zsh" {
				t.Errorf("On macOS, expected zsh, got %s", shell)
			}
		case "linux":
			if shell != "bash" {
				t.Errorf("On Linux, expected bash, got %s", shell)
			}
		case "windows":
			if shell != "cmd" {
				t.Errorf("On Windows, expected cmd, got %s", shell)
			}
		default:
			if shell != "bash" {
				t.Errorf("On %s, expected bash fallback, got %s", runtime.GOOS, shell)
			}
		}
	})

	if runtime.GOOS == "windows" {
		t.Run("Windows PowerShell detection", func(t *testing.T) {
			os.Unsetenv("SHELL")
			os.Setenv("PSModulePath", "C:\\Program Files\\PowerShell\\Modules")
			
			shell, err := Detect()
			if err != nil {
				t.Errorf("Detect() returned error: %v", err)
			}
			if shell != "powershell" {
				t.Errorf("With PSModulePath set, expected powershell, got %s", shell)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tests := []struct {
		shellType    string
		expectError  bool
		pathContains string
	}{
		{"bash", false, ".bash"}, // matches both .bashrc and .bash_profile
		{"zsh", false, ".zshrc"},
		{"fish", false, "config.fish"},
		{"powershell", false, "PowerShell"},
		{"cmd", true, ""},
		{"unknown", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.shellType, func(t *testing.T) {
			path, err := GetConfigPath(tt.shellType)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("GetConfigPath(%s) expected error, got nil", tt.shellType)
				}
				return
			}
			
			if err != nil {
				t.Errorf("GetConfigPath(%s) returned error: %v", tt.shellType, err)
				return
			}
			
			if !strings.Contains(path, homeDir) {
				t.Errorf("GetConfigPath(%s) = %s, should contain home directory %s", tt.shellType, path, homeDir)
			}
			
			if tt.pathContains != "" && !strings.Contains(path, tt.pathContains) {
				t.Errorf("GetConfigPath(%s) = %s, should contain %s", tt.shellType, path, tt.pathContains)
			}
		})
	}
}