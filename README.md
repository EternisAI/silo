# Silo CLI

A command-line tool to deploy, manage, and upgrade the Silo application using Docker Compose.

## Features

- **One-Command Deployment**: Deploy Silo with PostgreSQL and pgvector in seconds
- **Simple Management**: Start, stop, upgrade, and monitor your deployment
- **Docker-Based**: Leverages Docker Compose for reliable containerization
- **YAML Configuration**: Simple YAML-based configuration stored in ~/.config/silo
- **Automated Upgrades**: Pull and deploy new versions seamlessly

## Prerequisites

- Docker (version 20.10+)
- Docker Compose (v1 or v2)
- Linux or macOS
- User must be in the `docker` group

## Installation

### Quick Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/eternisai/silo/main/scripts/install.sh | bash
```

### Manual Installation

1. Download the latest release for your platform from [GitHub Releases](https://github.com/eternisai/silo/releases)
2. Extract the archive:
   ```bash
   tar -xzf silo_*_linux_amd64.tar.gz
   ```
3. Move the binary to your PATH:
   ```bash
   sudo mv silo /usr/local/bin/
   sudo chmod +x /usr/local/bin/silo
   ```

### Build from Source

```bash
git clone https://github.com/eternisai/silo.git
cd silo
make build
sudo make install
```

## Usage

### Start Silo

Deploy and start Silo application (auto-installs on first run):

```bash
silo up
```

Custom configuration on first install:

```bash
silo up --port 8080 --image-tag v1.2.3
```

Options (first install only):
- `--port` - Application port (default: 3000)
- `--image-registry` - Docker image registry (default: ghcr.io/eternisai)
- `--image-tag` - Docker image tag (default: latest)
- `--config-dir` - Configuration directory (default: ~/.config/silo)

### Check Status

View deployment status and container health:

```bash
silo status
```

### View Logs

View logs using Docker Compose directly:

```bash
docker compose -f ~/.config/silo/docker-compose.yml logs -f
```

Or for specific service:

```bash
docker compose -f ~/.config/silo/docker-compose.yml logs -f silo
```

### Upgrade Silo

Upgrade to the latest version:

```bash
silo upgrade
```

This will:
1. Backup your current configuration
2. Pull the latest Docker images
3. Recreate containers with new images
4. Preserve all data

### Configuration Management

Edit configuration file with your preferred editor:

```bash
vi ~/.config/silo/config.yml
```

After changing configuration, restart containers:

```bash
silo down && silo up
```

### Stop Silo

Stop containers (preserve data):

```bash
silo down
```

Remove everything including data:

```bash
silo down && rm -rf ~/.config/silo && docker volume prune
```

### Version Information

```bash
silo version
```

## Architecture

Silo deploys two containers:

1. **silo-app**: The main application
   - Exposed on port 3000 (configurable)
   - Includes health checks
   - Auto-restarts on failure

2. **silo-pgvector**: PostgreSQL with pgvector extension
   - Data persisted in Docker volume
   - Includes health checks
   - Auto-restarts on failure

## Directory Structure

```
~/.config/silo/               # Default configuration directory
├── config.yml                # Application configuration (YAML)
├── docker-compose.yml        # Docker Compose configuration
├── state.json                # Installation state
└── data/                     # Application data directory
```

## Configuration File

Example `config.yml`:

```yaml
version: "0.1.0"
image_registry: "ghcr.io/eternisai"
image_tag: "latest"
port: 3000

app:
  log_level: "info"
  data_dir: "/data"
```

## Troubleshooting

### Docker not running

```
Error: docker is not running. Please start the Docker daemon
```

**Solution**: Start Docker:
```bash
sudo systemctl start docker
```

### Permission denied

```
Error: permission denied while trying to connect to the Docker daemon socket
```

**Solution**: Add your user to the docker group:
```bash
sudo usermod -aG docker $USER
newgrp docker
```

### Port already in use

```
Error: port 3000 is already in use
```

**Solution**: Install on a different port:
```bash
silo up --port 8080
```

### Insufficient disk space

```
Error: insufficient disk space: 2 GB available, 5 GB required
```

**Solution**: Free up disk space or use a different configuration directory with more space:
```bash
silo up --config-dir /path/to/larger/disk
```

## Development

### Build

```bash
make build
```

### Run tests

```bash
make test
```

### Format code

```bash
make fmt
```

### Run locally

```bash
make dev ARGS="up"
```

## Global Flags

All commands support these global flags:

- `--verbose, -v` - Enable debug logging
- `--config-dir` - Override configuration directory (default: ~/.config/silo)

## License

[License information here]

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Support

For issues and questions:
- GitHub Issues: https://github.com/eternisai/silo/issues
- Documentation: https://github.com/eternisai/silo
