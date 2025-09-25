package generator

import (
	"fmt"
	"strings"

	"github.com/griffin/go-shellify/internal/shell"
)

// New creates a new generator for the specified shell type
func New(shellType string) (Generator, error) {
	// Use shell package validation
	if !shell.IsSupported(shellType) {
		return nil, fmt.Errorf("unsupported shell type: %s", shellType)
	}

	switch strings.ToLower(shellType) {
	case "bash", "zsh", "sh":
		return &BashGenerator{shellType: shellType}, nil
	case "fish":
		return &FishGenerator{}, nil
	case "powershell", "pwsh":
		return &PowerShellGenerator{}, nil
	default:
		return nil, fmt.Errorf("unsupported shell type: %s", shellType)
	}
}