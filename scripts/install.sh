#!/bin/bash

# GitHub Environment Manager Installer
# Supports: macOS (Intel/Apple Silicon), Linux (Ubuntu/Debian), Windows

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="AM-i-B-V/github-env-manager"
VERSION="latest"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="github-env-manager"

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

# Function to detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    case "$(uname -s)" in
        Darwin*)
            os="darwin"
            case "$(uname -m)" in
                x86_64) arch="amd64" ;;
                arm64)  arch="arm64" ;;
                *)      print_error "Unsupported architecture: $(uname -m)" && exit 1 ;;
            esac
            ;;
        Linux*)
            os="linux"
            case "$(uname -m)" in
                x86_64) arch="amd64" ;;
                aarch64) arch="arm64" ;;
                *)      print_error "Unsupported architecture: $(uname -m)" && exit 1 ;;
            esac
            ;;
        MINGW*|MSYS*|CYGWIN*)
            os="windows"
            case "$(uname -m)" in
                x86_64) arch="amd64" ;;
                *)      print_error "Unsupported architecture: $(uname -m)" && exit 1 ;;
            esac
            ;;
        *)
            print_error "Unsupported operating system: $(uname -s)" && exit 1
            ;;
    esac
    
    echo "${os}_${arch}"
}

# Function to download and install binary
install_binary() {
    local platform=$1
    local download_url=""
    
    if [ "$VERSION" = "latest" ]; then
        download_url="https://github.com/${REPO}/releases/latest/download/github-env-manager_${platform}.tar.gz"
    else
        download_url="https://github.com/${REPO}/releases/download/v${VERSION}/github-env-manager_${platform}.tar.gz"
    fi
    
    print_status "Downloading binary for ${platform}..."
    
    # Create temporary directory
    local temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    # Download binary
    if ! curl -fsSL "$download_url" -o "${BINARY_NAME}.tar.gz"; then
        print_error "Failed to download binary from ${download_url}"
        exit 1
    fi
    
    # Extract binary
    tar -xzf "${BINARY_NAME}.tar.gz"
    
    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"
    
    # Install binary
    cp "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # Clean up
    cd - > /dev/null
    rm -rf "$temp_dir"
    
    print_success "Binary installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

# Function to add to PATH
add_to_path() {
    local shell_rc=""
    
    case "$SHELL" in
        */zsh)  shell_rc="${HOME}/.zshrc" ;;
        */bash) shell_rc="${HOME}/.bashrc" ;;
        *)      shell_rc="${HOME}/.profile" ;;
    esac
    
    if ! grep -q "$INSTALL_DIR" "$shell_rc" 2>/dev/null; then
        echo "" >> "$shell_rc"
        echo "# GitHub Environment Manager" >> "$shell_rc"
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$shell_rc"
        print_success "Added to PATH in $shell_rc"
        print_warning "Please restart your terminal or run: source $shell_rc"
    else
        print_status "Already in PATH"
    fi
}

# Function to verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        print_success "Installation verified!"
        print_status "Run '${BINARY_NAME} --help' to get started"
    else
        print_warning "Binary not found in PATH. Please restart your terminal or run:"
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
}

# Main installation process
main() {
    print_status "Installing GitHub Environment Manager..."
    
    # Detect platform
    local platform=$(detect_platform)
    print_status "Detected platform: $platform"
    
    # Install binary
    install_binary "$platform"
    
    # Add to PATH
    add_to_path
    
    # Verify installation
    verify_installation
    
    print_success "Installation complete! 🎉"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --repo)
            REPO="$2"
            shift 2
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --version VERSION    Install specific version (default: latest)"
            echo "  --repo REPO          GitHub repository (default: your-username/github-env-manager)"
            echo "  --install-dir DIR    Installation directory (default: ~/.local/bin)"
            echo "  -h, --help           Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main installation
main
