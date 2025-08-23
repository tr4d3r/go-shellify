package registry

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// URLValidator handles URL validation for git repositories
type URLValidator struct {
	httpTimeout time.Duration
	client      *http.Client
}

// NewURLValidator creates a new URL validator
func NewURLValidator() *URLValidator {
	timeout := 15 * time.Second
	return &URLValidator{
		httpTimeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// ValidateURL performs comprehensive URL validation for git repositories
func (v *URLValidator) ValidateURL(rawURL string) error {
	// Step 1: Validate URL format
	if err := v.validateURLFormat(rawURL); err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Step 2: Check URL accessibility
	if err := v.checkAccessibility(rawURL); err != nil {
		return fmt.Errorf("URL accessibility check failed: %w", err)
	}

	return nil
}

// validateURLFormat validates the URL format and checks if it's a valid git repository URL
func (v *URLValidator) validateURLFormat(rawURL string) error {
	// Parse URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include a scheme (https:// or git@)")
	}

	// Support HTTPS and SSH protocols
	switch parsedURL.Scheme {
	case "https":
		return v.validateHTTPSURL(parsedURL)
	case "git":
		return v.validateSSHURL(rawURL)
	default:
		return fmt.Errorf("unsupported URL scheme '%s', supported schemes: https, git (SSH)", parsedURL.Scheme)
	}
}

// validateHTTPSURL validates HTTPS git repository URLs
func (v *URLValidator) validateHTTPSURL(parsedURL *url.URL) error {
	// Check if host is provided
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	// Check if path exists and looks like a git repository
	if parsedURL.Path == "" || parsedURL.Path == "/" {
		return fmt.Errorf("URL must include a repository path")
	}

	// Validate common git hosting patterns
	if err := v.validateGitHostingPattern(parsedURL.Host, parsedURL.Path); err != nil {
		return err
	}

	return nil
}

// validateSSHURL validates SSH git repository URLs (git@host:path format)
func (v *URLValidator) validateSSHURL(rawURL string) error {
	// SSH URLs typically look like: git@github.com:user/repo.git
	sshPattern := regexp.MustCompile(`^git@([^:]+):([^/].+)$`)
	
	if !sshPattern.MatchString(rawURL) {
		return fmt.Errorf("invalid SSH URL format, expected: git@host:path")
	}

	matches := sshPattern.FindStringSubmatch(rawURL)
	if len(matches) != 3 {
		return fmt.Errorf("invalid SSH URL format")
	}

	host := matches[1]
	path := matches[2]

	// Validate host
	if host == "" {
		return fmt.Errorf("SSH URL must include a host")
	}

	// Validate path
	if path == "" {
		return fmt.Errorf("SSH URL must include a repository path")
	}

	return nil
}

// validateGitHostingPattern validates patterns for common git hosting services
func (v *URLValidator) validateGitHostingPattern(host, path string) error {
	// Remove trailing slash from path if present
	path = strings.TrimSuffix(path, "/")
	
	// Path should contain at least owner/repo
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) < 2 {
		return fmt.Errorf("repository path should be in format owner/repository")
	}

	// Validate based on known hosting services
	switch {
	case strings.Contains(host, "github.com"):
		return v.validateGitHubURL(pathParts)
	case strings.Contains(host, "gitlab.com") || strings.Contains(host, "gitlab"):
		return v.validateGitLabURL(pathParts)
	case strings.Contains(host, "bitbucket.org"):
		return v.validateBitbucketURL(pathParts)
	default:
		// Generic git hosting validation
		return v.validateGenericGitURL(pathParts)
	}
}

// validateGitHubURL validates GitHub-specific URL patterns
func (v *URLValidator) validateGitHubURL(pathParts []string) error {
	if len(pathParts) < 2 {
		return fmt.Errorf("GitHub URLs must be in format: owner/repository")
	}

	owner := pathParts[0]
	repo := pathParts[1]

	if owner == "" || repo == "" {
		return fmt.Errorf("both owner and repository name are required")
	}

	// Remove .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")

	// Basic validation for GitHub naming conventions
	validName := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validName.MatchString(owner) {
		return fmt.Errorf("invalid GitHub owner name: %s", owner)
	}
	if !validName.MatchString(repo) {
		return fmt.Errorf("invalid GitHub repository name: %s", repo)
	}

	return nil
}

