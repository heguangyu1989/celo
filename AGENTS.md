# AGENTS.md - Celo CLI Tool

## Project Overview

Celo is a developer efficiency command-line tool written in Go, designed to provide essential utilities for daily development workflows. The tool focuses on speed and efficiency, offering features like MD5 checksum calculation, GitLab merge request creation, and environment file management.

**Key Features:**
- MD5 checksum calculation for files and strings with multiple output formats (JSON, YAML, table)
- GitLab merge request creation from command line
- Interactive .env file switching using terminal UI
- Build information display
- Secure random password generation with customizable options

The project uses a clean modular architecture with clear separation between command implementations, public packages, and internal logic.

## Technology Stack

- **Language**: Go 1.25.1
- **CLI Framework**: Cobra (github.com/spf13/cobra)
- **Terminal UI**: Charm Bracelet ecosystem
  - bubbletea: TUI framework
  - bubbles: UI components
  - lipgloss: Styled terminal rendering
- **HTTP Client**: Resty v2 (github.com/go-resty/resty/v2)
- **Logging**: Logrus (github.com/sirupsen/logrus)
- **Testing**: Testify (github.com/stretchr/testify)
- **Configuration**: YAML/JSON support via gopkg.in/yaml.v3

## Project Structure

```
.
├── main.go                 # Entry point, delegates to cmd package
├── cmd/                    # Command implementations
│   ├── root.go            # Main cobra command setup
│   ├── md5.go             # MD5 calculation command
│   ├── merge.go           # GitLab merge request command
│   ├── build_info.go      # Build information command
│   ├── env.go             # Interactive .env switching (TUI)
│   ├── password.go        # Password generation command
│   └── conf_cmd.go        # Config generation command
├── pkg/                    # Public packages
│   ├── config/            # Configuration management
│   ├── utils/             # File and math utilities
│   │   ├── files.go
│   │   ├── math.go
│   │   └── password.go    # Password generation logic
│   └── p/                 # Pretty printing with styling
├── internal/              # Internal packages
│   └── merge/             # GitLab integration logic
├── go.mod                 # Go module definition
├── Makefile              # Build automation
└── quick_start.md        # User documentation (Chinese)
```

## Build and Development Commands

### Building
```bash
# Build the binary
go build -o celo .

# Install to system
sudo mv celo /usr/local/bin/
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test -v ./pkg/utils
go test -v ./pkg/p
```

### Linting
```bash
# Run golangci-lint (requires installation)
make lint
# or
golangci-lint run
```

### Configuration
```bash
# Generate default config file
celo gen-default --dst celo.yaml

# Use custom config file
celo --config /path/to/config.yaml md5 file.txt
```

## Code Style Guidelines

### Language and Comments
- **Code**: English for all identifiers, function names, and documentation
- **Comments**: Mix of English and Chinese (tests primarily use Chinese comments)
- **User-facing output**: Mix of English and Chinese (GitLab features use English, env switching uses Chinese)

### Package Organization
- **cmd/**: Command implementations, each command in its own file
- **pkg/**: Public API packages, reusable across projects
- **internal/**: Project-specific internal logic not exposed externally

### Error Handling
- Always return errors from command RunE functions
- Use structured logging with Logrus for fatal errors
- User-facing errors use the `p` package for styled output

### Dependencies
- Minimal external dependencies
- Preference for well-maintained libraries (Charm Bracelet, Cobra)
- No vendor directory (uses Go modules)

## Testing Strategy

### Test Structure
- Test files follow Go conventions: `*_test.go`
- Tests exist for utility functions in `pkg/` packages
- Testify used for assertions (`assert` package)

### Running Tests
```bash
# Test utilities
go test ./pkg/utils -v

# Test printing functions
go test ./pkg/p -v
```

### Test Coverage Areas
- **File operations**: Existence checks, .env file discovery
- **Math utilities**: MaxInt function edge cases
- **Printing functions**: Styled output verification

## Key Commands Explained

### MD5 Command (`cmd/md5.go`)
- Calculates MD5 checksums for files or strings
- Auto-detects input type (file vs string)
- Supports JSON, YAML, and table output formats
- Uses Charm Bracelet table for formatted display

### Merge Command (`cmd/merge.go`, `internal/merge/`)
- Creates GitLab merge requests via API
- Automatically detects GitLab project from git remote
- Requires `gitlab_token` in config
- Uses Resty for HTTP requests

### Env Command (`cmd/env.go`)
- Interactive TUI for switching .env files
- Creates symbolic links from selected file to `.env`
- Uses bubbletea for terminal UI
- Lists all `.env.*` files (excluding `.env` itself)

### Password Command (`cmd/password.go`)
- Generates secure random passwords with customizable options
- Supports multiple output formats (JSON, YAML, table)
- Configurable character sets (uppercase, lowercase, digits, special characters)
- Custom character set support
- Batch password generation capability
- Uses crypto/rand for cryptographically secure random generation

### Config Command (`cmd/conf_cmd.go`)
- Generates default configuration file
- Supports both YAML and JSON formats
- Default location: `~/.celo.yaml`

## Configuration System

### Config Structure
```yaml
gitlab_token: "your_gitlab_personal_access_token"
```

### Config Loading Priority
1. Command-line flag: `--config /path/to/config.yaml`
2. Default locations: `~/.celo.yaml` or `~/.celo.json`
3. In-memory default config if no file exists

### Config Format Support
- YAML (`.yaml`, `.yml`)
- JSON (`.json`)
- Auto-detection based on file extension

## Security Considerations

### Sensitive Data
- GitLab tokens stored in config file
- Config file should have restricted permissions (0600 recommended)
- Never commit config files with tokens to version control

### GitLab Integration
- Uses Personal Access Tokens with `api` scope
- HTTPS only (no SSH for API calls)
- Token passed as `private_token` parameter

### File Operations
- Symlink creation for .env files (potential security risk if directory is writable by others)
- File existence checks before reading

## Development Workflow

### Adding New Commands
1. Create new file in `cmd/` directory
2. Implement command following Cobra patterns
3. Add command to root command in `cmd/root.go`
4. Add tests if command contains business logic
5. Update quick_start.md with usage examples

### Adding New Packages
- Public APIs: Place in `pkg/` directory
- Internal logic: Place in `internal/` directory
- Follow existing package structure and naming conventions

### Testing Changes
- Run full test suite: `go test ./...`
- Test individual commands manually
- Verify configuration loading works
- Check TUI commands with different terminal sizes

## IDE and Tooling

### IDE Support
- **GoLand/IntelliJ**: Project includes `.idea/` directory with configuration
- **VS Code**: Not explicitly configured, but supported via Go extension

### Required Tools
- Go 1.25.1 or higher
- golangci-lint (for linting)
- Git (for GitLab functionality)

### Git Integration
- Automatic GitLab project detection from `remote.origin.url`
- Supports both HTTPS and SSH git URLs
- Parses GitLab paths for API calls