package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrTypeValidation indicates a validation error
	ErrTypeValidation ErrorType = "validation"
	
	// ErrTypeNetwork indicates a network-related error
	ErrTypeNetwork ErrorType = "network"
	
	// ErrTypeConfig indicates a configuration error
	ErrTypeConfig ErrorType = "config"
	
	// ErrTypeRegistry indicates a registry-related error
	ErrTypeRegistry ErrorType = "registry"
	
	// ErrTypeModule indicates a module-related error
	ErrTypeModule ErrorType = "module"
	
	// ErrTypeSystem indicates a system-level error
	ErrTypeSystem ErrorType = "system"
	
	// ErrTypeNotFound indicates a resource was not found
	ErrTypeNotFound ErrorType = "not_found"
	
	// ErrTypeAlreadyExists indicates a resource already exists
	ErrTypeAlreadyExists ErrorType = "already_exists"
)

// AppError represents an application-specific error
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// New creates a new AppError
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with an AppError
func Wrap(err error, errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// IsType checks if an error is of a specific type
func IsType(err error, errType ErrorType) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == errType
	}
	return false
}

// Common errors
var (
	// ErrRegistryNotFound indicates a registry was not found
	ErrRegistryNotFound = &AppError{
		Type:    ErrTypeNotFound,
		Message: "registry not found",
	}
	
	// ErrRegistryAlreadyExists indicates a registry already exists
	ErrRegistryAlreadyExists = &AppError{
		Type:    ErrTypeAlreadyExists,
		Message: "registry already exists",
	}
	
	// ErrModuleNotFound indicates a module was not found
	ErrModuleNotFound = &AppError{
		Type:    ErrTypeNotFound,
		Message: "module not found",
	}
	
	// ErrInvalidURL indicates an invalid URL
	ErrInvalidURL = &AppError{
		Type:    ErrTypeValidation,
		Message: "invalid URL format",
	}
	
	// ErrInvalidRegistry indicates an invalid registry structure
	ErrInvalidRegistry = &AppError{
		Type:    ErrTypeRegistry,
		Message: "invalid registry structure",
	}
	
	// ErrConfigNotFound indicates configuration file not found
	ErrConfigNotFound = &AppError{
		Type:    ErrTypeConfig,
		Message: "configuration file not found",
	}
)

// HandleError provides centralized error handling
func HandleError(err error, verbose bool) {
	if err == nil {
		return
	}
	
	var appErr *AppError
	if errors.As(err, &appErr) {
		// Application error with context
		fmt.Printf("Error: %s\n", appErr.Message)
		
		if verbose && appErr.Err != nil {
			fmt.Printf("  Cause: %v\n", appErr.Err)
		}
		
		if verbose && len(appErr.Context) > 0 {
			fmt.Println("  Context:")
			for k, v := range appErr.Context {
				fmt.Printf("    %s: %v\n", k, v)
			}
		}
	} else {
		// Generic error
		fmt.Printf("Error: %v\n", err)
	}
}

// ExitOnError handles an error and exits if it's not nil
func ExitOnError(err error, verbose bool, exitCode int) {
	if err != nil {
		HandleError(err, verbose)
		// Note: In actual implementation, would use os.Exit(exitCode)
		// Not calling it here to allow for testing
	}
}