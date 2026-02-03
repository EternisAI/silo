# Testing Silo Upgrade Flow on macOS

This guide walks through testing the version upgrade detection and upgrade flow locally on macOS.

## Overview

The upgrade flow:
1. `silod` daemon checks Docker Hub for newer image tags
2. silo_box backend queries silod via Unix socket
3. Frontend displays "Update Available" in Settings
4. User clicks Upgrade → backend calls silod → silod pulls new images and restarts

## Prerequisites

- Docker Desktop for Mac running
- Go 1.21+

## Setup

### 1. Pull latest silo_cli

```bash
cd silo_local/silo_cli
git pull origin main
```

### 2. Build binaries

```bash
make build          # builds bin/silo
make build-daemon   # builds bin/silod
```

### 3. Create config with OLD image tag

The current latest is `v0.1.9`. We'll configure `v0.1.8` so silod detects an update.

```bash
mkdir -p ~/.config/silo ~/.local/share/silo

cat > ~/.config/silo/config.yml << 'EOF'
version: "dev"
image_tag: "0.1.8"
port: 3000

# LLM - use Ollama for local testing
llm_base_url: "http://host.docker.internal:11434/v1"
default_model: "qwen2.5:3b"

# Disable GPU inference engine (not available on macOS)
enable_inference_engine: false
enable_proxy_agent: false
EOF
```

### 4. Start silod manually

In terminal 1:

```bash
cd silo_local/silo_cli
./bin/silod
```

You should see:
```
INFO Starting Silo Daemon...
INFO Unix socket listening on /Users/YOU/.local/share/silo/silod.sock
INFO TCP server listening on 127.0.0.1:9999 (fallback)
```

### 5. Verify silod detects update

In terminal 2, query via Unix socket:

```bash
curl -s --unix-socket ~/.local/share/silo/silod.sock http://localhost/api/v1/version | jq
```

Expected output shows `needs_update: true` for images:
```json
{
  "cli": {
    "current": "dev",
    "latest": "v0.1.8",
    "needs_update": false
  },
  "images": [
    {
      "image_name": "Backend",
      "current": "0.1.8",
      "latest": "0.1.9",
      "needs_update": true
    },
    {
      "image_name": "Frontend", 
      "current": "0.1.8",
      "latest": "0.1.9",
      "needs_update": true
    }
  ]
}
```

### 6. Start silo containers

```bash
cd silo_local/silo_cli
./bin/silo up
```

This:
- Generates docker-compose.yml from template
- Mounts silod socket into backend container
- Starts postgres, backend, frontend

### 7. Test in UI

1. Open http://localhost:3000
2. Go to Settings
3. Should see "Update Available" indicator
4. Click Upgrade button
5. silod pulls v0.1.9 images and restarts containers

### 8. Verify upgrade completed

```bash
curl -s --unix-socket ~/.local/share/silo/silod.sock http://localhost/api/v1/version | jq '.images'
```

Should show `needs_update: false` now.

## Troubleshooting

### silod socket not found

Check socket exists:
```bash
ls -la ~/.local/share/silo/silod.sock
```

### Backend can't connect to silod

Check docker-compose.yml has the socket mount:
```bash
cat ~/.local/share/silo/docker-compose.yml | grep -A2 volumes
```

Should show:
```yaml
volumes:
  - /Users/YOU/.local/share/silo/silod.sock:/var/run/silod.sock
```

### View silod logs

silod logs to stdout when run manually. For container logs:
```bash
./bin/silo logs backend
```

## Cleanup

```bash
./bin/silo down
rm -rf ~/.config/silo ~/.local/share/silo
```
