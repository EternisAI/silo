#!/bin/bash
set -e

REPO="EternisAI/silo"
INSTALL_DIR="$HOME/.local/bin"
CLI_BINARY="silo"
DAEMON_BINARY="silod"
SERVICE_NAME="silod.service"

get_latest_release() {
  curl --silent "https://api.github.com/repos/$REPO/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/'
}

detect_arch() {
  case "$(uname -m)" in
    x86_64) echo "amd64" ;;
    *)
      echo "Error: Unsupported architecture $(uname -m)"
      exit 1
      ;;
  esac
}

check_dependencies() {
  if ! command -v docker &> /dev/null; then
    echo "Error: docker is not installed"
    exit 1
  fi

  if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "Error: docker-compose is not installed"
    exit 1
  fi
}

install_systemd_service() {
  local silod_path="$INSTALL_DIR/$DAEMON_BINARY"
  
  if [ ! -f "$silod_path" ]; then
    echo "Error: $DAEMON_BINARY not found at $silod_path"
    exit 1
  fi
  
  echo ""
  echo "Installing Silo Daemon systemd service..."
  
  # Check if service is already running and stop it
  if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "Stopping existing $SERVICE_NAME..."
    sudo systemctl stop "$SERVICE_NAME"
  fi
  
  local current_user=$(whoami)
  local current_group=$(id -gn)
  
  echo "Service will run as user: $current_user"
  echo "Home directory: $HOME"
  echo "Binary path: $silod_path"
  
  local tmp_service_dir=$(mktemp -d)
  trap "rm -rf $tmp_service_dir" EXIT
  
  local service_url="https://raw.githubusercontent.com/$REPO/main/scripts/silod.service"
  echo "Downloading service template..."
  curl -fsSL "$service_url" -o "$tmp_service_dir/$SERVICE_NAME"
  
  echo "Configuring service..."
  sed -e "s|__USER__|$current_user|g" \
      -e "s|__GROUP__|$current_group|g" \
      -e "s|__HOME__|$HOME|g" \
      -e "s|__SILOD_PATH__|$silod_path|g" \
      "$tmp_service_dir/$SERVICE_NAME" > "$tmp_service_dir/${SERVICE_NAME}.configured"
  
  echo "Installing service to /etc/systemd/system/$SERVICE_NAME..."
  sudo cp "$tmp_service_dir/${SERVICE_NAME}.configured" "/etc/systemd/system/$SERVICE_NAME"
  sudo systemctl daemon-reload
  
  # Enable and start the service
  echo "Enabling and starting service..."
  sudo systemctl enable "$SERVICE_NAME"
  sudo systemctl start "$SERVICE_NAME"
  
  echo ""
  echo "✓ Silo Daemon service installed and started successfully"
  echo ""
  echo "Service management commands:"
  echo "  Check status:  sudo systemctl status $SERVICE_NAME"
  echo "  View logs:     sudo journalctl -u $SERVICE_NAME -f"
  echo "  Restart:       sudo systemctl restart $SERVICE_NAME"
  echo "  Stop:          sudo systemctl stop $SERVICE_NAME"
}

main() {
  if [ "$(uname -s)" != "Linux" ]; then
    echo "Error: This script only supports Linux (Debian/Ubuntu)"
    exit 1
  fi

  check_dependencies

  echo "Installing Silo CLI and Daemon..."

  VERSION=${1:-$(get_latest_release)}
  ARCH=$(detect_arch)
  PLATFORM="linux_${ARCH}"

  echo "Version: $VERSION"
  echo "Platform: $PLATFORM"

  DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/silo_${VERSION#v}_${PLATFORM}.tar.gz"
  TMP_DIR=$(mktemp -d)
  trap "rm -rf $TMP_DIR" EXIT

  echo "Downloading binaries..."
  curl -L "$DOWNLOAD_URL" | tar -xz -C "$TMP_DIR"

  mkdir -p "$INSTALL_DIR"
  
  # Install CLI binary
  if [ -f "$TMP_DIR/$CLI_BINARY" ]; then
    mv "$TMP_DIR/$CLI_BINARY" "$INSTALL_DIR/$CLI_BINARY"
    chmod +x "$INSTALL_DIR/$CLI_BINARY"
    echo "✓ Silo CLI installed to $INSTALL_DIR/$CLI_BINARY"
  else
    echo "Error: $CLI_BINARY binary not found in release archive"
    exit 1
  fi

  # Install daemon binary
  if [ -f "$TMP_DIR/$DAEMON_BINARY" ]; then
    mv "$TMP_DIR/$DAEMON_BINARY" "$INSTALL_DIR/$DAEMON_BINARY"
    chmod +x "$INSTALL_DIR/$DAEMON_BINARY"
    echo "✓ Silo Daemon installed to $INSTALL_DIR/$DAEMON_BINARY"
  else
    echo "Warning: $DAEMON_BINARY binary not found in release archive (skipping daemon installation)"
  fi

  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "Add to PATH (add to ~/.bashrc):"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
  fi

  echo ""
  echo "Run 'silo --help' to get started"
  
  # Install systemd service if daemon was installed
  if [ -f "$INSTALL_DIR/$DAEMON_BINARY" ]; then
    install_systemd_service
  fi
}

main "$@"
