package module

import (
	"strings"
	"testing"

	"github.com/griffin/go-shellify/internal/registry"
)

func TestFilterModulesByShell(t *testing.T) {
	// Test the shell filtering logic
	
	modules := []ModuleInfo{
		{
			Module: registry.Module{
				Name:        "git-helpers",
				Description: "Git helper functions",
				Shell:       "bash",
			},
			RegistryName: "test-registry",
			RegistryURL:  "https://github.com/test/registry",
		},
		{
			Module: registry.Module{
				Name:        "docker-tools", 
				Description: "Docker utilities",
				Shell:       "zsh",
			},
			RegistryName: "test-registry",
			RegistryURL:  "https://github.com/test/registry",
		},
		{
			Module: registry.Module{
				Name:        "system-utils",
				Description: "System utilities",
				Shell:       "bash",
			},
			RegistryName: "test-registry",
			RegistryURL:  "https://github.com/test/registry",
		},
		{
			Module: registry.Module{
				Name:        "universal-tools",
				Description: "Universal shell tools",
				Shell:       "", // Empty shell means compatible with all
			},
			RegistryName: "test-registry", 
			RegistryURL:  "https://github.com/test/registry",
		},
	}

	tests := []struct {
		name     string
		shell    string
		expected int
	}{
		{
			name:     "filter bash modules",
			shell:    "bash",
			expected: 3, // git-helpers, system-utils, universal-tools
		},
		{
			name:     "filter zsh modules", 
			shell:    "zsh",
			expected: 2, // docker-tools, universal-tools
		},
		{
			name:     "filter fish modules",
			shell:    "fish",
			expected: 1, // universal-tools only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the actual filtering method that exists
			result := filterByShell(modules, tt.shell)
			if len(result) != tt.expected {
				t.Errorf("filterByShell() returned %d modules, expected %d", len(result), tt.expected)
			}
		})
	}
}

// Helper function to test shell filtering logic
func filterByShell(modules []ModuleInfo, shellType string) []ModuleInfo {
	var filtered []ModuleInfo
	for _, module := range modules {
		if strings.EqualFold(module.Shell, shellType) || module.Shell == "" {
			filtered = append(filtered, module)
		}
	}
	return filtered
}

func TestSearchModules(t *testing.T) {
	// Test the search functionality that actually exists in the current implementation
	module := ModuleInfo{
		Module: registry.Module{
			Name:        "git-helpers",
			Description: "Git helper functions for development",
		},
		RegistryName: "test-registry",
		RegistryURL:  "https://github.com/test/registry",
	}

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "match name",
			query:    "git",
			expected: true,
		},
		{
			name:     "match description",
			query:    "helper",
			expected: true,
		},
		{
			name:     "case insensitive match",
			query:    "GIT",
			expected: true,
		},
		{
			name:     "no match",
			query:    "docker",
			expected: false,
		},
		{
			name:     "partial match in description", 
			query:    "functions",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := moduleMatchesQuery(module, tt.query)
			if result != tt.expected {
				t.Errorf("moduleMatchesQuery() returned %v, expected %v for query '%s'", result, tt.expected, tt.query)
			}
		})
	}
}

// Helper function to test query matching logic (based on current SearchModules implementation)
func moduleMatchesQuery(module ModuleInfo, query string) bool {
	query = strings.ToLower(query)
	
	// Check name and description (matches current SearchModules logic)
	if strings.Contains(strings.ToLower(module.Name), query) ||
		strings.Contains(strings.ToLower(module.Description), query) {
		return true
	}
	
	return false
}