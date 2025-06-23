# Drivio

A command-line interface tool designed to help manage and update production environments efficiently and safely.

## Features

- 🚀 **Production Environment Management**: Safely update and manage production environments
- 🔒 **Security First**: Built with security best practices in mind
- 🛠️ **CLI Interface**: Easy-to-use command-line interface
- 📦 **Cross-platform**: Works on Linux, macOS, and Windows
- 🐳 **Docker Support**: Containerized deployment options
- 📄 **GitLab Integration**: Fetch configuration files from GitLab repositories
- ✅ **Validation**: Validate connections and repository access

## Installation

### Using Homebrew (macOS and Linux)

```bash
brew install yourusername/tap/drivio
```

### Manual Installation

1. Download the latest release for your platform from the [releases page](https://github.com/yourusername/drivio/releases)
2. Extract the archive
3. Move the binary to a directory in your PATH

```bash
# Example for Linux/macOS
sudo mv drivio /usr/local/bin/
```

### Using Docker

```bash
docker pull yourusername/drivio:latest
docker run --rm yourusername/drivio:latest --help
```

### From Source

```bash
git clone https://github.com/yourusername/drivio.git
cd drivio
make build
make install
```

## Usage

### Basic Commands

```bash
# Show help
drivio --help

# Show version
drivio --version

# Show available commands
drivio --help
```

### Fetch Configuration Files

The `fetch` command allows you to retrieve YAML configuration files from GitLab repositories.

#### Basic Usage

```bash
# Fetch with required token
drivio fetch --token YOUR_GITLAB_TOKEN

# Fetch specific repository and file
drivio fetch --token YOUR_TOKEN --repo owner/repo --file config/production.yaml

# Fetch from specific branch
drivio fetch --token YOUR_TOKEN --repo owner/repo --file config.yaml --branch develop

# Save to file
drivio fetch --token YOUR_TOKEN --repo owner/repo --file config.yaml --output local-config.yaml

# Validate connection only
drivio fetch --token YOUR_TOKEN --validate-only
```

#### Configuration Options

You can configure defaults using environment variables:

```bash
export GITLAB_URL="https://gitlab.com"
export GITLAB_TOKEN="your-gitlab-token"
export GITLAB_REPO_PATH="owner/repo"
export GITLAB_BRANCH="main"
export GITLAB_FILE_PATH="config/environment.yaml"
```

#### Examples

```bash
# Fetch production config
drivio fetch \
  --token $GITLAB_TOKEN \
  --repo mycompany/configs \
  --file environments/production.yaml \
  --branch main \
  --output prod-config.yaml

# Validate access to repository
drivio fetch \
  --token $GITLAB_TOKEN \
  --repo mycompany/configs \
  --validate-only

# Fetch from custom GitLab instance
drivio fetch \
  --url "https://gitlab.company.com" \
  --token $GITLAB_TOKEN \
  --repo team/project \
  --file config/app.yaml
```

## Development

### Prerequisites

- Go 1.24.4 or later
- Make (for using the Makefile)

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/drivio.git
   cd drivio
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Build the application:
   ```bash
   make build
   ```

### Development Commands

```bash
# Run the application
make run

# Run in development mode (with hot reload if air is installed)
make dev

# Run tests
make test

# Run tests with coverage
make test-coverage

# Lint the code
make lint

# Build for all platforms
make build-all

# Install the binary
make install

# Clean build artifacts
make clean
```

### Release Management

```bash
# Create a snapshot release (for testing)
make release-snapshot

# Create a full release
make release

# Validate goreleaser configuration
make validate-release
```

## Project Structure

```
drivio/
├── main.go              # Application entry point
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── Makefile             # Build and development commands
├── .goreleaser.yml      # Release configuration
├── Dockerfile           # Docker container definition
├── .gitignore           # Git ignore rules
├── README.md            # This file
├── LICENSE              # License file
├── example-config.yaml  # Example configuration file
└── pkg/
    ├── cmd/
    │   ├── root.go      # Root command implementation
    │   └── fetch.go     # Fetch command implementation
    ├── config/
    │   └── config.go    # Configuration management
    └── gitlab/
        └── client.go    # GitLab API client
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GITLAB_URL` | `https://gitlab.com` | GitLab instance URL |
| `GITLAB_TOKEN` | (required) | GitLab access token |
| `GITLAB_REPO_PATH` | `jparrill/drivio-config` | Repository path (owner/repo) |
| `GITLAB_BRANCH` | `main` | Branch name |
| `GITLAB_FILE_PATH` | `config/environment.yaml` | Path to file in repository |

### GitLab Token

You need a GitLab access token with the following permissions:
- `read_api` - To read repository files
- `read_repository` - To access repository content

To create a token:
1. Go to GitLab → Settings → Access Tokens
2. Create a new token with appropriate scopes
3. Use the token with the `--token` flag or `GITLAB_TOKEN` environment variable

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter any problems or have suggestions, please [open an issue](https://github.com/yourusername/drivio/issues) on GitHub.