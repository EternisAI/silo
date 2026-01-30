#!/bin/bash
set -e

REPO="EternisAI/silo"
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="silo"
DAEMON_NAME="silod"
SERVICE_FILE="silod.service"

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

main() {
  if [ "$(uname -s)" != "Linux" ]; then
    echo "Error: This script only supports Linux (Debian/Ubuntu)"
    exit 1
  fi

  check_dependencies

  echo "Installing Silo CLI..."

  VERSION=${1:-$(get_latest_release)}
  ARCH=$(detect_arch)
  PLATFORM="linux_${ARCH}"

  echo "Version: $VERSION"
  echo "Platform: $PLATFORM"

  DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/silo_${VERSION#v}_${PLATFORM}.tar.gz"
  TMP_DIR=$(mktemp -d)
  trap "rm -rf $TMP_DIR" EXIT

  curl -L "$DOWNLOAD_URL" | tar -xz -C "$TMP_DIR"

  mkdir -p "$INSTALL_DIR"
  mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
  mv "$TMP_DIR/$DAEMON_NAME" "$INSTALL_DIR/$DAEMON_NAME"
  chmod +x "$INSTALL_DIR/$BINARY_NAME"
  chmod +x "$INSTALL_DIR/$DAEMON_NAME"

  echo "✓ Silo CLI installed to $INSTALL_DIR/$BINARY_NAME"
  echo "✓ Silo Daemon installed to $INSTALL_DIR/$DAEMON_NAME"

  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "Add to PATH (add to ~/.bashrc):"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
  fi

  # Optional systemd service installation
  if command -v systemctl &> /dev/null; then
    echo ""
    read -p "Install silod as systemd service? (y/N) " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      echo "Installing systemd service..."
      sudo cp "$TMP_DIR/scripts/$SERVICE_FILE" /etc/systemd/system/$SERVICE_FILE
      sudo systemctl daemon-reload
      echo "✓ Service installed"
      echo ""
      read -p "Enable and start silod service now? (y/N) " -r
      if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo systemctl enable $DAEMON_NAME
        sudo systemctl start $DAEMON_NAME
        echo "✓ Service enabled and started"
        echo ""
        echo "Check status with: sudo systemctl status $DAEMON_NAME"
        echo "View logs with: sudo journalctl -u $DAEMON_NAME -f"
      else
        echo ""
        echo "Enable with: sudo systemctl enable $DAEMON_NAME"
        echo "Start with: sudo systemctl start $DAEMON_NAME"
      fi
    fi
  fi

  echo ""
  echo "Run 'silo --help' to get started"
}

main "$@"
