#!/bin/bash

# AWS to OpenTofu Transformer CLI Installer
# This script installs the transformer CLI tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Version
INSTALLER_VERSION="2.0.0"
TRANSFORMER_VERSION="1.0.0"

# GitHub repository
REPO="kaviyarasu16/transformer"
REPO_URL="https://github.com/$REPO"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${PURPLE}$1${NC}"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    case $OS in
        linux)
            PLATFORM="linux"
            ;;
        darwin)
            PLATFORM="darwin"
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
    
    print_status "Detected platform: $PLATFORM-$ARCH"
}

# Show help
show_help() {
    echo "AWS to OpenTofu Transformer CLI Installer v$INSTALLER_VERSION"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -v, --version       Show installer version"
    echo "  -f, --force         Force reinstall"
    echo "  -d, --directory     Specify installation directory (default: ~/.local/bin)"
    echo "  -g, --go-install    Install using Go (requires Go to be installed)"
    echo
    echo "Installation Methods:"
    echo "  1. Binary download (default): Downloads pre-built binary for your platform"
    echo "  2. Go install: go install github.com/kaviyarasu16/transformer@latest"
    echo
    echo "Examples:"
    echo "  $0                    # Install binary for current platform"
    echo "  $0 --force            # Force reinstall"
    echo "  $0 --directory /usr/local/bin  # Install to specific directory"
    echo "  $0 --go-install       # Install using Go"
}

# Check if transformer is already installed
check_existing() {
    if command -v transformer &> /dev/null; then
        EXISTING_VERSION=$(transformer --version 2>/dev/null || echo "unknown")
        print_warning "Transformer CLI is already installed (version: $EXISTING_VERSION)"
        
        if [ "$FORCE" != "true" ]; then
            echo
            read -p "Do you want to reinstall? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_status "Installation cancelled"
                exit 0
            fi
        fi
    fi
}

# Download binary from GitHub releases
download_binary() {
    local platform=$1
    local arch=$2
    local install_dir=$3
    
    print_status "Downloading transformer CLI for $platform-$arch..."
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Download the binary
    BINARY_NAME="transformer"
    if [ "$platform" = "windows" ]; then
        BINARY_NAME="transformer.exe"
    fi
    
    DOWNLOAD_URL="$REPO_URL/releases/download/v$TRANSFORMER_VERSION/transformer-$platform-$arch"
    if [ "$platform" = "windows" ]; then
        DOWNLOAD_URL="$REPO_URL/releases/download/v$TRANSFORMER_VERSION/transformer-$platform-$arch.exe"
    fi
    
    print_status "Downloading from: $DOWNLOAD_URL"
    
    # Download with curl
    if curl -L -o "$BINARY_NAME" "$DOWNLOAD_URL" --progress-bar; then
        print_success "Download completed"
    else
        print_error "Failed to download binary"
        print_warning "Trying alternative download method..."
        
        # Try alternative method
        if wget -O "$BINARY_NAME" "$DOWNLOAD_URL" 2>/dev/null; then
            print_success "Download completed (alternative method)"
        else
            print_error "Failed to download binary. Please check your internet connection."
            print_warning "You can manually download from: $REPO_URL/releases"
            exit 1
        fi
    fi
    
    # Make executable
    chmod +x "$BINARY_NAME"
    
    # Create installation directory if it doesn't exist
    mkdir -p "$install_dir"
    
    # Install the binary
    cp "$BINARY_NAME" "$install_dir/"
    
    # Clean up
    cd - > /dev/null
    rm -rf "$TEMP_DIR"
    
    print_success "Transformer CLI installed to $install_dir/$BINARY_NAME"
}

# Install using Go (alternative method)
install_with_go() {
    print_status "Installing using Go..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first:"
        echo
        echo "Installation options:"
        echo "  1. Official installer: https://golang.org/doc/install"
        echo "  2. Package managers:"
        echo "     - macOS: brew install go"
        echo "     - Ubuntu/Debian: sudo apt-get install golang-go"
        echo "     - CentOS/RHEL: sudo yum install golang"
        echo
        exit 1
    fi
    
    go install github.com/kaviyarasu16/transformer@latest
    
    if [ $? -eq 0 ]; then
        print_success "Transformer CLI installed successfully using Go!"
    else
        print_error "Failed to install transformer CLI using Go"
        exit 1
    fi
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    if command -v transformer &> /dev/null; then
        TRANSFORMER_VERSION_OUTPUT=$(transformer --version 2>/dev/null || echo "unknown")
        TRANSFORMER_PATH=$(which transformer)
        print_success "Transformer CLI is installed and working!"
        print_status "Version: $TRANSFORMER_VERSION_OUTPUT"
        print_status "Location: $TRANSFORMER_PATH"
        
        # Test basic functionality
        if transformer --help &> /dev/null; then
            print_success "Basic functionality test passed"
        else
            print_warning "Basic functionality test failed"
        fi
    else
        print_error "Transformer CLI not found in PATH"
        print_warning "You may need to add the installation directory to your PATH:"
        echo "  export PATH=\$PATH:$INSTALL_DIR"
        echo
        print_warning "Add this to your shell profile (.bashrc, .zshrc, etc.) for permanent access"
        exit 1
    fi
}

# Show usage instructions
show_usage() {
    print_success "Installation complete!"
    echo
    print_header "Usage Examples:"
    echo "  transformer --help"
    echo "  transformer aws --help"
    echo "  transformer aws --region=us-east-1 --resources=vpc,ec2 --output=./infrastructure"
    echo "  transformer aws --region=us-east-1 --all --output=./complete-infrastructure"
    echo
    print_header "Quick Start:"
    echo "  1. Configure AWS credentials:"
    echo "     aws configure"
    echo "  2. Discover specific resources:"
    echo "     transformer aws --region=us-east-1 --resources=vpc,ec2 --output=./my-infrastructure"
    echo "  3. Discover all resources:"
    echo "     transformer aws --region=us-east-1 --all --output=./complete-infrastructure"
    echo
    print_header "Documentation:"
    echo "  GitHub: $REPO_URL"
    echo "  Issues: $REPO_URL/issues"
}

# Parse command line arguments
FORCE="false"
INSTALL_METHOD="binary"
INSTALL_DIR="$HOME/.local/bin"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            echo "Installer version: $INSTALLER_VERSION"
            exit 0
            ;;
        -f|--force)
            FORCE="true"
            shift
            ;;
        -d|--directory)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -g|--go-install)
            INSTALL_METHOD="go"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Main installation process
main() {
    print_header "🚀 AWS to OpenTofu Transformer CLI Installer v$INSTALLER_VERSION"
    echo "================================================================"
    echo
    
    check_existing
    
    case $INSTALL_METHOD in
        "binary")
            detect_platform
            download_binary "$PLATFORM" "$ARCH" "$INSTALL_DIR"
            
            # Add to PATH if not already there
            if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
                print_warning "Adding $INSTALL_DIR to PATH for this session"
                export PATH="$PATH:$INSTALL_DIR"
            fi
            ;;
        "go")
            install_with_go
            ;;
        *)
            print_error "Unknown installation method: $INSTALL_METHOD"
            exit 1
            ;;
    esac
    
    verify_installation
    show_usage
}

# Run main function
main "$@" 