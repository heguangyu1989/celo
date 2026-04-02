# Celo - AI Agent Guide

## Project Overview

Celo is a Go-based CLI tool designed for developer productivity. The name represents "Efficiency, at speed." It provides a collection of practical utilities for daily development workflows including:

- **MD5 checksum calculation** for files and strings
- **Secure password generation** with customizable character sets
- **GitLab merge request creation** from command line
- **Environment file switching** with interactive TUI
- **Network port diagnostics** with process identification
- **Docker image verification** in registries
- **VSCode Server management** for remote development cleanup

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25.1+ |
| CLI Framework | [spf13/cobra](https://github.com/spf13/cobra) |
| TUI Components | [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea), [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles), [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) |
| HTTP Client | [go-resty/resty](https://github.com/go-resty/resty) |
| Logging | [sirupsen/logrus](https://github.com/sirupsen/logrus) |
| YAML/JSON | [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) |
| Testing | [stretchr/testify](https://github.com/stretchr/testify) |

## Project Structure

```
.
├── main.go              # Application entry point
├── go.mod               # Go module definition
├── go.sum               # Go dependency checksums
├── Makefile             # Build automation (lint only)
├── README.md            # Project readme
├── .gitignore           # Git ignore rules
├── AGENTS.md            # This file
├── cmd/                 # Command implementations
│   ├── root.go          # Root command & command registration
│   ├── build_info.go    # 'info' command
│   ├── md5.go           # 'md5' command
│   ├── password.go      # 'password' command
│   ├── merge.go         # 'merge' command (GitLab MR)
│   ├── env.go           # 'env' command (interactive env switcher)
│   ├── net.go           # 'net' command (port checking)
│   ├── docker.go        # 'docker' command (image verification)
│   ├── vc.go            # 'vc' command (VSCode Server management)
│   └── conf_cmd.go      # 'gen-default' command
├── internal/            # Internal packages
│   └── merge/           # GitLab merge request logic
│       ├── merge.go     # MR creation API call
│       ├── git_path.go  # Git URL parsing
│       ├── read_config.go # Git config reading
│       ├── types.go     # Domain types
│       └── gitlab_types.go # GitLab API types
└── pkg/                 # Public packages
    ├── config/          # Configuration management
    │   ├── config.go    # Config struct and load/save
    │   └── misc.go      # Default path utilities
    ├── utils/           # General utilities
    │   ├── files.go     # File operations
    │   ├── math.go      # Math helpers (MaxInt, MinInt)
    │   └── password.go  # Password generation
    └── p/               # Printing utilities
        └── print.go     # Styled console output
```

## Build and Run Commands

### Build from Source

```bash
# Build binary
go build -o celo .

# Install to system
sudo mv celo /usr/local/bin/
```

### Development Commands

```bash
# Run linter
make lint

# Or directly
golangci-lint run

# Run tests
go test ./...

# Run specific package tests
go test ./pkg/utils/...
```

## Available Commands

| Command | Description | Key Flags |
|---------|-------------|-----------|
| `celo info` | Show build information | - |
| `celo md5 <file/string...>` | Calculate MD5 checksums | `--output` (json/yaml/table) |
| `celo password` | Generate random passwords | `--length`, `--upper`, `--lower`, `--digits`, `--special`, `--custom`, `--count`, `--output` |
| `celo merge` | Create GitLab merge request | `--src`, `--dst`, `--title`, `--tags` |
| `celo env [path]` | Interactive environment file switcher | - |
| `celo net port [port/range...]` | Check TCP port status | `--output`, `--timeout` |
| `celo docker check [images...]` | Verify Docker images exist | `--output` |
| `celo vc skill-all` | Kill all VSCode Server processes | - |
| `celo vc clean` | Clean VSCode Server folders | `--keep`, `--yes` |
| `celo gen-default` | Generate default config file | `--dst` |

## Configuration

Configuration file location: `~/.celo.yaml` or `~/.celo.json`

### Example Configuration

```yaml
# ~/.celo.yaml
gitlab_token: "your_gitlab_personal_access_token"
```

The config file is loaded automatically on startup. Use `--config` flag to specify a custom path.

## Code Style Guidelines

### Package Organization

1. **cmd/**: Each command is in its own file. Most commands are registered in `cmd/root.go` via `init()` function, but some commands (e.g., `password`, `vc`) register themselves via `init()` in their own files.
2. **internal/**: Private implementation details. The `merge` package handles GitLab API interactions.
3. **pkg/**: Public reusable packages:
   - `config`: Configuration management with YAML/JSON support
   - `utils`: General-purpose utilities (file ops, math, password generation)
   - `p`: Styled printing utilities using lipgloss

### Naming Conventions

- Command functions: `GetXXXCmd()` returns `*cobra.Command`
- Command runners: `runXXXCmd(cmd *cobra.Command, args []string) error`
- Types use PascalCase, constants use CamelCase or ALL_CAPS
- Test files: `xxx_test.go` with `TestXXX` function names

### Error Handling

- Return errors from command runners for centralized handling
- Use `p.Error()` for styled error output
- Log fatal errors only in `main.go`

### Output Formats

Commands support multiple output formats (JSON, YAML, Table). Follow this pattern:

```go
switch output {
case "json":
    return p.PrintJSON(results)
case "yaml":
    return p.PrintYAML(results)
case "table":
    printTable(results)
}
```

## Testing

Tests are located alongside source files with `_test.go` suffix.

### Running Tests

```bash
# All tests
go test ./...

# Verbose output
go test -v ./...

# Specific package
go test ./pkg/utils/...

# With coverage
go test -cover ./...
```

### Test Patterns

- Use `t.TempDir()` for temporary files in tests
- Use `testify/assert` for assertions
- Use table-driven tests for multiple test cases
- Test both success and error cases

Example from `pkg/utils/password_test.go`:

```go
func TestGeneratePassword_Length(t *testing.T) {
    tests := []struct {
        name   string
        length int
    }{
        {"length 8", 8},
        {"length 12", 12},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Key Implementation Details

### Interactive TUI (env command)

Uses Charmbracelet's Bubble Tea framework for interactive list selection. See `cmd/env.go` for the model implementation pattern.

### GitLab Integration

The `merge` command:
1. Reads Git config to get remote URL
2. Parses GitLab URL (supports both HTTPS and SSH formats)
3. Creates MR via GitLab API v4
4. Requires `gitlab_token` in config file

### Platform-Specific Code

The `net port` command includes platform-specific process detection:
- macOS: `lsof` and `netstat`
- Linux: `ss`, `netstat`, `lsof`
- Windows: `netstat` and `tasklist`

## Security Considerations

1. **GitLab Token**: Stored in config file with user permissions. Never commit config files.
2. **Password Generation**: Uses `crypto/rand` for cryptographically secure random numbers, not `math/rand`.
3. **File Operations**: Check file existence and permissions before operations.
4. **Command Execution**: Validate inputs before passing to shell commands (in `vc` and `net` commands).

## Development Workflow

1. Add new commands in `cmd/xxx.go` with `GetXXXCmd()` function
2. Register in `cmd/root.go` `init()` function (or use `init()` in the command file)
3. Add tests in `cmd/xxx_test.go` or `pkg/xxx/xxx_test.go`
4. Run linter: `make lint`
5. Run tests: `go test ./...`
