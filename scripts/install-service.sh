#!/bin/bash
set -e

REPO="EternisAI/silo"
SERVICE_NAME="silod.service"
DAEMON_NAME="silod"

echo "Installing Silo Daemon systemd service..."

# Find silod binary
if [ -f "$HOME/.local/bin/$DAEMON_NAME" ]; then
    SILOD_PATH="$HOME/.local/bin/$DAEMON_NAME"
    echo "Found $DAEMON_NAME at $SILOD_PATH"
elif [ -f "/usr/local/bin/$DAEMON_NAME" ]; then
    SILOD_PATH="/usr/local/bin/$DAEMON_NAME"
    echo "Found $DAEMON_NAME at $SILOD_PATH"
else
    echo "Error: $DAEMON_NAME not found"
    echo "Please install silod first:"
    echo "  curl -fsSL https://raw.githubusercontent.com/$REPO/main/scripts/install.sh | bash"
    exit 1
fi

# Get current user and group
CURRENT_USER=$(whoami)
CURRENT_GROUP=$(id -gn)

echo "Service will run as user: $CURRENT_USER"
echo "Home directory: $HOME"
echo "Binary path: $SILOD_PATH"

# Download service template
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

SERVICE_URL="https://raw.githubusercontent.com/$REPO/main/scripts/silod.service"
echo "Downloading service template..."
curl -fsSL "$SERVICE_URL" -o "$TMP_DIR/$SERVICE_NAME"

# Replace placeholders
echo "Configuring service..."
sed -e "s|__USER__|$CURRENT_USER|g" \
    -e "s|__GROUP__|$CURRENT_GROUP|g" \
    -e "s|__HOME__|$HOME|g" \
    -e "s|__SILOD_PATH__|$SILOD_PATH|g" \
    "$TMP_DIR/$SERVICE_NAME" > "$TMP_DIR/${SERVICE_NAME}.configured"

# Install service (requires sudo)
echo "Installing service to /etc/systemd/system/$SERVICE_NAME..."
sudo cp "$TMP_DIR/${SERVICE_NAME}.configured" "/etc/systemd/system/$SERVICE_NAME"
sudo systemctl daemon-reload

echo ""
echo "✓ Silo Daemon service installed successfully"
echo ""
echo "Next steps:"
echo "  1. Enable service to start on boot:"
echo "     sudo systemctl enable $SERVICE_NAME"
echo ""
echo "  2. Start service now:"
echo "     sudo systemctl start $SERVICE_NAME"
echo ""
echo "  3. Check service status:"
echo "     sudo systemctl status $SERVICE_NAME"
echo ""
echo "  4. View service logs:"
echo "     sudo journalctl -u $SERVICE_NAME -f"
