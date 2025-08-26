package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/griffin/go-shellify/internal/logger"
)

// StructureValidator validates registry structure and content
type StructureValidator struct {
	repoPath string
}

// NewStructureValidator creates a new structure validator
func NewStructureValidator(repoPath string) *StructureValidator {
	return &StructureValidator{
		repoPath: repoPath,
	}
}

// ValidateStructure performs comprehensive registry structure validation
func (sv *StructureValidator) ValidateStructure() error {
	logger.Debug("Starting comprehensive registry structure validation for: %s", sv.repoPath)

	// Step 1: Validate index.json structure and content
	index, err := sv.validateIndexJSON()
	if err != nil {
		return fmt.Errorf("index.json validation failed: %w", err)
	}

	// Step 2: Validate each module definition
	if err := sv.validateModules(index.Modules); err != nil {
		return fmt.Errorf("module validation failed: %w", err)
	}

	// Step 3: Validate directory structure
	if err := sv.validateDirectoryStructure(); err != nil {
		return fmt.Errorf("directory structure validation failed: %w", err)
	}

	logger.Debug("Registry structure validation completed successfully")
	return nil
}

// validateIndexJSON validates the index.json file structure and required fields
func (sv *StructureValidator) validateIndexJSON() (*RegistryIndex, error) {
	indexFile := filepath.Join(sv.repoPath, "index.json")

	// Check if index.json exists
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("index.json not found")
	}

	// Read and parse JSON
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read index.json: %w", err)
	}

	var index RegistryIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	// Validate required fields
	if strings.TrimSpace(index.Name) == "" {
		return nil, fmt.Errorf("name field is required and cannot be empty")
	}

	if strings.TrimSpace(index.Description) == "" {
		return nil, fmt.Errorf("description field is required and cannot be empty")
	}

	if strings.TrimSpace(index.Version) == "" {
		return nil, fmt.Errorf("version field is required and cannot be empty")
	}

	// Validate version format (semantic versioning)
	if err := sv.validateSemanticVersion(index.Version); err != nil {
		return nil, fmt.Errorf("invalid version format: %w", err)
	}

	// Validate registry name format
	if err := sv.validateRegistryName(index.Name); err != nil {
		return nil, fmt.Errorf("invalid registry name: %w", err)
	}

	logger.Debug("Index.json validation passed: %s v%s", index.Name, index.Version)
	return &index, nil
}

// validateSemanticVersion validates semantic versioning format (e.g., 1.0.0, 2.1.3-beta)
func (sv *StructureValidator) validateSemanticVersion(version string) error {
	// Basic semantic versioning pattern: MAJOR.MINOR.PATCH with optional pre-release
	semverPattern := regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	
	if !semverPattern.MatchString(version) {
		return fmt.Errorf("version '%s' does not follow semantic versioning (e.g., 1.0.0)", version)
	}
	
	return nil
}

// validateRegistryName validates registry name format
func (sv *StructureValidator) validateRegistryName(name string) error {
	// Registry names should be lowercase with hyphens, no spaces or special chars
	namePattern := regexp.MustCompile(`^[a-z0-9-]+$`)
	
	if !namePattern.MatchString(name) {
		return fmt.Errorf("registry name '%s' must contain only lowercase letters, numbers, and hyphens", name)
	}
	
	if len(name) < 3 {
		return fmt.Errorf("registry name '%s' must be at least 3 characters long", name)
	}
	
	if len(name) > 50 {
		return fmt.Errorf("registry name '%s' must be 50 characters or less", name)
	}
	
	return nil
}

// validateModules validates each module definition in the registry
func (sv *StructureValidator) validateModules(modules map[string]Module) error {
	if len(modules) == 0 {
		return fmt.Errorf("registry must contain at least one module")
	}

	logger.Debug("Validating %d modules", len(modules))

	for moduleKey, module := range modules {
		if err := sv.validateSingleModule(moduleKey, module); err != nil {
			return fmt.Errorf("module '%s' validation failed: %w", moduleKey, err)
		}
	}

	return nil
}

