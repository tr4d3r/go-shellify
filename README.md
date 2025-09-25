# go-shellify

A CLI client for managing and consuming shellify module registries. Discover, validate, and install shell modules (aliases, functions, environment variables) across bash, zsh, fish, and PowerShell.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         go-shellify CLI                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐               │
│  │   Registry  │  │    Module    │  │    Shell     │               │
│  │  Management │  │   Discovery  │  │ Integration  │               │
│  └──────┬──────┘  └──────┬───────┘  └──────┬───────┘               │
│         │                 │                  │                       │
│  ┌──────▼─────────────────▼──────────────────▼──────┐               │
│  │            Core Configuration Manager             │               │
│  └───────────────────────┬───────────────────────────┘               │
│                          │                                           │
│  ┌───────────────────────▼───────────────────────────┐               │
│  │              Local Cache (~/.go-shellify)         │               │
│  │  ┌─────────────┐  ┌──────────┐  ┌──────────────┐ │               │
│  │  │ config.json │  │registries│  │ module cache │ │               │
│  │  └─────────────┘  └──────────┘  └──────────────┘ │               │
│  └───────────────────────────────────────────────────┘               │
│                                                                       │
└───────────────────────────┬───────────────────────────────────────────┘
                            │
                            ▼
        ┌───────────────────────────────────────────┐
        │         Remote Git Repositories          │
        │                                           │
        │  ┌─────────────────────────────────────┐ │
        │  │     Registry Structure              │ │
        │  │  ├── index.json                    │ │
        │  │  ├── categories.json               │ │
        │  │  ├── platforms.json                │ │
        │  │  └── modules/                      │ │
        │  │      ├── docker-basics/            │ │
        │  │      │   └── module.json           │ │
        │  │      └── git-shortcuts/            │ │
        │  │          └── module.json           │ │
        │  └─────────────────────────────────────┘ │
        └───────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.19 or later
- Git

### Option 1: Build from Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/griffin/go-shellify.git
cd go-shellify

# Build the binary
go build -o go-shellify

# Test the binary
./go-shellify --help
```

### Option 2: Install to Go Bin Directory

```bash
# Clone and install in one step
git clone https://github.com/griffin/go-shellify.git
cd go-shellify
go install

# Ensure Go's bin directory is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Or add to your shell profile (e.g., ~/.zshrc, ~/.bashrc):
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
```

### Option 3: Manual Global Installation

```bash
# Build the binary
go build -o go-shellify

# Install to system PATH (requires sudo)
sudo cp go-shellify /usr/local/bin/

# Or install to user bin directory (create if needed)
mkdir -p ~/.local/bin
cp go-shellify ~/.local/bin/
export PATH=$PATH:~/.local/bin
```

### Verify Installation

```bash
# Check if go-shellify is accessible
go-shellify --version

# If not found, check your PATH
echo $PATH | grep -o '[^:]*bin[^:]*'
```

## Quick Start

### 1. Initialize Profile Configuration

```bash
# Create a new profile configuration
go-shellify profile init

# Or create a YAML configuration
go-shellify profile init --format yaml
```

### 2. Configure Your Profile

Edit `~/.go-shellify/config.json` (or `.yaml`) to customize:
- Output location (e.g., `~/.zshrc-test`)
- Enabled modules
- Shell settings

Example for custom output:
```json
{
  "output": {
    "directory": "~",
    "filename": ".zshrc-test"
  },
  "modules": {
    "enabled": ["git-shortcuts", "docker-basics"]
  }
}
```

### 3. Generate Your Profile

```bash
# Generate shell script from configuration
go-shellify profile generate

# Generate with verbose output
go-shellify profile generate -v
```

### 4. Use Your Generated Profile

```bash
# Source the generated profile in your shell
source ~/.zshrc-test

# Or add to your main shell config
echo "source ~/.zshrc-test" >> ~/.zshrc
```

### Registry Management (Advanced)

```bash
# Add a registry
go-shellify registry add https://github.com/example/shellify-registry.git

