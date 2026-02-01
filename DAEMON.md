# Silo Daemon (`silod`)

Remote management service for Silo Box.

## Transition to Unix Socket (Plan)

The daemon is being updated to serve its API via a Unix domain socket instead of TCP port 9999.
- **Purpose**: Enhanced security (local-only) and simplified container communication.
- **Socket Path**: `~/.local/share/silo/silod.sock` (mounted to `/var/run/silod.sock` in containers).
- **Communication**: Backend service will connect via `http://unix:/var/run/silod.sock`.

## Overview

`silod` is a daemon service that provides an HTTP API for remote management of Silo Box containers. It allows users or external tools to perform operations that are normally handled by the Silo CLI via standard HTTP requests.

- **HTTP API**: Provides endpoints for container lifecycle (up, down, restart), upgrades, logs, and status checks.
- **Shared Configuration**: Uses the same configuration and state as the Silo CLI.
- **Graceful Shutdown**: Clean termination with signal handling.

## Architecture

The daemon shares internal packages with the CLI:
- `internal/config/` - Configuration and state management
- `internal/docker/` - Docker Compose operations
- `internal/version/` - Version checking
- `pkg/logger/` - Logging utilities

Daemon-specific packages:
- `internal/daemon/daemon.go` - Main orchestration
- `internal/daemon/server.go` - HTTP API server
- `internal/daemon/handlers.go` - API endpoint handlers

## Installation

### Build and Install

```bash
# Build daemon binary
make build-daemon

# Install to ~/.local/bin
make install-daemon

# Install as systemd service
make install-service
```

### Systemd Service

Enable and start the service:

```bash
# Enable auto-start on boot
sudo systemctl enable silod

# Start the service
sudo systemctl start silod

# Check status
sudo systemctl status silod

# View logs
sudo journalctl -u silod -f
```

## Configuration

The daemon uses the same configuration as the CLI:
- Config: `~/.config/silo/config.yml`
- State: `~/.local/share/silo/state.json`
- Docker Compose: `~/.local/share/silo/docker-compose.yml`

### Environment Variables

- `SILO_CONFIG_DIR`: Override config directory (default: `~/.config/silo`)
- `SILO_DATA_DIR`: Override data directory (default: `~/.local/share/silo`)
- `SILO_DAEMON_BIND_ADDRESS`: Override bind address (default: `0.0.0.0`)

### Daemon Settings

Default settings:
- Port: `9999` (TCP fallback)
- Bind Address: `0.0.0.0` (TCP fallback)
- Socket Path: `~/.local/share/silo/silod.sock` (Primary)

The daemon prioritizes the Unix domain socket. TCP listener is only used if the socket path is explicitly disabled or unavailable.

## HTTP API

The daemon provides a REST API on port `9999`.

### Basic Endpoints

#### `GET /health`
Basic health check.
```json
{
  "status": "ok"
}
```

#### `GET /status`
Detailed status including configuration, container states, and version info.

### Command Endpoints (`/api/v1/...`)

All command endpoints return a standard response format:
```json
{
  "success": true,
  "message": "Operation completed",
  "logs": [...]
}
```

#### `POST /api/v1/up`
Install or start Silo. Accepts optional `image_tag`, `port`, etc. in JSON body.

#### `POST /api/v1/down`
Stop Silo containers.

#### `POST /api/v1/restart`
Restart services. Accepts optional `service` name in JSON body.

#### `POST /api/v1/upgrade`
Upgrade Silo to the latest version.

#### `GET /api/v1/logs`
Fetch container logs. Parameters: `service`, `lines`.

#### `GET /api/v1/version`
Check for CLI and image updates.

#### `GET /api/v1/check`
Validate configuration and installation.

## Usage

### Manual Run (Development)

```bash
# Run with go run
make dev-daemon
```

### Query API

```bash
# Health check
curl http://127.0.0.1:9999/health

# Full status
curl http://127.0.0.1:9999/status | jq
```

## Lifecycle

### Startup
1. Load configuration from `~/.config/silo/config.yml`
2. Load state from `~/.local/share/silo/state.json`
3. Start HTTP API server on port 9999
4. Wait for shutdown signal

### Shutdown
1. Receive SIGINT or SIGTERM
2. Stop HTTP server gracefully
3. Exit cleanly

## Troubleshooting

### Daemon won't start

```bash
# Check daemon logs
sudo journalctl -u silod -n 50
```

### API not responding

```bash
# Check if daemon is running
sudo systemctl status silod

# Check port availability
sudo netstat -tlnp | grep 9999
```
