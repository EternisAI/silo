#!/bin/bash
set -e

REPO="eternisai/silo"
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="silo"

get_latest_release() {
  curl --silent "https://api.github.com/repos/$REPO/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/'
}

detect_platform() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)

  case "$ARCH" in
    x86_64)
      ARCH="amd64"
      ;;
    aarch64|arm64)
      ARCH="arm64"
      ;;
    *)
      echo "Unsupported architecture: $ARCH"
      exit 1
      ;;
  esac

  echo "${OS}_${ARCH}"
}

main() {
  echo "Installing Silo CLI..."

  VERSION=${1:-$(get_latest_release)}
  PLATFORM=$(detect_platform)

  echo "Version: $VERSION"
  echo "Platform: $PLATFORM"

  DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/silo_${VERSION#v}_${PLATFORM}.tar.gz"

  echo "Downloading from: $DOWNLOAD_URL"

  TMP_DIR=$(mktemp -d)
  trap "rm -rf $TMP_DIR" EXIT

  curl -L "$DOWNLOAD_URL" | tar -xz -C "$TMP_DIR"

  if [ ! -f "$TMP_DIR/$BINARY_NAME" ]; then
    echo "Error: Binary not found in downloaded archive"
    exit 1
  fi

  echo "Installing to $INSTALL_DIR..."
  mkdir -p "$INSTALL_DIR"
  mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
  chmod +x "$INSTALL_DIR/$BINARY_NAME"

  echo "✓ Silo CLI installed successfully!"
  echo ""

  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "⚠ Note: $INSTALL_DIR is not in your PATH"
    echo "Add the following to your ~/.bashrc or ~/.zshrc:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
  fi

  echo ""
  echo "Run 'silo --help' to get started"
}

main "$@"
