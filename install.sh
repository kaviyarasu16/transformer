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
INSTALLER_VERSION="1.0.0"

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
    echo "  -l, --local         Install from local source (requires Go)"
    echo "  -r, --remote        Install from remote repository (default)"
    echo
    echo "Installation Methods:"
    echo "  1. Remote install (default): go install github.com/kaviyarasu16/transformer@latest"
    echo "  2. Local install: go install . (from project directory)"
    echo "  3. Manual build: make build"
    echo
    echo "Examples:"
    echo "  $0                    # Install from remote repository"
    echo "  $0 --local            # Install from local source"
    echo "  $0 --force            # Force reinstall"
}

# Check if Go is installed
check_go() {
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
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go version $GO_VERSION found"
    
    # Check if GOPATH is set
    if [ -z "$GOPATH" ]; then
        GOPATH=$(go env GOPATH)
        print_warning "GOPATH not set, using default: $GOPATH"
    fi
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

# Install from remote repository
install_remote() {
    print_status "Installing AWS to OpenTofu Transformer CLI from remote repository..."
    
    # Install using go install
    go install github.com/kaviyarasu16/transformer@latest
    
    if [ $? -eq 0 ]; then
        print_success "Transformer CLI installed successfully from remote repository!"
    else
        print_error "Failed to install transformer CLI from remote repository"
        exit 1
    fi
}

# Install from local source
install_local() {
    print_status "Installing AWS to OpenTofu Transformer CLI from local source..."
    
    # Check if we're in the project directory
    if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
        print_error "Not in transformer project directory. Please run this script from the project root."
        exit 1
    fi
    
    # Install using go install
    go install .
    
    if [ $? -eq 0 ]; then
        print_success "Transformer CLI installed successfully from local source!"
    else
        print_error "Failed to install transformer CLI from local source"
        exit 1
    fi
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    if command -v transformer &> /dev/null; then
        TRANSFORMER_VERSION=$(transformer --version 2>/dev/null || echo "unknown")
        TRANSFORMER_PATH=$(which transformer)
        print_success "Transformer CLI is installed and working!"
        print_status "Version: $TRANSFORMER_VERSION"
        print_status "Location: $TRANSFORMER_PATH"
        
        # Test basic functionality
        if transformer --help &> /dev/null; then
            print_success "Basic functionality test passed"
        else
            print_warning "Basic functionality test failed"
        fi
    else
        print_error "Transformer CLI not found in PATH"
        print_warning "You may need to add Go bin directory to your PATH:"
        echo "  export PATH=\$PATH:\$(go env GOPATH)/bin"
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
    echo "  GitHub: https://github.com/kaviyarasu16/transformer"
    echo "  Issues: https://github.com/kaviyarasu16/transformer/issues"
}

# Parse command line arguments
FORCE="false"
INSTALL_METHOD="remote"

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
        -l|--local)
            INSTALL_METHOD="local"
            shift
            ;;
        -r|--remote)
            INSTALL_METHOD="remote"
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
    
    check_go
    check_existing
    
    case $INSTALL_METHOD in
        "local")
            install_local
            ;;
        "remote")
            install_remote
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