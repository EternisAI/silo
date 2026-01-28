# Silo CLI

CLI tool to deploy and manage **Silo Box** on customer hardware.

Silo Box is a self-hosted AI chat application. This CLI installs and orchestrates the Docker containers (postgres, backend, frontend, llama.cpp inference, proxy agent). The actual application code lives in `../silo_box/`.

## Commands

```bash
make build        # Build binary to bin/silo
make test         # Run tests
make fmt          # Format code
make lint         # Run golangci-lint
make dev ARGS="up"  # Run without building (go run)
```

## Structure

```
silo/
├── cmd/silo/           # CLI entry point
├── internal/
│   ├── assets/         # Embedded templates (docker-compose.yml.tmpl, config.yml.tmpl)
│   ├── cli/            # Cobra commands (up, down, status, logs, upgrade, check)
│   ├── config/         # Config loading/saving
│   ├── docker/         # Docker Compose operations
│   ├── installer/      # First-run installation logic
│   └── updater/        # Upgrade logic
├── pkg/logger/         # Logging utilities
└── scripts/            # install.sh, build.sh
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `silo up` | Install (first run) or start services |
| `silo down` | Stop services |
| `silo status` | Show service status |
| `silo logs` | View container logs |
| `silo upgrade` | Pull latest images and restart |
| `silo check` | Validate configuration file |

## Related

- **silo_box/** — The application this CLI deploys (Go backend + Next.js frontend)