// validateGitLabURL validates GitLab-specific URL patterns
func (v *URLValidator) validateGitLabURL(pathParts []string) error {
	if len(pathParts) < 2 {
		return fmt.Errorf("GitLab URLs must be in format: owner/repository or group/subgroup/repository")
	}

	// GitLab can have nested groups, so we need at least 2 parts
	repo := pathParts[len(pathParts)-1]
	if repo == "" {
		return fmt.Errorf("repository name is required")
	}

	// Remove .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")

	// Basic validation
	validName := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validName.MatchString(repo) {
		return fmt.Errorf("invalid repository name: %s", repo)
	}

	return nil
}

// validateBitbucketURL validates Bitbucket-specific URL patterns
func (v *URLValidator) validateBitbucketURL(pathParts []string) error {
	if len(pathParts) < 2 {
		return fmt.Errorf("Bitbucket URLs must be in format: workspace/repository")
	}

	workspace := pathParts[0]
	repo := pathParts[1]

	if workspace == "" || repo == "" {
		return fmt.Errorf("both workspace and repository name are required")
	}

	// Remove .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")

	return nil
}

// validateGenericGitURL validates generic git hosting URL patterns
func (v *URLValidator) validateGenericGitURL(pathParts []string) error {
	if len(pathParts) < 1 {
		return fmt.Errorf("repository path is required")
	}

	// For generic git hosts, just ensure we have a reasonable path structure
	repo := pathParts[len(pathParts)-1]
	if repo == "" {
		return fmt.Errorf("repository name is required")
	}

	return nil
}

// checkAccessibility performs a basic connectivity check to the repository
func (v *URLValidator) checkAccessibility(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL for accessibility check: %w", err)
	}

	// For SSH URLs, we can't easily check accessibility without SSH keys
	if parsedURL.Scheme == "git" {
		// For SSH URLs, we'll skip the accessibility check
		// In a real implementation, we might try to resolve the host
		return nil
	}

	// For HTTPS URLs, try to access the repository
	if parsedURL.Scheme == "https" {
		return v.checkHTTPSAccessibility(rawURL)
	}

	return nil
}

// checkHTTPSAccessibility checks if an HTTPS git repository is accessible
func (v *URLValidator) checkHTTPSAccessibility(rawURL string) error {
	// Try multiple common git repository endpoints
	endpoints := v.buildGitEndpoints(rawURL)

	var lastErr error
	for _, endpoint := range endpoints {
		if err := v.testEndpoint(endpoint); err != nil {
			lastErr = err
			continue
		}
		// If any endpoint is accessible, consider it valid
		return nil
	}

	return fmt.Errorf("repository not accessible at any known endpoints (last error: %v)", lastErr)
}

// buildGitEndpoints generates possible git repository endpoints to test
func (v *URLValidator) buildGitEndpoints(rawURL string) []string {
	endpoints := []string{}

	// Remove .git suffix and add various possible endpoints
	baseURL := strings.TrimSuffix(rawURL, ".git")
	
	// Try the direct URL
	endpoints = append(endpoints, rawURL)
	
	// Try with .git suffix if not present
	if !strings.HasSuffix(rawURL, ".git") {
		endpoints = append(endpoints, baseURL+".git")
	}

	// Try git info refs endpoint (standard git HTTP endpoint)
	endpoints = append(endpoints, baseURL+".git/info/refs")
	endpoints = append(endpoints, baseURL+"/info/refs")

	// Try the base repository page
	endpoints = append(endpoints, baseURL)

	return endpoints
}

// testEndpoint tests if an endpoint is accessible
func (v *URLValidator) testEndpoint(endpoint string) error {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set appropriate headers for git operations
	req.Header.Set("User-Agent", "go-shellify/1.0")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Accept various success status codes
	// Git repositories might return different codes based on authentication and setup
	switch resp.StatusCode {
	case http.StatusOK, http.StatusMovedPermanently, http.StatusFound, http.StatusUnauthorized:
		// These are all acceptable - they indicate the repository exists
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("repository not found (404)")
	case http.StatusForbidden:
		return fmt.Errorf("access forbidden (403) - repository may be private")
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}