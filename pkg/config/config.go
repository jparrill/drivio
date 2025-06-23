package config

import (
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	GitLabURL      string
	GitLabToken    string
	RepositoryPath string
	Branch         string
	FilePath       string
}

// Default values
const (
	DefaultGitLabURL      = "https://gitlab.com"
	DefaultRepositoryPath = "jparrill/drivio-config"
	DefaultBranch         = "main"
	DefaultFilePath       = "config/environment.yaml"
)

// LoadConfig loads configuration from environment variables and defaults
func LoadConfig() *Config {
	config := &Config{
		GitLabURL:      getEnvOrDefault("GITLAB_URL", DefaultGitLabURL),
		GitLabToken:    getEnvOrDefault("GITLAB_TOKEN", ""),
		RepositoryPath: getEnvOrDefault("GITLAB_REPO_PATH", DefaultRepositoryPath),
		Branch:         getEnvOrDefault("GITLAB_BRANCH", DefaultBranch),
		FilePath:       getEnvOrDefault("GITLAB_FILE_PATH", DefaultFilePath),
	}

	return config
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ValidateConfig validates the configuration
func (c *Config) ValidateConfig() error {
	// For public repositories, token is optional
	// For private repositories, token is required
	if c.RepositoryPath == "" {
		return &ConfigError{Message: "Repository path is required"}
	}

	if c.FilePath == "" {
		return &ConfigError{Message: "File path is required"}
	}

	return nil
}

// IsPublicRepository checks if this is likely a public repository
func (c *Config) IsPublicRepository() bool {
	// Common public repositories that don't require authentication
	publicRepos := []string{
		"gitlab-org/gitlab-foss",
		"gitlab-org/gitlab",
		"gitlab-org/gitlab-runner",
		"gitlab-org/charts",
	}

	for _, repo := range publicRepos {
		if strings.Contains(c.RepositoryPath, repo) {
			return true
		}
	}

	return false
}

// RequiresToken checks if a token is required for this configuration
func (c *Config) RequiresToken() bool {
	// If token is provided, it's always valid
	if c.GitLabToken != "" {
		return false
	}

	// For public repositories, token is not required
	if c.IsPublicRepository() {
		return false
	}

	// For private repositories or unknown repositories, token is required
	return true
}

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

// GetRepositoryOwnerAndName extracts owner and name from repository path
func (c *Config) GetRepositoryOwnerAndName() (string, string) {
	parts := strings.Split(c.RepositoryPath, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return "", ""
}
