package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"drivio/pkg/config"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Client represents a GitLab client
type Client struct {
	client *gitlab.Client
	config *config.Config
}

// NewClient creates a new GitLab client
func NewClient(cfg *config.Config) (*Client, error) {
	var client *gitlab.Client
	var err error

	// For public repositories, we can create a client without token
	if cfg.IsPublicRepository() && cfg.GitLabToken == "" {
		client, err = gitlab.NewClient("", gitlab.WithBaseURL(cfg.GitLabURL))
	} else {
		client, err = gitlab.NewClient(cfg.GitLabToken, gitlab.WithBaseURL(cfg.GitLabURL))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// GetFile retrieves a file from a GitLab repository
func (c *Client) GetFile(ctx context.Context) ([]byte, error) {
	owner, name := c.config.GetRepositoryOwnerAndName()
	if owner == "" || name == "" {
		return nil, fmt.Errorf("invalid repository path: %s", c.config.RepositoryPath)
	}

	// Get the file content
	file, resp, err := c.client.RepositoryFiles.GetFile(
		owner+"/"+name,
		c.config.FilePath,
		&gitlab.GetFileOptions{
			Ref: &c.config.Branch,
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("file not found: %s in branch %s", c.config.FilePath, c.config.Branch)
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	// The content is already decoded in the File struct
	return []byte(file.Content), nil
}

// ValidateConnection tests the connection to GitLab
func (c *Client) ValidateConnection(ctx context.Context) error {
	// For public repositories without token, skip user validation
	if c.config.IsPublicRepository() && c.config.GitLabToken == "" {
		return nil
	}

	_, resp, err := c.client.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("invalid GitLab token or insufficient permissions")
		}
		return fmt.Errorf("failed to validate GitLab connection: %w", err)
	}
	return nil
}

// GetRepositoryInfo retrieves basic information about the repository
func (c *Client) GetRepositoryInfo(ctx context.Context) (*gitlab.Project, error) {
	owner, name := c.config.GetRepositoryOwnerAndName()
	if owner == "" || name == "" {
		return nil, fmt.Errorf("invalid repository path: %s", c.config.RepositoryPath)
	}

	project, resp, err := c.client.Projects.GetProject(
		owner+"/"+name,
		nil,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("repository not found: %s", c.config.RepositoryPath)
		}
		return nil, fmt.Errorf("failed to get repository info: %w", err)
	}

	return project, nil
}
