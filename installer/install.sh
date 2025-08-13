#!/bin/bash

# Espresso Commander macOS Installer
# Namespace: io.mcred

set -e

BINARY_NAME="espresso-commander"
INSTALL_DIR="/usr/local/bin"
LAUNCHD_DIR="/Library/LaunchDaemons"
PLIST_NAME="io.mcred.espresso-commander.plist"
LOG_DIR="/var/log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Espresso Commander Installer${NC}"
echo "================================"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

# Check if binary exists
if [ ! -f "./bin/$BINARY_NAME" ]; then
    echo -e "${YELLOW}Building binary...${NC}"
    make build
fi

echo -e "${GREEN}Installing Espresso Commander...${NC}"

# Create directories if they don't exist
mkdir -p "$INSTALL_DIR"
mkdir -p "$LOG_DIR"

# Stop existing service if running
if launchctl list | grep -q "io.mcred.espresso-commander"; then
    echo "Stopping existing service..."
    launchctl unload "$LAUNCHD_DIR/$PLIST_NAME" 2>/dev/null || true
fi

# Copy binary
echo "Installing binary to $INSTALL_DIR..."
cp "./bin/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
chmod 755 "$INSTALL_DIR/$BINARY_NAME"

# Copy LaunchDaemon plist
echo "Installing LaunchDaemon..."
cp "./installer/$PLIST_NAME" "$LAUNCHD_DIR/$PLIST_NAME"
chmod 644 "$LAUNCHD_DIR/$PLIST_NAME"
chown root:wheel "$LAUNCHD_DIR/$PLIST_NAME"

# Create log files
touch "$LOG_DIR/espresso-commander.log"
touch "$LOG_DIR/espresso-commander.error.log"
chmod 644 "$LOG_DIR/espresso-commander.log"
chmod 644 "$LOG_DIR/espresso-commander.error.log"

# Load the service
echo "Starting service..."
launchctl load "$LAUNCHD_DIR/$PLIST_NAME"

# Verify installation
if launchctl list | grep -q "io.mcred.espresso-commander"; then
    echo -e "${GREEN}✓ Service installed and started successfully${NC}"
    echo ""
    echo "Service Information:"
    echo "  - Binary: $INSTALL_DIR/$BINARY_NAME"
    echo "  - Service: io.mcred.espresso-commander"
    echo "  - Logs: $LOG_DIR/espresso-commander.log"
    echo "  - Errors: $LOG_DIR/espresso-commander.error.log"
    echo "  - API: http://localhost:8080/execute"
    echo ""
    echo "Commands:"
    echo "  - Check status: sudo launchctl list | grep io.mcred"
    echo "  - View logs: tail -f $LOG_DIR/espresso-commander.log"
    echo "  - Stop service: sudo launchctl unload $LAUNCHD_DIR/$PLIST_NAME"
    echo "  - Start service: sudo launchctl load $LAUNCHD_DIR/$PLIST_NAME"
    echo "  - Uninstall: sudo ./installer/uninstall.sh"
else
    echo -e "${RED}✗ Failed to start service${NC}"
    exit 1
fi

echo -e "${GREEN}Installation complete!${NC}"