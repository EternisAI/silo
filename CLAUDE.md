# Silo CLI

CLI tool to deploy and manage **Silo Box** on customer hardware.

Silo Box is a self-hosted AI chat application. This CLI installs and orchestrates Docker containers: PostgreSQL, backend, frontend, llama.cpp inference engine, and proxy agent. The actual application code lives in `../silo_box/`.

## Architecture

This is a Cobra-based CLI that wraps Docker Compose operations. The CLI is stateful - it tracks installation state in `~/.local/share/silo/state.json` and configuration in `~/.config/silo/config.yml`.

### Key Design Patterns

1. **Template-driven Configuration**: Docker-compose and config files generated from embedded templates (`internal/assets/`)
2. **Single-responsibility Packages**: Each package handles one concern (installer, updater, docker, config)
3. **Stateful Operations**: Tracks install timestamps, versions, and config in state file
4. **Selective Image Pulls**: Only pulls backend/frontend images (inference engine pre-packaged)
5. **Non-blocking Upgrades**: Version checks warn but don't fail operations

## Package Structure

```
internal/
├── cli/            # Cobra command implementations (up, down, status, logs, upgrade, check, version)
├── daemon/         # HTTP API server for remote management
├── config/         # Config/State loading, saving, validation, path management
├── docker/         # Docker Compose wrapper (up, down, pull, ps, logs, restart)
├── installer/      # Multi-step installation (preflight → dirs → configs → images → containers)
├── updater/        # Upgrade flow (backup → pull → recreate → update state)
├── version/        # GitHub/Docker Hub API version checks
└── assets/         # Embedded templates (docker-compose.yml.tmpl, config.yml.tmpl)

pkg/
└── logger/         # Colored structured logging

cmd/
├── silo/           # CLI entry point
└── silod/          # Daemon entry point
```

## Make Commands

```bash
make build         # Build binary to bin/silo
make test          # Run tests
make fmt           # Format code
make lint          # Run golangci-lint
make dev ARGS="up" # Run without building (go run)
```

## CLI Commands

| Command | Description | Key Logic |
|---------|-------------|-----------|
| `silo up` | Install (first run) or start services | Checks state.json, runs installer if missing, else docker up |
| `silo down` | Stop services, preserve data | Docker down (no volumes removed) |
| `silo status` | Show service status | Docker ps + parse container states |
| `silo logs` | View container logs | Docker compose logs with follow/tail |
| `silo upgrade` | Pull latest images and restart | Backup config → pull → down → up → update state |
| `silo check` | Validate config and installation | Validate YAML schema + check file existence |
| `silo version` | Show versions + check for updates | Query GitHub releases + Docker Hub tags |

## File Paths

```
~/.config/silo/config.yml              # User configuration (YAML)
~/.local/share/silo/state.json         # Installation state (installed_at, last_updated)
~/.local/share/silo/docker-compose.yml # Generated from template
~/.local/share/silo/data/models/       # LLM model files
~/.local/share/silo/data/postgres/     # PostgreSQL data
```

## Configuration Schema

See `internal/config/manager.go` for the full `Config` struct. Key fields:
- `ImageTag`: Docker image version
- `Port`: Frontend port
- `InferenceModelFile`: GGUF model filename
- `InferenceGPULayers`: GPU layers (999 = all)
- `InferenceContextSize`: Token context window
- `InferenceGPUDevices`: GPU device IDs (quoted CSV string)

## Installation Flow

1. **Preflight** (`installer/preflight.go`): Check Docker installed/running, port available, disk space ≥5GB
2. **Create Directories**: `~/.config/silo` and `~/.local/share/silo/data`
3. **Generate Configs**: Render templates with user flags → save config.yml and docker-compose.yml
4. **Pull Images**: Only backend and frontend (inference pre-packaged)
5. **Start Containers**: Docker compose up
6. **Save State**: Write state.json with install timestamp

## Upgrade Flow

1. **Version Check** (`version/version.go`): Query GitHub releases + Docker Hub for latest versions
2. **Backup**: Copy config.yml to config.yml.bak
3. **Update Config**: Set new image_tag in config
4. **Regenerate Compose**: Re-render docker-compose.yml template
5. **Pull Images**: Docker compose pull (backend, frontend only)
6. **Recreate Containers**: Down existing → Up new
7. **Update State**: Write last_updated timestamp

## Docker Integration

- Auto-detects `docker compose` (v2) vs `docker-compose` (v1)
- Uses `restart: unless-stopped` for auto-restart on reboot
- Preserves named volumes during down/up cycles
- Selective pulling: frontend, backend (skips inference-engine)

## Testing Notes

- Tests use `testify/assert` and `testify/require`
- Mock Docker operations by testing config/path logic independently
- Integration tests would require Docker daemon

## Common Code Patterns

```go
// Load config + state
cfg, err := config.LoadConfig(configDir)
state, err := config.LoadState(stateDir)

// Check if installed
if !state.IsInstalled() {
    // run installer
}

// Docker operations
dc := docker.NewCompose(composeFile)
dc.Up()
dc.Down()
dc.Pull([]string{"frontend", "backend"})

// Logging
logger.Info("Starting services...")
logger.Success("Services started")
logger.Error("Failed: %v", err)
```

## Releasing

Releases are automated via GitHub Actions. **Do not manually create tags.**

```bash
# Push changes to main, then trigger release
gh workflow run Release
```

The workflow auto-increments the version, builds binaries via GoReleaser, and creates a GitHub Release.

## Related

- **silo_box/** — The application this CLI deploys (Go backend + Next.js frontend)
