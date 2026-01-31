#!/bin/bash
set -e

REPO="EternisAI/silo"
INSTALL_DIR="$HOME/.local/bin"
CLI_BINARY="silo"
DAEMON_BINARY="silod"

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
  
  # Show daemon installation instructions if daemon was installed
  if [ -f "$INSTALL_DIR/$DAEMON_BINARY" ]; then
    echo ""
    echo "To install silod as a systemd service (optional):"
    echo "  curl -fsSL https://raw.githubusercontent.com/$REPO/main/scripts/install-service.sh | bash"
    echo ""
    echo "Or if you have the repo locally:"
    echo "  cd /path/to/silo && sudo make install-service"
  fi
}

main "$@"
