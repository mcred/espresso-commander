#!/bin/bash

# Build macOS .pkg installer for Espresso Commander
# Namespace: io.mcred

set -e

# Configuration
APP_NAME="Espresso Commander"
IDENTIFIER="io.mcred.espresso-commander"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
BINARY_NAME="espresso-commander"

# Directories
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_DIR/dist"
PKG_ROOT="$BUILD_DIR/pkg-root"
PKG_SCRIPTS="$BUILD_DIR/pkg-scripts"
OUTPUT_DIR="$BUILD_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Building macOS Package for $APP_NAME v$VERSION${NC}"
echo "================================================"

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf "$BUILD_DIR"
mkdir -p "$PKG_ROOT"
mkdir -p "$PKG_SCRIPTS"
mkdir -p "$OUTPUT_DIR"

# Build the binary if not exists
if [ ! -f "$PROJECT_DIR/bin/$BINARY_NAME" ]; then
    echo -e "${YELLOW}Building binary...${NC}"
    cd "$PROJECT_DIR"
    make build
    cd "$SCRIPT_DIR"
fi

# Create package structure
echo "Creating package structure..."

# Create directories in package root
mkdir -p "$PKG_ROOT/usr/local/bin"
mkdir -p "$PKG_ROOT/Library/LaunchDaemons"

# Copy binary
cp "$PROJECT_DIR/bin/$BINARY_NAME" "$PKG_ROOT/usr/local/bin/"
chmod 755 "$PKG_ROOT/usr/local/bin/$BINARY_NAME"

# Copy LaunchDaemon plist
cp "$SCRIPT_DIR/io.mcred.espresso-commander.plist" "$PKG_ROOT/Library/LaunchDaemons/"
chmod 644 "$PKG_ROOT/Library/LaunchDaemons/io.mcred.espresso-commander.plist"

# Create postinstall script
cat > "$PKG_SCRIPTS/postinstall" << 'EOF'
#!/bin/bash

# Post-installation script for Espresso Commander

# Create log files
touch /var/log/espresso-commander.log
touch /var/log/espresso-commander.error.log
chmod 644 /var/log/espresso-commander.log
chmod 644 /var/log/espresso-commander.error.log

# Load the LaunchDaemon
launchctl load /Library/LaunchDaemons/io.mcred.espresso-commander.plist 2>/dev/null || true

echo "Espresso Commander has been installed successfully!"
echo "Service: io.mcred.espresso-commander"
echo "API endpoint: http://localhost:8080/execute"
echo ""
echo "To check status: sudo launchctl list | grep io.mcred"
echo "To view logs: tail -f /var/log/espresso-commander.log"

exit 0
EOF
chmod 755 "$PKG_SCRIPTS/postinstall"

# Create preinstall script
cat > "$PKG_SCRIPTS/preinstall" << 'EOF'
#!/bin/bash

# Pre-installation script for Espresso Commander

# Stop and unload existing service if running
if launchctl list | grep -q "io.mcred.espresso-commander"; then
    echo "Stopping existing Espresso Commander service..."
    launchctl unload /Library/LaunchDaemons/io.mcred.espresso-commander.plist 2>/dev/null || true
fi

exit 0
EOF
chmod 755 "$PKG_SCRIPTS/preinstall"

# Build the package
echo "Building package..."
PKG_FILE="$OUTPUT_DIR/${APP_NAME// /-}-${VERSION}.pkg"

pkgbuild \
    --root "$PKG_ROOT" \
    --scripts "$PKG_SCRIPTS" \
    --identifier "$IDENTIFIER" \
    --version "$VERSION" \
    --install-location "/" \
    "$PKG_FILE"

