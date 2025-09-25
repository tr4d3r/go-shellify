package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/griffin/go-shellify/internal/registry"
)

// PowerShellGenerator generates PowerShell compatible scripts
type PowerShellGenerator struct{}

// Generate creates a PowerShell script from modules
func (g *PowerShellGenerator) Generate(moduleList []registry.Module, outputPath string, verbose bool) (string, error) {
	// Filter modules for current platform and shell
	filtered := filterModulesForPlatform(moduleList)
	filtered = filterModulesForShell(filtered, "powershell")

	var script strings.Builder

	// Add header - PowerShell uses # for comments too
	script.WriteString(generateHeader("powershell"))

	// Add verbose flag check
	if verbose {
		script.WriteString("# Verbose mode enabled\n")
		script.WriteString("Write-Host \"Loading go-shellify configuration...\"\n\n")
	}

	// Process each module
	for _, module := range filtered {
		if verbose {
			script.WriteString(fmt.Sprintf("Write-Host \"Loading module: %s\"\n", module.Name))
		}

		script.WriteString(generateModuleSection(module.Name, module.Description, "powershell"))

		// Add environment variables
		for _, env := range module.Environment {
			value := escapeString(env.Value, "powershell")
			script.WriteString(fmt.Sprintf("$env:%s = '%s'\n", env.Name, value))
		}

		// Add aliases
		for _, alias := range module.Aliases {
			command := escapeString(alias.Command, "powershell")
			script.WriteString(fmt.Sprintf("Set-Alias -Name %s -Value '%s'\n", alias.Name, command))
		}

		// Add functions
		for _, function := range module.Functions {
			script.WriteString(fmt.Sprintf("function %s {\n", function.Name))

			// Add function description as comment
			if function.Description != "" {
				script.WriteString(fmt.Sprintf("    # %s\n", function.Description))
			}

			// Add function body
			for _, command := range function.Commands {
				script.WriteString(fmt.Sprintf("    %s\n", command))
			}

			script.WriteString("}\n")
		}

		// Add PATH entries (PowerShell style)
		for _, pathEntry := range module.PathEntries {
			pathValue := pathEntry.Path
			if pathValue == "" {
				pathValue = pathEntry.Directory
			}
			directory := escapeString(pathValue, "powershell")

			if pathEntry.Prepend {
				script.WriteString(fmt.Sprintf("if (Test-Path '%s') { $env:PATH = '%s;' + $env:PATH }\n", directory, directory))
			} else {
				script.WriteString(fmt.Sprintf("if (Test-Path '%s') { $env:PATH = $env:PATH + ';%s' }\n", directory, directory))
			}
		}

		// Add files to be sourced/executed
		for _, file := range module.Files {
			if file.Source {
				script.WriteString(fmt.Sprintf("if (Test-Path '%s') { . '%s' }\n",
					escapeString(file.Path, "powershell"),
					escapeString(file.Path, "powershell")))
			} else if file.Execute {
				script.WriteString(fmt.Sprintf("if (Test-Path '%s') { & '%s' }\n",
					escapeString(file.Path, "powershell"),
					escapeString(file.Path, "powershell")))
			}
		}

		// Add checks with conditional execution
		for _, check := range module.Checks {
			script.WriteString(fmt.Sprintf("# Check: %s\n", check.Description))

			switch check.Type {
			case "command":
				script.WriteString(fmt.Sprintf("if (Get-Command '%s' 2>$null) {\n", check.Command))
			case "file":
				script.WriteString(fmt.Sprintf("if (Test-Path '%s' -PathType Leaf) {\n", escapeString(check.Path, "powershell")))
			case "directory":
				script.WriteString(fmt.Sprintf("if (Test-Path '%s' -PathType Container) {\n", escapeString(check.Path, "powershell")))
			case "env":
				script.WriteString(fmt.Sprintf("if ($env:%s) {\n", check.Variable))
			}

			// Add success commands
			for _, command := range check.OnSuccess {
				script.WriteString(fmt.Sprintf("    %s\n", command))
			}

			if len(check.OnFailure) > 0 {
				script.WriteString("} else {\n")
				for _, command := range check.OnFailure {
					script.WriteString(fmt.Sprintf("    %s\n", command))
				}
			}

			script.WriteString("}\n")
		}

		script.WriteString("\n")
	}

	// Add footer
	if verbose {
		script.WriteString("Write-Host \"go-shellify configuration loaded successfully\"\n")
	}

	// Resolve output path and write file
	outputFile := resolveOutputPath(outputPath, "powershell")

	err := os.WriteFile(outputFile, []byte(script.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write script file: %w", err)
	}

	return outputFile, nil
}

// GetFileExtension returns the file extension for PowerShell scripts
func (g *PowerShellGenerator) GetFileExtension() string {
	return ".ps1"
}

// GetShellType returns the shell type
func (g *PowerShellGenerator) GetShellType() string {
	return "powershell"
}