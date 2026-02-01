# Silo Architecture

## Overview

Silo consists of two complementary binaries:

1. **`silo`** - CLI tool for on-demand container management
2. **`silod`** - Background daemon for remote management API

Both share internal packages while serving different purposes.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          User Layer                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────┐                        ┌─────────────┐         │
│  │   silo CLI  │                        │  silod      │         │
│  │             │                        │  (daemon)   │         │
│  │  - up       │                        │             │         │
│  │  - down     │                        │  - HTTP API │         │
│  │  - status   │                        │             │         │
│  │  - logs     │                        │             │         │
│  │  - upgrade  │                        │             │         │
│  │  - check    │◄──────── API ─────────►│  :9999      │         │
│  │  - version  │      (optional)        │             │         │
│  └──────┬──────┘                        └──────┬──────┘         │
│         │                                      │                 │
└─────────┼──────────────────────────────────────┼─────────────────┘
          │                                      │
          │                                      │
┌─────────▼──────────────────────────────────────▼─────────────────┐
│                     Shared Internal Packages                      │
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │  config  │  │  docker  │  │ version  │  │  logger  │        │
│  │          │  │          │  │          │  │          │        │
│  │ - Load   │  │ - Up     │  │ - Check  │  │ - Info   │        │
│  │ - Save   │  │ - Down   │  │ - Get    │  │ - Error  │        │
│  │ - State  │  │ - Ps     │  │ - Latest │  │ - Warn   │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
│                                                                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                      │
│  │installer │  │ updater  │  │  assets  │                      │
│  │          │  │          │  │          │                      │
│  │ - Run    │  │ - Update │  │ - tmpls  │                      │
│  └──────────┘  └──────────┘  └──────────┘                      │
│                                                                   │
└───────────────────────────────┬───────────────────────────────────┘
                                │
┌───────────────────────────────▼───────────────────────────────────┐
│                      Docker Compose Layer                         │
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  docker-compose.yml                                              │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌───────────┐ │
│  │  postgres  │  │  backend   │  │  frontend  │  │ inference │ │
│  │  :5432     │  │  :8080     │  │  :80       │  │ :30000    │ │
│  └────────────┘  └────────────┘  └────────────┘  └───────────┘ │
│                                                                   │
└───────────────────────────────────────────────────────────────────┘
```

## Component Breakdown

### CLI (`cmd/silo/`)

**Purpose**: On-demand user operations

**Commands**:
- `up` - Install or start containers
- `down` - Stop containers
- `status` - Show container status
- `logs` - View logs
- `upgrade` - Update images and restart
- `check` - Validate configuration
- `version` - Show versions and check for updates

**Lifecycle**: Short-lived processes (seconds to minutes)

**Use cases**:
- Initial installation
- Manual start/stop
- Interactive debugging
- Configuration changes
- Manual upgrades

### Daemon (`cmd/silod/`)

**Purpose**: Remote management API

**Components**:
- **Server** - HTTP API (port 9999)

**Lifecycle**: Long-running process (always on)

**Use cases**:
- Remote container management (up, down, restart)
- Remote upgrades
- Configuration validation via API
- Status API for external tooling

## Package Structure

```
silo/
├── cmd/
│   ├── silo/                 # CLI entry point
│   │   └── main.go
│   └── silod/                # Daemon entry point
│       └── main.go
│
├── internal/
│   ├── cli/                  # CLI command handlers (Cobra)
│   │   ├── root.go
│   │   ├── up.go
│   │   ├── down.go
│   │   ├── status.go
│   │   └── ...
│   │
│   ├── daemon/               # Daemon-specific logic
│   │   ├── daemon.go         # Main orchestration
│   │   ├── monitor.go        # Container monitoring
│   │   ├── scheduler.go      # Scheduled tasks
│   │   └── server.go         # HTTP API
│   │
│   ├── config/               # [Shared] Config/state I/O
│   │   ├── manager.go
│   │   ├── paths.go
│   │   └── state.go
│   │
│   ├── docker/               # [Shared] Docker Compose wrapper
│   │   ├── compose.go
│   │   └── check.go
│   │
│   ├── installer/            # [Shared] Installation flow
│   │   ├── installer.go
│   │   └── preflight.go
│   │
│   ├── updater/              # [Shared] Upgrade flow
│   │   └── updater.go
│   │
│   ├── version/              # [Shared] Version checking
│   │   └── version.go
│   │
│   └── assets/               # [Shared] Embedded templates
│       ├── config.yml.tmpl
│       └── docker-compose.yml.tmpl
│
└── pkg/
    └── logger/               # [Shared] Logging utility
        └── logger.go
```

## Data Flow

### Installation Flow (CLI)

```
User: silo up
    │
    ├─► Load Config
    ├─► Run Preflight Checks
    ├─► Create Directories
    ├─► Generate docker-compose.yml
    ├─► Pull Images (backend, frontend)
    ├─► docker compose up -d
    └─► Save State (installed_at)