# Verify the package
if [ -f "$PKG_FILE" ]; then
    echo -e "${GREEN}✓ Package created successfully!${NC}"
    echo "Package: $PKG_FILE"
    echo ""
    echo "Package information:"
    pkgutil --payload-files "$PKG_FILE" | head -10
    echo "..."
    echo ""
    echo "To install: sudo installer -pkg \"$PKG_FILE\" -target /"
    echo "To verify: pkgutil --pkg-info $IDENTIFIER"
else
    echo -e "${RED}✗ Failed to create package${NC}"
    exit 1
fi

# Optional: Create a distribution package with more options
echo ""
echo -e "${YELLOW}Creating distribution package with UI...${NC}"

# Create distribution XML
cat > "$BUILD_DIR/distribution.xml" << EOF
<?xml version="1.0" encoding="utf-8"?>
<installer-gui-script minSpecVersion="2">
    <title>$APP_NAME $VERSION</title>
    <organization>io.mcred</organization>
    <!-- welcome file="welcome.txt"/ -->
    <!-- license file="license.txt"/ -->
    <pkg-ref id="$IDENTIFIER"/>
    <options customize="never" require-scripts="false" hostArchitectures="x86_64,arm64"/>
    <domains enable_anywhere="false" enable_currentUserHome="false" enable_localSystem="true"/>
    <installation-check script="pm_install_check();"/>
    <script>
    function pm_install_check() {
        if(system.compareVersions(system.version.ProductVersion, '10.14.0') &lt; 0) {
            my.result.title = 'macOS 10.14 or later required';
            my.result.message = 'This software requires macOS 10.14 Mojave or later.';
            my.result.type = 'Fatal';
            return false;
        }
        return true;
    }
    </script>
    <pkg-ref id="$IDENTIFIER" version="$VERSION" onConclusion="none">EspressoCommander.pkg</pkg-ref>
    <choices-outline>
        <line choice="$IDENTIFIER"/>
    </choices-outline>
    <choice id="$IDENTIFIER" visible="false">
        <pkg-ref id="$IDENTIFIER"/>
    </choice>
</installer-gui-script>
EOF

# Create welcome text
cat > "$BUILD_DIR/welcome.txt" << EOF
Welcome to $APP_NAME Installer

This installer will guide you through the installation of $APP_NAME version $VERSION on your Mac.

$APP_NAME is a system command executor service that provides:
• Remote command execution via HTTP API
• System information gathering
• Network connectivity testing
• Automatic start on boot

The service will be installed to:
• Binary: /usr/local/bin/$BINARY_NAME
• Service: /Library/LaunchDaemons/io.mcred.espresso-commander.plist
• Logs: /var/log/espresso-commander.log

After installation, the service will start automatically and be available at:
http://localhost:8080/execute

Click Continue to proceed with the installation.
EOF

# Create license text
cat > "$BUILD_DIR/license.txt" << EOF
MIT License

Copyright (c) $(date +%Y) mcred

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF

# Copy the component package
cp "$PKG_FILE" "$BUILD_DIR/EspressoCommander.pkg"

# Build distribution package
DIST_PKG_FILE="$OUTPUT_DIR/${APP_NAME// /-}-${VERSION}-installer.pkg"
productbuild \
    --distribution "$BUILD_DIR/distribution.xml" \
    --resources "$BUILD_DIR" \
    --package-path "$BUILD_DIR" \
    "$DIST_PKG_FILE"

if [ -f "$DIST_PKG_FILE" ]; then
    echo -e "${GREEN}✓ Distribution package created successfully!${NC}"
    echo "Installer: $DIST_PKG_FILE"
    echo ""
    echo "This package includes:"
    echo "  • Welcome screen"
    echo "  • License agreement"
    echo "  • System requirements check"
    echo "  • Installation progress"
    echo ""
    echo "To install: Open the .pkg file in Finder or run:"
    echo "  sudo installer -pkg \"$DIST_PKG_FILE\" -target /"
else
    echo -e "${YELLOW}⚠ Distribution package creation failed, but component package is available${NC}"
fi

echo -e "${GREEN}Build complete!${NC}"