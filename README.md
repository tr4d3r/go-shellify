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

```bash
# Clone the repository
git clone https://github.com/griffin/go-shellify.git
cd go-shellify

# Build the binary
go build -o go-shellify

# Install to PATH (optional)
go install
```

## Quick Start

```bash
# Add a registry
go-shellify registry add https://github.com/example/shellify-registry.git

# List available modules
go-shellify module list

# Filter modules by category
go-shellify module list --category devops

# Show module details
go-shellify module show docker-basics

# Search for modules
go-shellify module search docker
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

Configuration is stored in `~/.go-shellify/config.json`:

```json
{
  "registries": [
    {
      "url": "https://github.com/example/registry.git",
      "name": "example-registry",
      "last_sync": "2025-08-23T10:00:00Z"
    }
  ],
  "cache_dir": "~/.go-shellify/cache",
  "shell": "auto",
  "platform": "auto"
}
```

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
- [x] Cobra CLI structure
- [ ] Configuration management
- [ ] Registry management
- [ ] Module discovery
- [ ] Shell integration foundation

See [prd.md](prd.md) for detailed roadmap and development plans.