```

### API Flow (Daemon)

```
silod start
    │
    ├─► Load Config & State
    │
    └─► Start API Server (goroutine)
        └─► HTTP endpoints on :9999
            ├─► /health
            ├─► /status
            └─► /api/v1/* (up, down, restart, etc.)
```

### Upgrade Flow (CLI)

```
User: silo upgrade
    │
    ├─► Backup config.yml
    ├─► Check latest versions
    ├─► Update config.yml (new image_tag)
    ├─► Regenerate docker-compose.yml
    ├─► docker compose pull
    ├─► docker compose down
    ├─► docker compose up -d
    └─► Update State (last_updated)
```

## Shared State

Both CLI and daemon access the same files:

### Configuration (`~/.config/silo/`)
```
config.yml          # User-editable settings
config.yml.backup   # Backup during upgrades
```

### State (`~/.local/share/silo/`)
```
state.json          # Installation metadata
docker-compose.yml  # Generated from template
data/
  ├── models/       # LLM models
  └── postgres/     # Database data
```

## Communication

### CLI ↔ Daemon (Optional)

The daemon exposes an HTTP API that the CLI *could* query:

```
┌──────┐          GET /status          ┌───────┐
│ CLI  ├───────────────────────────────►│Daemon │
│      │◄───────────────────────────────┤       │
└──────┘       Container status         └───────┘
```

Currently, both read the same files independently. Future enhancement: CLI queries daemon for real-time status instead of running `docker compose ps` itself.

## Concurrency

### Daemon Goroutines

```
main()
  └─► API Server (goroutine)
        └─► http.ListenAndServe

Server:
  - Shares context.Context
  - Cancel on SIGINT/SIGTERM
  - Use sync.WaitGroup for graceful shutdown
```

## File System Layout

```
~/.config/silo/
└── config.yml              # User configuration

~/.local/share/silo/
├── state.json              # Installation state
├── docker-compose.yml      # Generated compose file
└── data/
    ├── models/             # LLM models
    │   └── *.gguf
    └── postgres/           # PostgreSQL data
        └── pgdata/

/usr/local/bin/
├── silo                    # CLI binary
└── silod                   # Daemon binary

/etc/systemd/system/
└── silod.service           # Systemd service file
```

## Security Model

### Daemon Privileges

The daemon runs as a systemd service with:
- Access to Docker socket (requires root or docker group)
- Read/write to `~/.config/silo/` and `~/.local/share/silo/`
- Network access to GitHub/Docker Hub APIs
- HTTP server on localhost only

### API Exposure

The daemon API binds to `127.0.0.1:9999` (localhost only):
- Not exposed to external network
- No authentication required (local-only)
- Read-only status endpoints

## Extension Points

### Future CLI Enhancements
- Query daemon API for real-time status
- Trigger daemon actions via API
- Show daemon monitoring stats in `silo status`

### Future Daemon Enhancements
- Configurable monitoring intervals
- Alert notifications (email, Slack)
- Auto-upgrade mode (optional)
- Metrics collection (Prometheus)
- Web dashboard UI

## Build System

### Makefile Targets

```makefile
# CLI
make build          # Build silo binary
make install        # Install to /usr/local/bin

# Daemon
make build-daemon   # Build silod binary
make install-daemon # Install to /usr/local/bin
make install-service # Install systemd service

# Combined
make build-all      # Build both
make install-all    # Install everything

# Development
make dev            # Run CLI (go run)
make dev-daemon     # Run daemon (go run)
make test           # Run tests
```

### Build Script

`scripts/build.sh` supports both binaries:
- `MAIN=cmd/silo/main.go` for CLI
- `MAIN=cmd/silod/main.go` for daemon
- Injects version, commit, buildDate via LDFLAGS

## Testing Strategy

### Unit Tests
- Config loading/saving
- Docker command construction
- Version checking logic
- State management

### Integration Tests
- Require Docker daemon
- Test container lifecycle
- Verify compose operations

### Daemon Tests
- Mock Docker operations
- Test API handlers
- Verify server lifecycle

## Deployment

### Development
```bash
make dev              # Run CLI directly
make dev-daemon       # Run daemon directly
```

### Production
```bash
make install-all      # Install binaries + service
sudo systemctl enable silod
sudo systemctl start silod
```

## Summary

| Aspect | CLI | Daemon |
|--------|-----|--------|
| **Binary** | `silo` | `silod` |
| **Type** | Command-line tool | Background service |
| **Lifecycle** | Short-lived | Long-running |
| **Framework** | Cobra | Standard Go HTTP |
| **Purpose** | User operations | Remote management API |
| **Config** | Shared (`~/.config/silo/`) | Shared (`~/.config/silo/`) |
| **State** | Shared (`~/.local/share/silo/state.json`) | Shared |
| **Packages** | `internal/cli/`, shared | `internal/daemon/`, shared |
| **API** | None | HTTP REST (:9999) |
| **Install** | `make install` | `make install-service` |

## Design Philosophy

1. **Separation of Concerns**: CLI for user operations, daemon for remote management
2. **Shared Infrastructure**: Reuse internal packages (config, docker, version)
3. **Independent Operation**: CLI and daemon work independently
4. **Optional Integration**: Daemon is optional; CLI fully functional alone
5. **Graceful Degradation**: Failures in daemon don't affect CLI operations
