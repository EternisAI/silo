# Silo Daemon (`silod`)

Background service for monitoring and managing Silo Box containers.

## Overview

`silod` is a daemon service that runs alongside the Silo CLI. While the CLI (`silo`) provides on-demand container management, the daemon provides continuous monitoring and automated maintenance:

- **Container Health Monitoring**: Continuously checks container status
- **Auto-restart**: Automatically restarts failed containers
- **Scheduled Version Checks**: Daily checks for available updates
- **HTTP API**: Provides status endpoint for CLI queries
- **Graceful Shutdown**: Clean termination with signal handling

## Architecture

The daemon shares internal packages with the CLI:
- `internal/config/` - Configuration and state management
- `internal/docker/` - Docker Compose operations
- `internal/version/` - Version checking
- `pkg/logger/` - Logging utilities

Daemon-specific packages:
- `internal/daemon/daemon.go` - Main orchestration
- `internal/daemon/monitor.go` - Container health monitoring
- `internal/daemon/scheduler.go` - Scheduled tasks
- `internal/daemon/server.go` - HTTP API server

## Installation

### Build and Install

```bash
# Build daemon binary
make build-daemon

# Install to /usr/local/bin
make install-daemon

# Install as systemd service
make install-service

# All-in-one: CLI + daemon + service
make install-all
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

```bash
SILO_CONFIG_DIR=/path/to/config  # Override config directory
SILO_STATE_DIR=/path/to/state    # Override state directory
```

### Daemon Settings

Default configuration (in `daemon.go`):

```go
MonitorInterval:      30s              // Container check interval
VersionCheckCron:     "0 2 * * *"      // Daily at 2 AM
HealthCheckCron:      "*/5 * * * *"    // Every 5 minutes
AutoRestart:          true             // Auto-restart failed containers
ServerEnabled:        true             // Enable HTTP API
ServerPort:           9999             // API port
ServerBindAddress:    "0.0.0.0"        // Bind to all interfaces (allows container access)
```

## Features

### 1. Container Monitoring

Continuously monitors container health every 30 seconds:
- Detects state changes (running → exited)
- Auto-restarts exited containers (if enabled)
- Tracks monitoring statistics

### 2. Scheduled Tasks

**Version Checks** (Daily at 2 AM):
- Checks for CLI updates from GitHub
- Checks for image updates from Docker Hub
- Logs available updates

**Health Checks** (Every 5 minutes):
- Verifies container status
- Counts running containers
- Logs health metrics

### 3. HTTP API

REST API on `http://127.0.0.1:9999`:

**Access from:**
- **Host machine**: `http://127.0.0.1:9999` or `http://localhost:9999`
- **Docker containers**: `http://host.docker.internal:9999` (requires `extra_hosts` config)

The daemon binds to `0.0.0.0:9999` to allow access from Docker containers (e.g., backend UI). Port 9999 is not exposed externally in docker-compose, so it's only accessible from the host and containers on the same host.

#### `GET /health`
Basic health check:
```json
{
  "status": "ok"
}
```

#### `GET /status`
Detailed daemon status:
```json
{
  "State": {...},
  "Config": {...},
  "Containers": [...],
  "CLIVersion": {...},
  "ImageVersions": {...},
  "MonitorStats": {...}
}
```

#### `GET /stats`
Monitoring statistics:
```json
{
  "LastCheck": "2024-01-30T10:30:00Z",
  "CheckCount": 120,
  "RestartCount": 2,
  "FailedChecks": 0,
  "ContainerState": {
    "backend": "running",
    "frontend": "running",
    "postgres": "running"
  }
}
```

## Usage

### Manual Run (Development)

```bash
# Run directly (foreground)
make run-daemon

# Or with go run
make dev-daemon
```

### Production (Systemd)

```bash
# Start service
sudo systemctl start silod

# Stop service
sudo systemctl stop silod

# Restart service
sudo systemctl restart silod

# View logs
sudo journalctl -u silod -f

# View recent logs
sudo journalctl -u silod -n 100
```

### Query API

```bash
# Health check
curl http://127.0.0.1:9999/health

# Full status
curl http://127.0.0.1:9999/status | jq

# Monitoring stats
curl http://127.0.0.1:9999/stats | jq
```

## Lifecycle

### Startup
1. Load configuration from `~/.config/silo/config.yml`
2. Load state from `~/.local/share/silo/state.json`
3. Verify installation (must run `silo up` first)
4. Start monitor goroutine
5. Start scheduler goroutine
6. Start HTTP API server
7. Wait for shutdown signal

### Shutdown
1. Receive SIGINT or SIGTERM
2. Cancel context (stops all goroutines)
3. Stop HTTP server gracefully
4. Wait for goroutines to finish
5. Exit cleanly

## Monitoring Stats

The monitor tracks:
- **LastCheck**: Timestamp of last health check
- **CheckCount**: Total number of checks performed
- **RestartCount**: Number of automatic restarts
- **FailedChecks**: Number of checks that failed
- **ContainerState**: Current state of each container

## Auto-Restart

When `AutoRestart: true`:
- Monitors for containers in `exited` state
- Attempts restart using `docker compose restart`
- Logs success/failure
- Increments `RestartCount` on success

## Version Checks

Daily checks at 2 AM:
- Queries GitHub API for latest CLI release
- Queries Docker Hub for latest backend/frontend tags
- Compares current vs latest versions
- Logs if updates available (non-blocking)

## Error Handling

The daemon handles errors gracefully:
- Failed health checks logged as errors
- API timeouts don't crash daemon
- Invalid state displays helpful error messages
- Signal handling ensures clean shutdown

## Troubleshooting

### Daemon won't start

```bash
# Check if silo is installed
silo status

# If not, run:
silo up

# Check daemon logs
sudo journalctl -u silod -n 50
```

### API not responding

```bash
# Check if daemon is running
sudo systemctl status silod

# Check port availability
sudo netstat -tlnp | grep 9999

# Restart daemon
sudo systemctl restart silod
```

### Containers not auto-restarting

- Verify `AutoRestart: true` in daemon config
- Check daemon logs for restart attempts
- Ensure Docker Compose file exists
- Verify container restart policy

## Development

### Run Tests

```bash
make test
```

### Format Code

```bash
make fmt
```

### Lint

```bash
make lint
```

### Build for Development

```bash
# Build daemon
make build-daemon

# Run daemon (foreground)
./bin/silod
```

## Related Files

- `cmd/silod/main.go` - Daemon entry point
- `internal/daemon/` - Daemon implementation
- `scripts/silod.service` - Systemd service file
- `Makefile` - Build targets

## Comparison: CLI vs Daemon

| Feature | CLI (`silo`) | Daemon (`silod`) |
|---------|--------------|------------------|
| **Invocation** | On-demand | Background service |
| **Monitoring** | Manual status checks | Continuous monitoring |
| **Auto-restart** | No | Yes (configurable) |
| **Version checks** | Manual command | Scheduled daily |
| **API** | No | HTTP REST API |
| **Process** | Short-lived | Long-running |
| **Installation** | `silo up/down/status` | Container monitoring |
| **Use case** | User operations | Automated maintenance |

## Future Enhancements

Potential features for future releases:
- Configurable monitor intervals
- Email/Slack notifications on failures
- Auto-upgrade on version detection
- Resource usage metrics (CPU, memory, disk)
- Alert thresholds and triggers
- Web UI dashboard
- Log rotation and archival
