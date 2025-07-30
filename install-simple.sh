#!/bin/bash

# Simple one-liner installer for AWS to OpenTofu Transformer
# No Go required - just downloads the pre-built binary

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
VERSION="1.0.0"
REPO="kaviyarasu16/transformer"
INSTALL_DIR="${HOME}/.local/bin"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo -e "${RED}Unsupported architecture: $ARCH${NC}" && exit 1 ;;
esac

case $OS in
    linux) PLATFORM="linux" ;;
    darwin) PLATFORM="darwin" ;;
    *) echo -e "${RED}Unsupported OS: $OS${NC}" && exit 1 ;;
esac

echo -e "${BLUE}🚀 Installing AWS to OpenTofu Transformer CLI...${NC}"
echo -e "${BLUE}Platform: $PLATFORM-$ARCH${NC}"

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/v$VERSION/transformer-$PLATFORM-$ARCH"

echo -e "${BLUE}Downloading from GitHub releases...${NC}"

# Download binary
if curl -L -o "$INSTALL_DIR/transformer" "$DOWNLOAD_URL" --progress-bar; then
    chmod +x "$INSTALL_DIR/transformer"
    echo -e "${GREEN}✅ Installation successful!${NC}"
    echo -e "${BLUE}Location: $INSTALL_DIR/transformer${NC}"
    
    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "${BLUE}Adding to PATH for this session...${NC}"
        export PATH="$PATH:$INSTALL_DIR"
    fi
    
    # Test installation
    if "$INSTALL_DIR/transformer" --version &>/dev/null; then
        echo -e "${GREEN}✅ Installation verified!${NC}"
        echo
        echo -e "${BLUE}Usage:${NC}"
        echo "  transformer --help"
        echo "  transformer aws --region=us-east-1 --resources=vpc,ec2 --output=./infrastructure"
        echo
        echo -e "${BLUE}For permanent PATH access, add this to your shell profile:${NC}"
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    else
        echo -e "${RED}❌ Installation verification failed${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ Download failed${NC}"
    echo -e "${BLUE}You can manually download from: https://github.com/$REPO/releases${NC}"
    exit 1
fi 