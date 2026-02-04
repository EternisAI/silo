# Silo CLI - Developer Documentation

CLI tool to deploy and manage **Silo Box** on customer hardware.

Silo Box is a self-hosted AI chat application. This CLI installs and orchestrates the Docker containers (PostgreSQL, backend, frontend, llama.cpp inference, proxy agent). The actual application code lives in `../silo_box/`.

## Quick Start

```bash
make build        # Build binary to bin/silo
make test         # Run tests
make fmt          # Format code
make lint         # Run golangci-lint
make dev ARGS="up"  # Run without building (go run)
```

## Architecture Overview

Two binaries: **silo** (CLI) and **silod** (daemon). CLI orchestrates Docker Compose operations. Daemon provides HTTP API for remote management. Both share config/state from `~/.config/silo/` and `~/.local/share/silo/`.

```
CLI (cmd/silo)                     Daemon (cmd/silod)
┌──────────────┐                   ┌──────────────┐
│ Cobra Router │                   │ HTTP Server  │
│   Commands   │                   │  (port 9999) │
└──────┬───────┘                   └──────┬───────┘
       │                                  │
       ├─ up/down/status/logs            ├─ /health
       ├─ upgrade/check/version          ├─ /status
       ├─ inference up/down/status       ├─ /api/v1/up, /api/v1/down, /api/v1/restart
       │                                  ├─ /api/v1/upgrade, /api/v1/logs, /api/v1/version
       └─ Installer/Docker/Updater       ├─ /api/v1/check
                                          ├─ /api/v1/inference/up
                                          ├─ /api/v1/inference/down
                                          ├─ /api/v1/inference/status
                                          └─ /api/v1/inference/logs
```

## Project Structure

```
silo/
├── cmd/
│   ├── silo/main.go         # CLI entry point
│   └── silod/main.go        # Daemon entry point
│
├── internal/
│   ├── assets/
│   │   └── templates.go     # Embedded docker-compose.yml.tmpl, config.yml.tmpl
│   │
│   ├── cli/
│   │   ├── root.go          # Cobra root command + global flags
│   │   ├── up.go            # Install or start services (--all includes inference)
│   │   ├── down.go          # Stop services (--all includes inference)
│   │   ├── status.go        # Show service status (includes inference engine)
│   │   ├── logs.go          # View container logs
│   │   ├── upgrade.go       # Pull latest images and recreate
│   │   ├── check.go         # Validate config + installation
│   │   ├── version.go       # Show versions + check for updates
│   │   └── inference.go     # Inference engine commands (up/down/status/logs)
│   │
│   ├── inference/
│   │   └── inference.go     # SGLang inference engine management (docker run)
│   │
│   ├── daemon/
│   │   ├── daemon.go        # Daemon lifecycle and configuration
│   │   ├── server.go        # HTTP API server
│   │   └── handlers.go      # API endpoint handlers
│   │
│   ├── config/
│   │   ├── manager.go       # Config/State structs, Load/Save, validation
│   │   ├── paths.go         # Directory path management
│   │   └── templates.go     # Template rendering for docker-compose/config
│   │
│   ├── docker/
│   │   ├── compose.go       # Docker Compose operations wrapper
│   │   └── check.go         # Docker environment validation
│   │
│   ├── installer/
│   │   ├── installer.go     # Multi-step installation orchestration
│   │   └── preflight.go     # Pre-installation checks (disk, port, docker)
│   │
│   ├── updater/
│   │   └── updater.go       # Upgrade flow (backup → pull → recreate)
│   │
│   └── version/
│       └── version.go       # Version checking (GitHub + Docker Hub APIs)
│
├── pkg/
│   └── logger/
│       └── logger.go        # Colored structured logging (Info, Success, Warn, Error)
│
└── scripts/
    ├── install.sh           # Remote installation script (CLI + daemon)
    ├── install-service.sh   # Systemd service installer
    ├── silod.service        # Systemd service template
    └── build.sh             # Build automation
```

## Key Components

### 1. CLI Commands (`internal/cli/`)

| Command | File | Description | Key Logic |
|---------|------|-------------|-----------|
| `silo up` | up.go | Install or start | Check state.json → installer if new, else docker up |
| `silo up --all` | up.go | Include inference engine | Also starts SGLang inference container |
| `silo down` | down.go | Stop services | Docker compose down (preserve volumes), excludes inference |
| `silo down --all` | down.go | Include inference engine | Also stops SGLang inference container |
| `silo status` | status.go | Service status | Parse docker ps + inference engine status |
| `silo logs` | logs.go | Container logs | Docker compose logs with -f/--tail |
| `silo upgrade` | upgrade.go | Update images | Backup → pull → down → up (never touches inference) |
| `silo check` | check.go | Validate config | YAML validation + file existence checks |
| `silo version` | version.go | Show versions | Local version + GitHub/Docker Hub API calls |
| `silo inference up` | inference.go | Start inference | docker run sglang container |
| `silo inference down` | inference.go | Stop inference | docker stop/rm sglang container |
| `silo inference status` | inference.go | Inference status | Container info + health check |
| `silo inference logs` | inference.go | Inference logs | docker logs with -f/--tail |

