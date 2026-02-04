# Silo CLI

CLI tool to deploy and manage **Silo Box** on customer hardware.

Silo Box is a self-hosted AI chat application. This CLI installs and orchestrates Docker containers: PostgreSQL, backend, frontend, llama.cpp inference engine, and proxy agent.

## Requirements

- Docker 20.10+ with Compose v2
- Linux (Debian/Ubuntu) or macOS
- User in `docker` group (Linux) or Docker Desktop (macOS)
- 5GB+ free disk space

## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/EternisAI/silo/main/scripts/install.sh | bash
```

### Build from Source

```bash
git clone https://github.com/EternisAI/silo.git
cd silo
make build
cp bin/silo ~/.local/bin/
```

Add to PATH if needed (~/.bashrc or ~/.zshrc on macOS):

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Commands

### Start Services

```bash
silo up                    # first run installs, subsequent runs start
silo up --port 8080        # custom port (first install only)
silo up --image-tag 0.1.3  # specific version
```

Services auto-restart on system reboot (uses `restart: unless-stopped`).

### Stop Services

```bash
silo down                  # stops containers, preserves data
```

### Status

```bash
silo status                # show deployment and service status
```

### Logs

```bash
silo logs                  # view all logs
silo logs -f               # follow logs in real-time
silo logs --tail 50        # show last 50 lines
silo logs backend          # logs for specific service
```

### Upgrade

```bash
silo upgrade               # pull latest images and recreate containers
silo upgrade --json        # JSON output for automation
```

**Important:** Upgrade only updates Docker images (backend, frontend). It preserves:

- Configuration files
- Database volumes
- Model files
- Application data

To change the LLM model after upgrade:

1. Place new `.gguf` file in `~/.local/share/silo/data/models/`
2. Edit `inference_model_file` in `~/.config/silo/config.yml`
3. Run `silo down && silo up`

### Check Configuration

```bash
silo check                 # validate config and installation state
```

### Version

```bash
silo version               # show CLI and application versions
silo version --json        # JSON output
```

## Daemon (Remote Control)

HTTP API for remote management over LAN:

```bash
make build-daemon
./bin/silod                                    # localhost only
SILO_DAEMON_BIND_ADDRESS=0.0.0.0 ./bin/silod  # LAN access
```

### API Endpoints

```bash
curl http://localhost:9999/health                                      # Health check
curl http://localhost:9999/status | jq                                 # Container status
curl http://localhost:9999/api/v1/version | jq                         # Version info
curl -X POST http://localhost:9999/api/v1/up                           # Start/install
curl -X POST http://localhost:9999/api/v1/down                         # Stop
curl -X POST http://localhost:9999/api/v1/restart -d '{"service":"backend"}'
curl -X POST http://localhost:9999/api/v1/upgrade                      # Upgrade
curl "http://localhost:9999/api/v1/logs?service=backend&lines=50"
curl http://localhost:9999/api/v1/check | jq                           # Validate config
```

## Configuration

Edit `~/.config/silo/config.yml` then restart:

```bash
silo down && silo up
```

### Key Settings

| Setting                  | Default             | Description            |
| ------------------------ | ------------------- | ---------------------- |
| `port`                   | 80                  | Frontend port          |
| `image_tag`              | 0.1.2               | Docker image version   |
| `inference_model_file`   | GLM-4.7-Q4_K_M.gguf | LLM model file         |
| `inference_gpu_layers`   | 999                 | GPU layers (999 = all) |
| `inference_context_size` | 8192                | Token context window   |
| `inference_gpu_devices`  | "0", "1", "2"       | GPU device IDs         |

## Directory Structure

```
~/.local/bin/silo                      # CLI binary
~/.config/silo/config.yml              # Configuration file
~/.local/share/silo/
├── docker-compose.yml                 # Generated compose file
├── state.json                         # Installation state
└── data/
    ├── models/                        # LLM model files
    └── postgres/                      # Database data
```

## Global Flags

```bash
-v, --verbose              # Enable debug logging
--config-dir PATH          # Custom config directory
```

## Uninstall

```bash
silo down
rm -rf ~/.config/silo ~/.local/share/silo ~/.local/bin/silo
docker volume prune
```

## Development

```bash
make build                 # Build CLI binary to bin/silo
make build-daemon          # Build daemon binary to bin/silod
make test                  # Run tests
make fmt                   # Format code
make lint                  # Run golangci-lint
make dev ARGS="up"         # Run CLI without building
```

## Releasing

Releases are automated via GitHub Actions. **Do not manually create tags.**

### Creating a Release

1. Push your changes to `main` (via PR or direct push)
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

### Release Artifacts

Each release includes:
- `silo_<version>_linux_amd64.tar.gz`
- `silo_<version>_linux_arm64.tar.gz`
- `silo_<version>_darwin_amd64.tar.gz`
- `silo_<version>_darwin_arm64.tar.gz`
- `checksums.txt`

### CI Pipeline

Pull requests to `main` trigger CI checks:
- Build (both `silo` and `silod`)
- Tests
- Linting (golangci-lint)

## Related Projects

- **silo_box/** — The application this CLI deploys (Go backend + Next.js frontend)
