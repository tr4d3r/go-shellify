# Configuration Examples

This directory contains example configuration files for go-shellify in both JSON and YAML formats.

## Files

- **`config.example.json`** - Complete JSON configuration example
- **`config.example.yaml`** - Complete YAML configuration example

## Usage

### Using JSON Configuration

1. Copy the example JSON configuration:
   ```bash
   cp examples/config/config.example.json ~/.go-shellify/config.json
   ```

2. Edit the configuration to match your preferences:
   ```bash
   nano ~/.go-shellify/config.json
   ```

### Using YAML Configuration

1. Copy the example YAML configuration:
   ```bash
   cp examples/config/config.example.yaml ~/.go-shellify/config.yaml
   ```

2. Edit the configuration to match your preferences:
   ```bash
   nano ~/.go-shellify/config.yaml
   ```

## Configuration Options

### Shell Configuration

```yaml
shell:
  auto_detect: true    # Automatically detect current shell
  type: "zsh"         # Override shell type (bash, zsh, fish, powershell)
```

### Output Configuration

```yaml
output:
  directory: "~/.go-shellify/generated"  # Where to generate scripts
  filename: "go-shellify"                # Base filename (extension added automatically)
```

**Custom Output Examples:**
- Generate to `.zshrc-test`: Set `directory: "~"` and `filename: ".zshrc-test"`
- Generate to custom directory: Set `directory: "/path/to/custom"` and `filename: "my-profile"`

### Module Configuration

```yaml
modules:
  enabled:
    - "git-shortcuts"     # List of enabled modules
    - "docker-basics"
    - "productivity-tools"
  registries:
    - "https://github.com/example/shellify-registry.git"  # Registry URLs
```

### Generation Options

```yaml
generation:
  verbose: true              # Enable verbose output during generation
  backup_existing: true      # Backup existing files before overwriting
  integration_mode: "source" # How to integrate: "source" or "manual"
```

## Format Auto-Detection

go-shellify automatically detects the configuration format based on the file extension:

- `.json` files are parsed as JSON
- `.yaml` and `.yml` files are parsed as YAML
- Default fallback is JSON format

## Validation

Both formats support the same configuration options and are validated identically. You can:

- Use either format interchangeably
- Convert between formats by copying values
- Mix and match (though only one config file will be used at a time)

## Priority Order

When multiple config files exist, go-shellify checks in this order:

1. `config.yaml`
2. `config.yml`
3. `config.json`

The first file found will be used.