### 2. Configuration Management (`internal/config/`)

**Config Struct** (`manager.go`):
```go
type Config struct {
    Version               string // CLI version
    ImageTag              string // Docker image tag
    Port                  int    // Frontend port
    LLMBaseURL            string // Inference engine URL
    DefaultModel          string // Default LLM model
    InferencePort         int
    InferenceModelFile    string // GGUF filename
    InferenceShmSize      string
    InferenceContextSize  int
    InferenceBatchSize    int
    InferenceGPULayers    int    // 999 = all layers on GPU
    InferenceTensorSplit  string
    InferenceMainGPU      int
    InferenceThreads      int
    InferenceHTTPThreads  int
    InferenceFit          string
    InferenceGPUDevices   string // Quoted CSV: "0", "1", "2"
}
```

**State Struct**:
```go
type State struct {
    Version     string    // CLI version
    InstalledAt time.Time
    LastUpdated time.Time
}
```

**Path Management** (`paths.go`):
- Config dir: `~/.config/silo/` (contains config.yml)
- Data dir: `~/.local/share/silo/` (contains docker-compose.yml, state.json, data/)
- Environment variables: `SILO_CONFIG_DIR`, `SILO_DATA_DIR` (daemon only)
- Auto-creates directories with 0755 permissions

### 3. Docker Integration (`internal/docker/`)

**Compose Operations** (`compose.go`):
```go
type Compose struct {
    ComposeFile string
    ProjectName string
}

// Auto-detects docker compose v2 vs v1
func (c *Compose) Up() error
func (c *Compose) Down() error
func (c *Compose) Pull(services []string) error
func (c *Compose) Ps() ([]byte, error)
func (c *Compose) Logs(follow bool, tail int, service string) error
func (c *Compose) Restart(services []string) error
func (c *Compose) IsRunning() (bool, error)
```

**Environment Checks** (`check.go`):
- Docker installed (checks `docker` command)
- Docker daemon running (checks `docker info`)
- Compose available (checks `docker compose` or `docker-compose`)

### 4. Installation (`internal/installer/`)

**Installation Steps** (`installer.go`):
1. **Preflight Checks**: Docker, disk space (≥5GB), port availability
2. **Create Directories**: Config and data directories
3. **Generate Configs**: Render templates → save config.yml, docker-compose.yml
4. **Pull Images**: Only backend and frontend (inference pre-packaged)
5. **Start Containers**: Docker compose up
6. **Save State**: Write state.json with install timestamp

**Preflight Checks** (`preflight.go`):
- System checks (Docker installed/running)
- Disk space check (minimum 5GB free)
- Port availability check (default 80, or user-specified)

### 5. Upgrades (`internal/updater/`)

**Upgrade Flow** (`updater.go`):
1. Check latest versions (GitHub releases, Docker Hub tags)
2. Backup config.yml → config.yml.bak
3. Update image_tag in config if newer available
4. Regenerate docker-compose.yml from template
5. Pull new images (frontend, backend only)
6. Down existing containers (preserve volumes)
7. Up new containers
8. Update state.json with last_updated timestamp

### 6. Version Management (`internal/version/`)

**Version Checking** (`version.go`):
- CLI updates: GitHub Releases API (`GET /repos/EternisAI/silo/releases/latest`)
- Image updates: Docker Hub API (`GET /v2/repositories/eternis/silo-box/tags`)
- Semantic version comparison (e.g., "0.1.3" > "0.1.2")

## Configuration Files

### config.yml Schema

```yaml
version: "0.1.2"                       # CLI version
image_tag: "0.1.2"                     # Docker image tag
port: 80                               # Frontend port
llm_base_url: "http://inference-engine:30000/v1"
default_model: "GLM-4.7-Q4_K_M.gguf"
inference_port: 30000
inference_model_file: "GLM-4.7-Q4_K_M.gguf"
inference_shm_size: "16g"
inference_context_size: 8192
inference_batch_size: 256
inference_gpu_layers: 999              # 999 = all layers on GPU
inference_tensor_split: "1,1,1"
inference_main_gpu: 0
inference_threads: 16
inference_http_threads: 8
inference_fit: "off"
inference_gpu_devices: "\"0\", \"1\", \"2\""  # Quoted CSV for YAML

# Service toggles
enable_inference_engine: false         # Enable llama.cpp inference
enable_proxy_agent: false              # Enable remote proxy agent
enable_deep_research: true             # Enable deep research service

# Deep research configuration
deep_research_image: "ghcr.io/eternisai/deep_research:sha-ff37ec2"
deep_research_port: 3031
search_provider: "perplexity"
perplexity_api_key: ""                 # Required for deep research web search
```

### state.json Schema

```json
{
  "version": "0.1.2",
  "installed_at": "2025-01-30T12:00:00Z",
  "last_updated": "2025-01-30T14:30:00Z"
}
```

## Design Patterns

