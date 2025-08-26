package module

import (
	"strings"
	"testing"

	"github.com/griffin/go-shellify/internal/registry"
)

func TestFilterModules(t *testing.T) {
	service := &Service{}
	
	modules := []ModuleInfo{
		{
			Module: registry.Module{
				Name:        "git-helpers",
				Description: "Git helper functions",
				Shell:       "bash",
			},
			Category: "development",
			Platform: "darwin",
		},
		{
			Module: registry.Module{
				Name:        "docker-tools",
				Description: "Docker utilities",
				Shell:       "zsh",
			},
			Category: "devops",
			Platform: "linux",
		},
		{
			Module: registry.Module{
				Name:        "system-utils",
				Description: "System utilities",
				Shell:       "bash",
			},
			Category: "utilities",
			Platform: "darwin",
		},
	}

	tests := []struct {
		name     string
		category string
		platform string
		shell    string
		expected int
	}{
		{
			name:     "no filters",
			expected: 3,
		},
		{
			name:     "filter by category development",
			category: "development",
			expected: 1,
		},
		{
			name:     "filter by platform darwin",
			platform: "darwin",
			expected: 2,
		},
		{
			name:     "filter by shell bash",
			shell:    "bash",
			expected: 2,
		},
		{
			name:     "filter by category and platform",
			category: "development",
			platform: "darwin",
			expected: 1,
		},
		{
			name:     "filter with no matches",
			category: "nonexistent",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.FilterModules(modules, tt.category, tt.platform, tt.shell)
			if len(result) != tt.expected {
				t.Errorf("FilterModules() returned %d modules, expected %d", len(result), tt.expected)
			}
		})
	}
}

func TestModuleMatchesQuery(t *testing.T) {
	service := &Service{}
	
	module := ModuleInfo{
		Module: registry.Module{
			Name:        "git-helpers",
			Description: "Git helper functions for development",
		},
		Category: "development",
		Tags:     []string{"git", "version-control", "productivity"},
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
			name:     "match category",
			query:    "development",
			expected: true,
		},
		{
			name:     "match tag",
			query:    "productivity",
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
			// moduleMatchesQuery expects lowercase input
			query := strings.ToLower(tt.query)
			result := service.moduleMatchesQuery(module, query)
			if result != tt.expected {
				t.Errorf("moduleMatchesQuery() returned %v, expected %v for query '%s'", result, tt.expected, tt.query)
			}
		})
	}
}