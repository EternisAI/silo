#!/bin/bash
set -e

REPO="eternisai/silo"
INSTALL_DIR="/usr/local/bin"
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

  if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
  else
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
  fi

  echo "✓ Silo CLI installed successfully!"
  echo "Run 'silo --help' to get started"
}

main "$@"