1. **Template-driven Configuration**: docker-compose.yml and config.yml generated from embedded templates with user values
2. **Single-responsibility Packages**: Each package handles one concern (installer, updater, docker, config)
3. **Stateful Operations**: Tracks install timestamps and versions in state.json
4. **Selective Image Pulls**: Only pulls backend/frontend (inference engine image is larger, pre-packaged)
5. **Non-blocking Updates**: Version checks warn but don't fail operations
6. **Graceful Degradation**: Warns on errors but continues where possible

## Development Workflow

```bash
# Build
make build                             # Build CLI to bin/silo
make build-daemon                      # Build daemon to bin/silod
make build-all                         # Build both

# Install locally
make install-daemon                    # Install silod to ~/.local/bin
make install-service                   # Install systemd service (templates with your user)

# Local development (no build)
make dev ARGS="up"                     # Run CLI
make dev-daemon                        # Run daemon

# Testing
make test                              # Run tests
make fmt                               # Format code
make lint                              # Run golangci-lint
```

## Testing Strategy

**Unit Tests**:
- Config validation logic
- Path management functions
- Version comparison logic
- Template rendering

**Integration Tests** (require Docker):
- Full installation flow
- Upgrade flow
- Docker compose operations

**Manual Testing Checklist**:
- [ ] Fresh install (`silo up`)
- [ ] Stop/start cycle (`silo down` → `silo up`)
- [ ] Upgrade flow (`silo upgrade`)
- [ ] Config validation (`silo check`)
- [ ] Log viewing (`silo logs`)
- [ ] Version checking (`silo version`)
- [ ] Daemon service installation and startup

## Dependencies

- **spf13/cobra**: CLI framework
- **gopkg.in/yaml.v3**: YAML parsing
- **fatih/color**: Terminal colors
- **Go 1.25.0+**: Required Go version

## Installation & Deployment

### Remote Installation (End Users)

```bash
# Install CLI + daemon
curl -fsSL https://raw.githubusercontent.com/EternisAI/silo/main/scripts/install.sh | bash

# Install systemd service (optional)
curl -fsSL https://raw.githubusercontent.com/EternisAI/silo/main/scripts/install-service.sh | bash
sudo systemctl enable silod
sudo systemctl start silod
```

**Installation Details:**
- Binaries installed to: `~/.local/bin/silo` and `~/.local/bin/silod`
- Config directory: `~/.config/silo/`
- Data directory: `~/.local/share/silo/`
- Service runs as installing user (not root)
- Service file auto-configured during installation

### Daemon (silod)

**HTTP API** (port 9999):
- `/health` - Health check
- `/status` - Detailed daemon status
- `/api/v1/up`, `/api/v1/down`, `/api/v1/restart` - Container control
- `/api/v1/upgrade`, `/api/v1/logs`, `/api/v1/version`, `/api/v1/check` - Operations
- `/api/v1/inference/up`, `/api/v1/inference/down` - Inference engine control
- `/api/v1/inference/status`, `/api/v1/inference/logs` - Inference engine info

**Environment Variables:**
- `SILO_CONFIG_DIR` - Override config directory (default: `~/.config/silo`)
- `SILO_DATA_DIR` - Override data directory (default: `~/.local/share/silo`)

**Service Management:**
```bash
sudo systemctl status silod   # Check status
sudo journalctl -u silod -f   # View logs
sudo systemctl restart silod  # Restart
```

## Common Tasks

### Adding a New Command

1. Create `internal/cli/newcommand.go`
2. Define command with `cobra.Command`
3. Register in `internal/cli/root.go` (`rootCmd.AddCommand(newCmd)`)
4. Implement logic using existing packages (config, docker, etc.)

### Adding Configuration Fields

1. Add field to `Config` struct in `internal/config/manager.go`
2. Update `Validate()` method for validation rules
3. Update templates in `internal/assets/` to use new field
4. Update default values in `NewDefaultConfig()`

### Modifying Installation Flow

1. Edit `internal/installer/installer.go`
2. Keep steps sequential and idempotent
3. Add preflight checks to `internal/installer/preflight.go` if needed
4. Update state tracking if adding new metadata

## Releasing

Releases are automated via GitHub Actions. **Do not manually create tags.**

### Creating a Release

1. Push changes to `main` (via PR or direct push)
2. Go to **Actions** → **Release** → **Run workflow**
3. Choose version bump type:
   - `patch` (default): Bug fixes, minor updates (0.1.2 → 0.1.3)
   - `minor`: New features, backward compatible (0.1.2 → 0.2.0)
   - `major`: Breaking changes (0.1.2 → 1.0.0)
4. Click **Run workflow**

### What Happens

1. **Tag created** - Auto-increments version based on bump type
2. **GoReleaser runs** - Builds binaries for Linux/macOS (amd64/arm64)
3. **GitHub Release created** - With binaries, checksums, and changelog
4. **Install script updated** - Users running `curl | bash` get the new version

### CI Pipeline

PRs to `main` trigger CI: build, test, lint.

## Related Projects

- **silo_box/** — The application this CLI deploys (Go backend + Next.js frontend)
- Backend: REST API, database management, model inference
- Frontend: Next.js chat UI
- Inference: llama.cpp server with GGUF models
