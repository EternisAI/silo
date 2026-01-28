# Silo CLI

CLI tool to deploy and manage Silo application.

## Requirements

- Docker 20.10+
- Debian/Ubuntu Linux
- User in `docker` group

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

Services auto-restart on system reboot (uses `restart: unless-stopped`).

### Stop

```bash
silo down
```

### Status

```bash
silo status
```

### Check Configuration

```bash
silo check                 # validate config file and installation
```

### Upgrade

```bash
silo upgrade               # pull latest images and restart
```

**Note:** `upgrade` pulls Docker images only. It does not download new LLM models or regenerate `docker-compose.yml` from config changes.

To change LLM model:
1. Place new `.gguf` file in `~/.local/share/silo/data/models/`
2. Edit `inference_model_file` in `~/.config/silo/config.yml`
3. Manually edit `~/.local/share/silo/docker-compose.yml` or reinstall

### Configuration

Edit `~/.config/silo/config.yml` then restart:

```bash
silo down && silo up
```

### Logs

```bash
silo logs -f             # follow logs
silo logs --tail 50      # show last 50 lines
silo logs backend        # logs for specific service
```

### Uninstall

```bash
silo down && rm -rf ~/.config/silo ~/.local/share/silo && docker volume prune
```

## Directory Structure

```
~/.local/bin/silo                    # CLI binary
~/.config/silo/config.yml            # Configuration
~/.local/share/silo/                 # Application state and data
```

## Development

```bash
make build                 # Build binary
make test                  # Run tests
```
