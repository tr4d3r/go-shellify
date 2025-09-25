package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/griffin/go-shellify/internal/registry"
)

const (
	commandIndent = "    %s\n"
	commentIndent = "    # %s\n"
)

// BashGenerator generates Bash/Zsh compatible scripts
type BashGenerator struct {
	shellType string
}

// Generate creates a Bash/Zsh script from modules
func (g *BashGenerator) Generate(modules []registry.Module, outputPath string, verbose bool) (string, error) {
	// Filter modules for current platform and shell
	filtered := filterModulesForPlatform(modules)
	filtered = filterModulesForShell(filtered, g.shellType)

	var script strings.Builder

	// Add header
	script.WriteString(generateHeader("bash"))

	// Add PATH safety check to ensure basic commands are available
	script.WriteString(addPathSafetyCheck())

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

		script.WriteString(generateModuleSection(module.Name, module.Description, "bash"))

		// Add environment variables
		for _, env := range module.Environment {
			value := escapeString(env.Value, "bash")
			// Use double quotes for variables that contain $ to allow expansion
			quote := "'"
			if strings.Contains(env.Value, "$") {
				quote = "\""
			}
			if env.Export {
				script.WriteString(fmt.Sprintf("export %s=%s%s%s\n", env.Name, quote, value, quote))
			} else {
				script.WriteString(fmt.Sprintf("%s=%s%s%s\n", env.Name, quote, value, quote))
			}
		}

		// Add aliases
		for _, alias := range module.Aliases {
			command := escapeString(alias.Command, "bash")
			script.WriteString(fmt.Sprintf("alias %s='%s'\n", alias.Name, command))
		}

		// Add functions
		for _, function := range module.Functions {
			// Unset any existing alias with the same name to avoid conflicts
			script.WriteString(fmt.Sprintf("unalias %s 2>/dev/null || true\n", function.Name))
			script.WriteString(fmt.Sprintf("%s() {\n", function.Name))

			// Add function description as comment
			if function.Description != "" {
				script.WriteString(fmt.Sprintf(commentIndent, function.Description))
			}

			// Add function body
			for _, command := range function.Commands {
				script.WriteString(fmt.Sprintf(commandIndent, command))
			}

			script.WriteString("}\n")
		}

		// Add path entries using conditional functions
		for _, pathEntry := range module.PathEntries {
			// Use 'path' field if available, fallback to 'directory' for backward compatibility
			pathValue := pathEntry.Path
			if pathValue == "" {
				pathValue = pathEntry.Directory
			}
			directory := escapeString(pathValue, "bash")

			// Use conditional PATH functions based on prepend setting
			if pathEntry.Prepend {
				script.WriteString(fmt.Sprintf("add_to_path \"%s\"\n", directory))
			} else {
				script.WriteString(fmt.Sprintf("add_to_path_end \"%s\"\n", directory))
			}
		}

		// Add files to be sourced
		for _, file := range module.Files {
			if file.Source {
				script.WriteString(fmt.Sprintf("[ -f '%s' ] && source '%s'\n",
					escapeString(file.Path, "bash"),
					escapeString(file.Path, "bash")))
			} else if file.Execute {
				script.WriteString(fmt.Sprintf("[ -x '%s' ] && '%s'\n",
					escapeString(file.Path, "bash"),
					escapeString(file.Path, "bash")))
			}
		}

		// Add checks with conditional execution
		for _, check := range module.Checks {
			script.WriteString(fmt.Sprintf("# Check: %s\n", check.Description))

			switch check.Type {
			case "command":
				// For command checks, actually execute the command and check its exit status
				script.WriteString(fmt.Sprintf("if %s >/dev/null 2>&1; then\n", check.Command))
			case "file":
				script.WriteString(fmt.Sprintf("if [ -f '%s' ]; then\n", escapeString(check.Path, "bash")))
			case "directory":
				script.WriteString(fmt.Sprintf("if [ -d '%s' ]; then\n", escapeString(check.Path, "bash")))
			case "env":
				script.WriteString(fmt.Sprintf("if [ -n \"${%s}\" ]; then\n", check.Variable))
			}

			// Add success commands
			for _, command := range check.OnSuccess {
				script.WriteString(fmt.Sprintf(commandIndent, command))
			}

			if len(check.OnFailure) > 0 {
				script.WriteString("else\n")
				for _, command := range check.OnFailure {
					script.WriteString(fmt.Sprintf(commandIndent, command))
				}
			}

			script.WriteString("fi\n")
		}

		script.WriteString("\n")
	}

	// Add footer
	if verbose {
		script.WriteString("echo \"go-shellify configuration loaded successfully\"\n")
	}

	// Resolve output path and write file
	outputFile := resolveOutputPath(outputPath, g.shellType)

	err := os.WriteFile(outputFile, []byte(script.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write script file: %w", err)
	}

	return outputFile, nil
}

// GetFileExtension returns the file extension for Bash scripts
func (g *BashGenerator) GetFileExtension() string {
	return ".sh"
}

// GetShellType returns the shell type
func (g *BashGenerator) GetShellType() string {
	return g.shellType
}

// addPathSafetyCheck ensures basic system commands are available
func addPathSafetyCheck() string {
	return `# ================================================
# SMART PATH CONFIGURATION - Works on any Mac
# ================================================

# Function to safely add to PATH (prepend - highest priority)
add_to_path() {
    if [ -d "$1" ] && [[ ":$PATH:" != *":$1:"* ]]; then
        export PATH="$1:$PATH"
        [ "$VERBOSE" = "true" ] && echo "Added to PATH (priority): $1"
    fi
}

# Function to safely add to PATH (append - standard priority)
add_to_path_end() {
    if [ -d "$1" ] && [[ ":$PATH:" != *":$1:"* ]]; then
        export PATH="$PATH:$1"
        [ "$VERBOSE" = "true" ] && echo "Added to PATH (standard): $1"
    fi
}

# Start with essential system paths
if [[ ":$PATH:" != *":/usr/bin:"* ]] || [[ ":$PATH:" != *":/bin:"* ]]; then
    export PATH="/usr/bin:/bin:/usr/sbin:/sbin:$PATH"
    [ "$VERBOSE" = "true" ] && echo "Added essential system paths to PATH"
fi

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
for cmd in mkdir sed grep awk; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "Warning: Essential command '$cmd' not found in PATH"
    fi
done

`
}