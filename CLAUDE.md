# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build & Development
- `go build -v ./...` - Build all packages
- `go run main.go` - Run the FTP server locally
- `go run main.go -conf ftpserver.json` - Run with specific config file
- `go run main.go -conf-only` - Generate configuration file only

### Testing
- `go test -race -v ./...` - Run all tests with race detection
- Individual test: `go test -v ./fs/utils/`

### Linting & Code Quality
- `golangci-lint run` - Run comprehensive linting (configured via .golangci.yml)
- `go fmt ./...` - Format code
- `goimports -w ./` - Fix imports

## Architecture

### Core Structure
This is a Go FTP server that acts as a gateway between FTP clients and various cloud storage backends. It's built on top of [ftpserverlib](https://github.com/fclairamb/ftpserverlib) and uses the [afero](https://github.com/spf13/afero) filesystem abstraction.

### Key Components

#### main.go:25
Entry point that handles configuration loading, server instantiation, and graceful shutdown. Supports SIGTERM and SIGHUP (config reload) signals.

#### config/ package
- `config/config.go` - Configuration management and user authentication
- `config/confpar/` - Configuration parameter definitions
- Uses JSON configuration files (default: `ftpserver.json`)

#### fs/ package - Filesystem Backends
Each subdirectory implements a specific storage backend:
- `fs/afos/` - Local OS filesystem (afero wrapper)
- `fs/s3/` - Amazon S3 storage
- `fs/gdrive/` - Google Drive integration
- `fs/dropbox/` - Dropbox storage
- `fs/gcs/` - Google Cloud Storage
- `fs/sftp/` - SFTP backend
- `fs/mail/` - Email storage backend
- `fs/telegram/` - Telegram bot storage
- `fs/keycloak/` - Keycloak authentication integration
- `fs/utils/` - Shared filesystem utilities
- `fs/stripprefix/` - Path manipulation utilities
- `fs/fslog/` - Filesystem logging wrapper

#### server/ package
Contains the main FTP server driver implementation that bridges ftpserverlib with the various filesystem backends.

### Configuration System
- JSON-based configuration with schema validation
- Supports multiple user accounts with different backends
- Each user can have different filesystem backends (S3, Dropbox, etc.)
- TLS/SSL support for secure connections
- Configurable passive transfer port ranges

### Backend Integration Pattern
Each filesystem backend in `fs/` follows a consistent pattern:
1. Implements afero.Fs interface
2. Handles backend-specific authentication/configuration
3. Maps FTP operations to backend API calls
4. Provides error handling and logging

### Build Information
- `build.go` contains build-time variables (version, date, commit)
- Cross-platform builds supported via GitHub Actions
- Docker images available and built automatically

## Development Notes

### Testing Strategy
- Limited test coverage (only `fs/utils/env_test.go` currently)
- Integration testing likely done via Docker containers
- GitHub Actions runs tests with race detection

### Code Quality
- Comprehensive golangci-lint configuration with 40+ linters enabled
- Strict formatting and import organization rules
- Local package prefix: `github.com/fclairamb/ftpserver/`
- Line length limit: 120 characters
- Function length limits: 80 lines/40 statements

### Dependencies
- Go 1.24+ required (toolchain 1.25+)
- Key dependencies: ftpserverlib, afero, various cloud SDK packages
- Logging via go-kit/log framework
- Authentication via go-crypt for password hashing