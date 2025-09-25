package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/griffin/go-shellify/internal/registry"
)

// FishGenerator generates Fish shell compatible scripts
type FishGenerator struct{}

// Generate creates a Fish shell script from modules
func (g *FishGenerator) Generate(moduleList []registry.Module, outputPath string, verbose bool) (string, error) {
	// Filter modules for current platform and shell
	filtered := filterModulesForPlatform(moduleList)
	filtered = filterModulesForShell(filtered, "fish")

	var script strings.Builder

	// Add header
	script.WriteString(generateHeader("fish"))

	// Add PATH safety check to ensure basic commands are available
	script.WriteString(addPathSafetyCheckFish())

	// Add verbose flag check
	if verbose {
		script.WriteString("# Verbose mode enabled\n")
		script.WriteString("echo \"Loading go-shellify configuration...\"\n\n")
	}

	// Process each module
	for _, module := range filtered {
		if verbose {
			script.WriteString(fmt.Sprintf("echo \"Loading module: %s\"\n", module.Name))
		}

		script.WriteString(generateModuleSection(module.Name, module.Description, "fish"))

		// Add environment variables
		for _, env := range module.Environment {
			value := escapeString(env.Value, "fish")
			if env.Export {
				script.WriteString(fmt.Sprintf("set -gx %s '%s'\n", env.Name, value))
			} else {
				script.WriteString(fmt.Sprintf("set -g %s '%s'\n", env.Name, value))
			}
		}

		// Add aliases
		for _, alias := range module.Aliases {
			command := escapeString(alias.Command, "fish")
			script.WriteString(fmt.Sprintf("alias %s='%s'\n", alias.Name, command))
		}

		// Add functions
		for _, function := range module.Functions {
			script.WriteString(fmt.Sprintf("function %s\n", function.Name))

			// Add function description as comment
			if function.Description != "" {
				script.WriteString(fmt.Sprintf("    # %s\n", function.Description))
			}

			// Add function body
			for _, command := range function.Commands {
				script.WriteString(fmt.Sprintf("    %s\n", command))
			}

			script.WriteString("end\n")
		}

		// Add path entries using conditional functions
		for _, pathEntry := range module.PathEntries {
			// Use 'path' field if available, fallback to 'directory' for backward compatibility
			pathValue := pathEntry.Path
			if pathValue == "" {
				pathValue = pathEntry.Directory
			}
			pathValue = escapeString(pathValue, "fish")

			// Use conditional PATH functions based on prepend setting
			if pathEntry.Prepend {
				script.WriteString(fmt.Sprintf("add_to_path \"%s\"\n", pathValue))
			} else {
				script.WriteString(fmt.Sprintf("add_to_path_end \"%s\"\n", pathValue))
			}
		}

		// Add files to be sourced
		for _, file := range module.Files {
			if file.Source {
				script.WriteString(fmt.Sprintf("test -f '%s'; and source '%s'\n",
					escapeString(file.Path, "fish"),
					escapeString(file.Path, "fish")))
			} else if file.Execute {
				script.WriteString(fmt.Sprintf("test -x '%s'; and '%s'\n",
					escapeString(file.Path, "fish"),
					escapeString(file.Path, "fish")))
			}
		}

		// Add checks with conditional execution
		for _, check := range module.Checks {
			script.WriteString(fmt.Sprintf("# Check: %s\n", check.Description))

			switch check.Type {
			case "command":
				// For command checks, actually execute the command and check its exit status
				script.WriteString(fmt.Sprintf("if %s >/dev/null 2>&1\n", check.Command))
			case "file":
				script.WriteString(fmt.Sprintf("if test -f '%s'\n", escapeString(check.Path, "fish")))
			case "directory":
				script.WriteString(fmt.Sprintf("if test -d '%s'\n", escapeString(check.Path, "fish")))
			case "env":
				script.WriteString(fmt.Sprintf("if set -q %s\n", check.Variable))
			}

			// Add success commands
			for _, command := range check.OnSuccess {
				script.WriteString(fmt.Sprintf("    %s\n", command))
			}

			if len(check.OnFailure) > 0 {
				script.WriteString("else\n")
				for _, command := range check.OnFailure {
					script.WriteString(fmt.Sprintf("    %s\n", command))
				}
			}

			script.WriteString("end\n")
		}

		script.WriteString("\n")
	}

	// Add footer
	if verbose {
		script.WriteString("echo \"go-shellify configuration loaded successfully\"\n")
	}

	// Resolve output path and write file
	outputFile := resolveOutputPath(outputPath, "fish")

	err := os.WriteFile(outputFile, []byte(script.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write script file: %w", err)
	}

	return outputFile, nil
}

// addPathSafetyCheckFish adds smart PATH management for fish shell
func addPathSafetyCheckFish() string {
	return `# ================================================
# SMART PATH CONFIGURATION - Works on any Mac
# ================================================

# Function to safely add to PATH (prepend - highest priority)
function add_to_path
    if test -d $argv[1]; and not contains $argv[1] $PATH
        set -gx PATH $argv[1] $PATH
        if test "$VERBOSE" = "true"
            echo "Added to PATH (priority): $argv[1]"
        end
    end
end

# Function to safely add to PATH (append - standard priority)
function add_to_path_end
    if test -d $argv[1]; and not contains $argv[1] $PATH
        set -gx PATH $PATH $argv[1]
        if test "$VERBOSE" = "true"
            echo "Added to PATH (standard): $argv[1]"
        end
    end
end

# Start with essential system paths
if not contains /usr/bin $PATH; or not contains /bin $PATH
    set -gx PATH /usr/bin /bin /usr/sbin /sbin $PATH
    if test "$VERBOSE" = "true"
        echo "Added essential system paths to PATH"
    end
end

# Add macOS system paths conditionally
add_to_path_end "/System/Cryptexes/App/usr/bin"
add_to_path_end "/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/local/bin"
add_to_path_end "/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/bin"
add_to_path_end "/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/appleinternal/bin"

# Add common development tool paths (highest priority)
add_to_path "/opt/homebrew/bin"              # Homebrew (Apple Silicon)
add_to_path "/usr/local/bin"                 # Homebrew (Intel) + local binaries
add_to_path "/opt/homebrew/sbin"             # Homebrew system binaries
add_to_path "/usr/local/sbin"                # Local system binaries

# Verify essential commands are available
for cmd in mkdir sed grep awk
    if not type -q $cmd
        echo "Warning: Essential command '$cmd' not found in PATH"
    end
end

`
}

// GetFileExtension returns the file extension for Fish scripts
func (g *FishGenerator) GetFileExtension() string {
	return ".fish"
}

// GetShellType returns the shell type
func (g *FishGenerator) GetShellType() string {
	return "fish"
}