# Malice - Modernized for Ubuntu 22.04 and Go 1.21+

## Recent Modernization Changes

This version of Malice has been updated to work with modern systems (Ubuntu 22.04 LTS) and current Go versions (1.21+).

### Key Updates:

1. **Go Module Management**
   - Migrated from Gopkg.toml (dep) to go.mod (Go Modules)
   - All dependencies updated to latest stable versions
   - Removed vendored dependencies (use `go mod download` instead)

2. **Logging Framework**
   - Updated: `github.com/Sirupsen/logrus` → `github.com/sirupsen/logrus` (lowercase)
   - Uses modern logrus API

3. **CLI Framework**
   - Updated: `github.com/urfave/cli` → `github.com/urfave/cli/v2`
   - Uses modern CLI flag definitions with pointers

4. **Context Handling**
   - Updated: `golang.org/x/net/context` → `context` (standard library)
   - All code uses Go 1.7+ standard context package

5. **Docker Support**
   - Updated: Docker API client to v24.0.7
   - docker-compose.yml updated to version 3.8 format
   - Uses modern Elasticsearch 8.10.0 with Kibana

6. **Build System**
   - Modern Makefile with standard targets
   - Multi-stage Docker build
   - Ubuntu 22.04 compatible Dockerfile

7. **API Server**
   - Removed deprecated Docker API middleware
   - Simplified to use standard http.Handler pattern
   - Compatible with modern Go HTTP server practices

## Prerequisites for Ubuntu 22.04

```bash
# Update system
sudo apt-get update
sudo apt-get upgrade -y

# Install Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify Go installation
go version

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
newgrp docker

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install Git
sudo apt-get install -y git
```

## Installation

```bash
# Clone repository
git clone https://github.com/maliceio/malice.git
cd malice

# Setup development environment
make setup

# Build the project
make build

# Install globally
make install

# Run tests
make test

# View coverage
make coverage
```

## Running Malice

### As CLI
```bash
# With debug mode
./build/malice -D

# View help
./build/malice --help
```

### Using Docker
```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Using docker-compose
docker-compose up -d

# Access Kibana
# http://localhost:5601
```

### Development
```bash
# Run in development mode with auto-rebuild
make run

# Format code
make fmt

# Run linters
make lint

# Clean build artifacts
make clean
```

## Available Make Targets

```
help              Display this help screen
setup             Setup development environment - install dependencies
fmt               Format Go code
lint              Run linters (requires golangci-lint)
build             Build the malice binary
test              Run tests
coverage          Run tests and generate coverage report
clean             Clean build artifacts
docker-build      Build Docker image for Ubuntu 22.04
docker-run        Run Docker container
run               Run the malice binary
install           Install the binary to GOPATH/bin
release           Create a release
deps              Show dependencies
deps-update       Update all dependencies to latest compatible versions
```

## Docker Compose Services

The modernized docker-compose.yml includes:

- **Elasticsearch 8.10.0**: Search and analytics engine
- **Kibana 8.10.0**: Visualization and exploration
- Health checks for automatic restart
- Ubuntu 22.04 compatible configuration

## Configuration

Edit `config/config.toml` to customize:
- Web server settings
- Email notifications
- Database connection
- Logger configuration

## Troubleshooting

### Build Issues
```bash
# Clean and rebuild from scratch
make clean
make setup
make build
```

### Docker Issues
```bash
# Check Docker daemon
docker info

# View container logs
docker-compose logs -f elasticsearch

# Rebuild images
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Module Issues
```bash
# Update all dependencies
make deps-update

# Clean module cache
go clean -modcache
make setup
```

## Project Structure

```
├── api/              - REST API server
├── cmd/              - Command-line commands
├── commands/         - CLI command implementations
├── config/           - Configuration management
├── malice/           - Core malice library
│   ├── docker/       - Docker client
│   ├── database/     - Database interface
│   ├── logger/       - Logging setup
│   └── ...
├── plugins/          - Plugin system
├── web/              - Web frontend
├── go.mod            - Go module definition
├── go.sum            - Dependency checksums
├── Dockerfile        - Multi-stage Docker build
├── docker-compose.yml - Service orchestration
└── Makefile          - Build targets
```

## Development Guidelines

- Use Go 1.21+ idioms
- Follow standard library patterns for context
- Use `github.com/sirupsen/logrus` for logging
- Test with `go test ./...`
- Format code with `gofmt` and `goimports`

## Security Notes

- Elasticsearch is configured without authentication in compose file - add security for production
- Docker socket is mounted in containers - use with caution
- Consider network policies and firewalls in production

## Contributing

See CONTRIBUTING.md for guidelines.

## License

Apache License 2.0 - See LICENSE file

## Support

For issues, visit: https://github.com/maliceio/malice/issues
For documentation: https://github.com/maliceio/malice/docs
