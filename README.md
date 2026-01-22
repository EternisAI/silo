# Silo CLI

A command-line tool to install, manage, and update the Silo application deployed via Docker Compose.

## Features

- **One-Command Installation**: Deploy Silo with PostgreSQL and pgvector in seconds
- **Easy Management**: Start, stop, update, and monitor your deployment
- **Docker-Based**: Leverages Docker Compose for reliable containerization
- **Configuration Management**: Simple YAML-based configuration
- **Automated Updates**: Pull and deploy new versions seamlessly

## Prerequisites

- Docker (version 20.10+)
- Docker Compose (v1 or v2)
- Linux or macOS
- User must be in the `docker` group (or use `sudo`)

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

### Install Silo

Deploy Silo application with default settings:

```bash
sudo silo install
```

Custom configuration:

```bash
sudo silo install --port 8080 --image-tag v1.2.3
```

Options:
- `--port` - Application port (default: 3000)
- `--image-registry` - Docker image registry (default: ghcr.io/eternisai)
- `--image-tag` - Docker image tag (default: latest)
- `--config-dir` - Installation directory (default: /opt/silo)

### Check Status

View deployment status and container health:

```bash
sudo silo status
```

### View Logs

View logs from all containers:

```bash
sudo silo logs
```

View logs from specific service:

```bash
sudo silo logs silo
sudo silo logs pgvector
```

Follow logs in real-time:

```bash
sudo silo logs -f
```

Show last 50 lines:

```bash
sudo silo logs -n 50
```

### Update Silo

Update to the latest version:

```bash
sudo silo update
```

This will:
1. Backup your current configuration
2. Pull the latest Docker images
3. Recreate containers with new images
4. Preserve all data

### Configuration Management

Show current configuration:

```bash
sudo silo config show
```

Edit configuration file:

```bash
sudo silo config edit
```

Set a configuration value:

```bash
sudo silo config set port 8080
```

After changing configuration, restart containers:

```bash
sudo silo update
```

### Uninstall

Stop containers (preserve data):

```bash
sudo silo uninstall
```

Remove everything including data:

```bash
sudo silo uninstall --purge-data
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
/opt/silo/                    # Default installation directory
├── config.yml                # Application configuration
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
sudo silo install --port 8080
```

### Insufficient disk space

```
Error: insufficient disk space: 2 GB available, 5 GB required
```

**Solution**: Free up disk space or use a different installation directory with more space:
```bash
sudo silo install --config-dir /path/to/larger/disk
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
make dev ARGS="install"
```

## Global Flags

All commands support these global flags:

- `--verbose, -v` - Enable debug logging
- `--config-dir` - Override installation directory (default: /opt/silo)

## License

[License information here]

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Support

For issues and questions:
- GitHub Issues: https://github.com/eternisai/silo/issues
- Documentation: https://github.com/eternisai/silo
