package module

import (
	"fmt"
	"sort"
	"strings"

	"github.com/griffin/go-shellify/internal/registry"
)

// ModuleInfo represents module information with registry context
type ModuleInfo struct {
	registry.Module
	RegistryName string `json:"registry_name"`
	RegistryURL  string `json:"registry_url"`
}

// Service provides module discovery and management
type Service struct {
	registryClient *registry.Client
}

// NewService creates a new module service
func NewService(registryClient *registry.Client) *Service {
	return &Service{
		registryClient: registryClient,
	}
}

// ListAllModules lists all modules from all registered registries
func (s *Service) ListAllModules() ([]ModuleInfo, error) {
	var allModules []ModuleInfo
	registries := s.registryClient.ListRegistries()

	for _, reg := range registries {
		index, err := s.registryClient.GetRegistryIndex(reg.URL)
		if err != nil {
			// Log error but continue with other registries
			fmt.Printf("Warning: Failed to fetch modules from registry %s: %v\n", reg.Name, err)
			continue
		}

		for _, module := range index.Modules {
			moduleInfo := ModuleInfo{
				Module:       module,
				RegistryName: reg.Name,
				RegistryURL:  reg.URL,
			}
			allModules = append(allModules, moduleInfo)
		}
	}

	// Sort modules by name for consistent output
	sort.Slice(allModules, func(i, j int) bool {
		return allModules[i].Name < allModules[j].Name
	})

	return allModules, nil
}

// ListModulesByRegistry lists modules from a specific registry
func (s *Service) ListModulesByRegistry(registryIdentifier string) ([]ModuleInfo, error) {
	registries := s.registryClient.ListRegistries()
	var targetRegistry *registry.Registry

	// Find the registry by name or URL
	for _, reg := range registries {
		if reg.Name == registryIdentifier || reg.URL == registryIdentifier {
			targetRegistry = &reg
			break
		}
	}

	if targetRegistry == nil {
		return nil, fmt.Errorf("registry not found: %s", registryIdentifier)
	}

	index, err := s.registryClient.GetRegistryIndex(targetRegistry.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch modules from registry %s: %w", targetRegistry.Name, err)
	}

	var modules []ModuleInfo
	for _, module := range index.Modules {
		moduleInfo := ModuleInfo{
			Module:       module,
			RegistryName: targetRegistry.Name,
			RegistryURL:  targetRegistry.URL,
		}
		modules = append(modules, moduleInfo)
	}

	// Sort modules by name
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})

	return modules, nil
}

// SearchModules searches for modules by name or description
func (s *Service) SearchModules(query string) ([]ModuleInfo, error) {
	allModules, err := s.ListAllModules()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matchingModules []ModuleInfo

	for _, module := range allModules {
		// Search in name and description
		if strings.Contains(strings.ToLower(module.Name), query) ||
			strings.Contains(strings.ToLower(module.Description), query) {
			matchingModules = append(matchingModules, module)
		}
	}

	return matchingModules, nil
}

// FilterModulesByShell filters modules by shell type
func (s *Service) FilterModulesByShell(shellType string) ([]ModuleInfo, error) {
	allModules, err := s.ListAllModules()
	if err != nil {
		return nil, err
	}

	var filteredModules []ModuleInfo
	for _, module := range allModules {
		if strings.EqualFold(module.Shell, shellType) || module.Shell == "" {
			filteredModules = append(filteredModules, module)
		}
	}

	return filteredModules, nil
}

// GetModuleDetails gets detailed information about a specific module
func (s *Service) GetModuleDetails(moduleName string) (*ModuleInfo, error) {
	allModules, err := s.ListAllModules()
	if err != nil {
		return nil, err
	}

	for _, module := range allModules {
		if module.Name == moduleName {
			return &module, nil
		}
	}

	return nil, fmt.Errorf("module not found: %s", moduleName)
}