# Silo CLI

CLI tool to deploy and manage Silo application using Docker Compose.

## Requirements

- Docker 20.10+
- Docker Compose
- Debian/Ubuntu Linux
- User in `docker` group

## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/eternisai/silo/main/scripts/install.sh | bash
```

### Build from Source

```bash
git clone https://github.com/eternisai/silo.git
cd silo
make build
cp bin/silo ~/.local/bin/
```

Add to PATH if needed (~/.bashrc):
```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Usage

### Start

```bash
silo up                    # first run installs, subsequent runs start
silo up --port 8080        # custom port (first install only)
```

### Stop

```bash
silo down
```

### Status

```bash
silo status
```

### Upgrade

```bash
silo upgrade               # pull latest images and restart
```

### Configuration

Edit `~/.config/silo/config.yml` then restart:

```bash
silo down && silo up
```

### Logs

```bash
docker compose -f ~/.local/share/silo/docker-compose.yml logs -f
```

### Uninstall

```bash
silo down && rm -rf ~/.config/silo ~/.local/share/silo && docker volume prune
```

## Directory Structure

```
~/.local/bin/silo                    # CLI binary
~/.config/silo/config.yml            # Configuration
~/.local/share/silo/                 # Docker compose, state, data
```

## Development

```bash
make build                 # Build binary
make test                  # Run tests
```