// validateSingleModule validates a single module definition
func (sv *StructureValidator) validateSingleModule(moduleKey string, module Module) error {
	// Validate required fields
	if strings.TrimSpace(module.Name) == "" {
		return fmt.Errorf("name field is required")
	}

	// Module key should match module name
	if module.Name != moduleKey {
		return fmt.Errorf("module name '%s' does not match key '%s'", module.Name, moduleKey)
	}

	if strings.TrimSpace(module.Description) == "" {
		return fmt.Errorf("description field is required")
	}

	if strings.TrimSpace(module.Path) == "" {
		return fmt.Errorf("path field is required")
	}

	// Validate module name format
	if err := sv.validateModuleName(module.Name); err != nil {
		return fmt.Errorf("invalid module name: %w", err)
	}

	// Validate version if provided
	if module.Version != "" {
		if err := sv.validateSemanticVersion(module.Version); err != nil {
			return fmt.Errorf("invalid module version: %w", err)
		}
	}

	// Validate shell if provided
	if module.Shell != "" {
		if err := sv.validateShell(module.Shell); err != nil {
			return fmt.Errorf("invalid shell specification: %w", err)
		}
	}

	// Validate module path exists
	if err := sv.validateModulePath(module.Path); err != nil {
		return fmt.Errorf("module path validation failed: %w", err)
	}

	logger.Debug("Module '%s' validation passed", module.Name)
	return nil
}

// validateModuleName validates module name format
func (sv *StructureValidator) validateModuleName(name string) error {
	// Module names should be lowercase with hyphens, no spaces
	namePattern := regexp.MustCompile(`^[a-z0-9-]+$`)
	
	if !namePattern.MatchString(name) {
		return fmt.Errorf("module name '%s' must contain only lowercase letters, numbers, and hyphens", name)
	}
	
	if len(name) < 2 {
		return fmt.Errorf("module name '%s' must be at least 2 characters long", name)
	}
	
	if len(name) > 30 {
		return fmt.Errorf("module name '%s' must be 30 characters or less", name)
	}
	
	return nil
}

// validateShell validates shell specification
func (sv *StructureValidator) validateShell(shell string) error {
	supportedShells := map[string]bool{
		"bash":       true,
		"zsh":        true,
		"fish":       true,
		"powershell": true,
		"sh":         true,
	}

	if !supportedShells[strings.ToLower(shell)] {
		return fmt.Errorf("unsupported shell '%s', supported shells: bash, zsh, fish, powershell, sh", shell)
	}

	return nil
}

// validateModulePath validates that the module path exists and contains required files
func (sv *StructureValidator) validateModulePath(modulePath string) error {
	fullPath := filepath.Join(sv.repoPath, modulePath)

	// Check if module directory exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("module directory does not exist: %s", modulePath)
	}

	// Check if module.json exists
	moduleJsonPath := filepath.Join(fullPath, "module.json")
	if _, err := os.Stat(moduleJsonPath); os.IsNotExist(err) {
		return fmt.Errorf("module.json not found in: %s", modulePath)
	}

	// Validate module.json structure
	if err := sv.validateModuleJSON(moduleJsonPath); err != nil {
		return fmt.Errorf("module.json validation failed in %s: %w", modulePath, err)
	}

	return nil
}

// validateModuleJSON validates the structure of an individual module.json file
func (sv *StructureValidator) validateModuleJSON(moduleJsonPath string) error {
	data, err := os.ReadFile(moduleJsonPath)
	if err != nil {
		return fmt.Errorf("failed to read module.json: %w", err)
	}

	var moduleConfig map[string]interface{}
	if err := json.Unmarshal(data, &moduleConfig); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Check for required fields
	requiredFields := []string{"name", "description", "type"}
	for _, field := range requiredFields {
		if _, exists := moduleConfig[field]; !exists {
			return fmt.Errorf("missing required field '%s'", field)
		}
	}

	// Validate module type
	if moduleType, ok := moduleConfig["type"].(string); ok {
		if err := sv.validateModuleType(moduleType); err != nil {
			return fmt.Errorf("invalid module type: %w", err)
		}
	} else {
		return fmt.Errorf("type field must be a string")
	}

	return nil
}

// validateModuleType validates the module type
func (sv *StructureValidator) validateModuleType(moduleType string) error {
	validTypes := map[string]bool{
		"aliases":   true,
		"functions": true,
		"exports":   true,
		"scripts":   true,
		"config":    true,
	}

	if !validTypes[strings.ToLower(moduleType)] {
		return fmt.Errorf("unsupported module type '%s', supported types: aliases, functions, exports, scripts, config", moduleType)
	}

	return nil
}

// validateDirectoryStructure validates the overall directory structure
func (sv *StructureValidator) validateDirectoryStructure() error {
	// Check for modules directory
	modulesDir := filepath.Join(sv.repoPath, "modules")
	if stat, err := os.Stat(modulesDir); err != nil {
		if os.IsNotExist(err) {
			logger.Debug("modules directory not found, checking if modules are in root")
			// Allow modules to be in root directory structure
			return nil
		}
		return fmt.Errorf("failed to access modules directory: %w", err)
	} else if !stat.IsDir() {
		return fmt.Errorf("modules path exists but is not a directory")
	}

	logger.Debug("Directory structure validation passed")
	return nil
}