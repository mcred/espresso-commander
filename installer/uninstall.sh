#!/bin/bash

# Espresso Commander macOS Uninstaller
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

echo -e "${YELLOW}Espresso Commander Uninstaller${NC}"
echo "================================"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

echo -e "${YELLOW}This will remove Espresso Commander from your system.${NC}"
read -p "Are you sure? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Uninstall cancelled."
    exit 0
fi

echo -e "${YELLOW}Uninstalling Espresso Commander...${NC}"

# Stop and unload service if running
if launchctl list | grep -q "io.mcred.espresso-commander"; then
    echo "Stopping service..."
    launchctl unload "$LAUNCHD_DIR/$PLIST_NAME" 2>/dev/null || true
fi

# Remove LaunchDaemon plist
if [ -f "$LAUNCHD_DIR/$PLIST_NAME" ]; then
    echo "Removing LaunchDaemon..."
    rm -f "$LAUNCHD_DIR/$PLIST_NAME"
fi

# Remove binary
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo "Removing binary..."
    rm -f "$INSTALL_DIR/$BINARY_NAME"
fi

# Ask about log files
echo -e "${YELLOW}Do you want to remove log files?${NC}"
read -p "Remove logs? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Removing log files..."
    rm -f "$LOG_DIR/espresso-commander.log"
    rm -f "$LOG_DIR/espresso-commander.error.log"
else
    echo "Log files preserved at:"
    echo "  - $LOG_DIR/espresso-commander.log"
    echo "  - $LOG_DIR/espresso-commander.error.log"
fi

# Verify uninstallation
if ! launchctl list | grep -q "io.mcred.espresso-commander" && \
   [ ! -f "$INSTALL_DIR/$BINARY_NAME" ] && \
   [ ! -f "$LAUNCHD_DIR/$PLIST_NAME" ]; then
    echo -e "${GREEN}✓ Espresso Commander has been uninstalled successfully${NC}"
else
    echo -e "${YELLOW}⚠ Some components may not have been removed completely${NC}"
    
    # Check what's still present
    if launchctl list | grep -q "io.mcred.espresso-commander"; then
        echo "  - Service is still registered"
    fi
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        echo "  - Binary still exists at $INSTALL_DIR/$BINARY_NAME"
    fi
    if [ -f "$LAUNCHD_DIR/$PLIST_NAME" ]; then
        echo "  - LaunchDaemon plist still exists"
    fi
fi

echo -e "${GREEN}Uninstall complete!${NC}"