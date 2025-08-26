package generator

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		wantType  string
		wantErr   bool
	}{
		{
			name:      "bash generator",
			shellType: "bash",
			wantType:  "bash",
			wantErr:   false,
		},
		{
			name:      "zsh generator",
			shellType: "zsh",
			wantType:  "zsh",
			wantErr:   false,
		},
		{
			name:      "fish generator",
			shellType: "fish",
			wantType:  "fish",
			wantErr:   false,
		},
		{
			name:      "powershell generator",
			shellType: "powershell",
			wantType:  "powershell",
			wantErr:   false,
		},
		{
			name:      "unsupported shell",
			shellType: "tcsh",
			wantErr:   true,
		},
		{
			name:      "cmd shell not supported",
			shellType: "cmd",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := New(tt.shellType)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("New(%s) expected error, got nil", tt.shellType)
				}
				return
			}
			
			if err != nil {
				t.Errorf("New(%s) unexpected error: %v", tt.shellType, err)
				return
			}
			
			if generator == nil {
				t.Errorf("New(%s) returned nil generator", tt.shellType)
				return
			}
			
			// Test that generator returns expected shell type
			shellType := generator.GetShellType()
			if shellType != tt.wantType {
				t.Errorf("Generator.GetShellType() = %s, want %s", shellType, tt.wantType)
			}
			
			// Test that generator has expected file extension
			ext := generator.GetFileExtension()
			if ext == "" {
				t.Errorf("Generator.GetFileExtension() returned empty string")
			}
		})
	}
}