# List available modules
go-shellify module list

# Show module details
go-shellify module show docker-basics
```

## Commands

### Registry Management

```bash
# Add a new registry
go-shellify registry add <git-url>

# List all registries
go-shellify registry list

# Remove a registry
go-shellify registry remove <git-url>

# Validate a registry
go-shellify registry validate <git-url>
```

### Module Discovery

```bash
# List all modules
go-shellify module list

# Filter by category
go-shellify module list --category development

# Filter by platform
go-shellify module list --platform darwin

# Filter by shell
go-shellify module list --shell zsh

# Show module details
go-shellify module show <module-name>

# Search modules
go-shellify module search <query>
```

## Module Categories

- `development` - Programming and development tools
- `devops` - DevOps and infrastructure management
- `productivity` - Productivity enhancements
- `utilities` - System utilities and helpers
- `cloud` - Cloud platform tools
- `database` - Database management
- `networking` - Network utilities
- `security` - Security tools

## Supported Shells

- **bash** - Bourne Again Shell
- **zsh** - Z Shell
- **fish** - Friendly Interactive Shell
- **powershell** - PowerShell Core

## Supported Platforms

- **darwin** - macOS
- **linux** - Linux distributions
- **windows** - Windows 10/11

## Configuration

Configuration supports both JSON and YAML formats, stored in `~/.go-shellify/config.json` or `~/.go-shellify/config.yaml`:

**JSON Format:**
```json
{
  "version": "1.0.0",
  "shell": {
    "auto_detect": true,
    "type": "zsh"
  },
  "output": {
    "directory": "~/.go-shellify/generated",
    "filename": "go-shellify"
  },
  "modules": {
    "enabled": ["git-shortcuts", "docker-basics"],
    "registries": ["https://github.com/example/shellify-registry.git"]
  },
  "generation": {
    "verbose": false,
    "backup_existing": true,
    "integration_mode": "source"
  }
}
```

**YAML Format:**
```yaml
version: 1.0.0
shell:
  auto_detect: true
  type: zsh
output:
  directory: ~/.go-shellify/generated
  filename: go-shellify
modules:
  enabled:
    - git-shortcuts
    - docker-basics
  registries:
    - https://github.com/example/shellify-registry.git
generation:
  verbose: false
  backup_existing: true
  integration_mode: source
```

See [`examples/config/`](examples/config/) for complete configuration examples including custom output paths.

## Development

### Prerequisites

- Go 1.19 or later
- Git

### Building from Source

```bash
# Clone repository
git clone https://github.com/griffin/go-shellify.git
cd go-shellify

# Install dependencies
go mod download

# Build
go build -o bin/go-shellify

# Run tests
go test ./...
```

### Project Structure

```
go-shellify/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── registry.go        # Registry commands
│   └── module.go          # Module commands
├── internal/              # Internal packages
│   ├── config/           # Configuration management
│   ├── registry/         # Registry operations
│   ├── module/           # Module handling
│   └── shell/            # Shell detection
├── pkg/                   # Public packages
├── main.go               # Entry point
└── go.mod                # Go module file
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines and contribution process.

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Status

**Current Version**: v0.1.0-dev (MVP in development)

### Phase 1 Progress

#### Subtask 1.1: Project Foundation - ✅ COMPLETED

- [x] 1.1.1: Set up proper Cobra CLI structure with root command
- [x] 1.1.2: Create configuration management system with JSON persistence
- [x] 1.1.3: Implement basic error handling patterns with typed errors
- [x] 1.1.4: Set up logging framework with structured output

#### Next Phase Ready: Subtask 1.2: Registry Management

- [ ] Registry add command with URL validation
- [ ] Git repository cloning functionality  
- [ ] Registry structure validation
- [ ] Registry list/remove commands with persistence
- [ ] Local cache management system

See [prd.md](prd.md) for detailed roadmap and development plans.
