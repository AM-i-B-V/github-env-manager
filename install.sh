#!/bin/bash

# GitHub Environment Manager Installation Script
# This script handles all edge cases for running the tool with Docker

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CONTAINER_NAME="github-env-manager"
IMAGE_NAME="amibotuser/github-env-manager:latest"
PORT="8005"
HOST_PORT="8005"

# Alternative ports to try if the default port is in use
ALTERNATIVE_PORTS=(8006 8007 8008 8009 8010)

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

# Function to check if Docker is installed and running
check_docker() {
    print_status "Checking Docker installation..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        print_status "Visit: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    
    print_success "Docker is installed and running"
}

# Function to check if a specific port is available (cross-platform compatible)
is_port_available() {
    local port_to_check=$1
    local port_in_use=false
    
    # Try different methods based on OS
    case "$(uname -s)" in
        Darwin*)    # macOS
            if lsof -Pi :$port_to_check -sTCP:LISTEN -t >/dev/null 2>&1; then
                port_in_use=true
            fi
            ;;
        Linux*)     # Linux
            if command -v ss &> /dev/null; then
                if ss -tuln | grep -q ":$port_to_check "; then
                    port_in_use=true
                fi
            elif command -v netstat &> /dev/null; then
                if netstat -tuln | grep -q ":$port_to_check "; then
                    port_in_use=true
                fi
            elif lsof -Pi :$port_to_check -sTCP:LISTEN -t >/dev/null 2>&1; then
                port_in_use=true
            fi
            ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*)  # Windows
            if command -v netstat &> /dev/null; then
                if netstat -an | grep -q ":$port_to_check "; then
                    port_in_use=true
                fi
            fi
            ;;
        *)
            # Fallback to lsof if available
            if command -v lsof &> /dev/null; then
                if lsof -Pi :$port_to_check -sTCP:LISTEN -t >/dev/null 2>&1; then
                    port_in_use=true
                fi
            fi
            ;;
    esac
    
    [ "$port_in_use" = false ]
}

# Function to find an available port
find_available_port() {
    local original_port=$HOST_PORT
    
    # First try the original port
    if is_port_available $HOST_PORT; then
        print_success "Port $HOST_PORT is available"
        return 0
    fi
    
    print_warning "Port $HOST_PORT is already in use"
    
    # Check if it's our container
    if docker ps --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        print_status "Port is used by our existing container"
        return 0
    fi
    
    # Try alternative ports
    print_status "Trying alternative ports..."
    for alt_port in "${ALTERNATIVE_PORTS[@]}"; do
        if is_port_available $alt_port; then
            HOST_PORT=$alt_port
            print_success "Found available port: $HOST_PORT"
            return 0
        fi
    done
    
    # If no ports are available, show error
    print_error "No available ports found. Tried: $original_port, ${ALTERNATIVE_PORTS[*]}"
    print_status "Please free up a port or modify the ALTERNATIVE_PORTS array in this script."
    exit 1
}

# Function to check if port is available (cross-platform compatible)
check_port() {
    print_status "Checking for available port..."
    find_available_port
}

# Function to stop and remove existing container
cleanup_existing_container() {
    print_status "Checking for existing container..."
    
    if docker ps -a --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        print_warning "Found existing container '$CONTAINER_NAME'"
        
        # Check if container is running
        if docker ps --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
            print_status "Stopping existing container..."
            docker stop $CONTAINER_NAME
            print_success "Container stopped"
        fi
        
        print_status "Removing existing container..."
        docker rm $CONTAINER_NAME
        print_success "Container removed"
    else
        print_status "No existing container found"
    fi
}

# Function to pull latest image
pull_image() {
    print_status "Pulling latest image..."
    docker pull $IMAGE_NAME
    print_success "Image pulled successfully"
}

# Function to run the container
run_container() {
    print_status "Starting GitHub Environment Manager..."
    
    docker run -itd \
        -p $HOST_PORT:$PORT \
        --restart always \
        --name $CONTAINER_NAME \
        $IMAGE_NAME
    
    print_success "Container started successfully"
}

# Function to verify container is running
verify_container() {
    print_status "Verifying container is running..."
    
    # Wait a moment for container to fully start
    sleep 3
    
    if docker ps --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        print_success "Container is running successfully"
        
        # Check if the service is responding
        print_status "Checking if service is responding on port $HOST_PORT..."
        if curl -s -f http://localhost:$HOST_PORT > /dev/null 2>&1; then
            print_success "Service is responding on http://localhost:$HOST_PORT"
        else
            print_warning "Service might still be starting up. Please wait a moment and try accessing http://localhost:$HOST_PORT"
        fi
    else
        print_error "Container failed to start"
        print_status "Checking container logs..."
        docker logs $CONTAINER_NAME
        exit 1
    fi
}

# Function to open browser (cross-platform compatible)
open_browser() {
    local url="http://localhost:$HOST_PORT"
    print_status "Opening browser to $url..."
    
    # Detect OS and open browser accordingly
    case "$(uname -s)" in
        Darwin*)    # macOS
            if command -v open &> /dev/null; then
                open "$url"
                print_success "Browser opened on macOS"
            else
                print_warning "Could not open browser automatically on macOS"
            fi
            ;;
        Linux*)     # Linux (Ubuntu, etc.)
            if command -v xdg-open &> /dev/null; then
                xdg-open "$url" 2>/dev/null &
                print_success "Browser opened on Linux"
            elif command -v sensible-browser &> /dev/null; then
                sensible-browser "$url" 2>/dev/null &
                print_success "Browser opened on Linux"
            else
                print_warning "Could not open browser automatically on Linux"
            fi
            ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*)  # Windows
            if command -v start &> /dev/null; then
                start "$url"
                print_success "Browser opened on Windows"
            else
                print_warning "Could not open browser automatically on Windows"
            fi
            ;;
        *)
            print_warning "Unknown operating system, cannot open browser automatically"
            ;;
    esac
}

# Function to show usage instructions
show_instructions() {
    echo
    print_success "GitHub Environment Manager has been installed successfully!"
    echo
    echo "🌐 Access the application at: http://localhost:$HOST_PORT"
    echo
    echo "📋 Next steps:"
    echo "1. Your browser should open automatically to http://localhost:$HOST_PORT"
    echo "2. Click 'Connect with Token' to authenticate with GitHub"
    echo "3. Enter your GitHub Personal Access Token"
    echo "4. Start managing your environment variables and secrets!"
    echo
    echo "🔧 Useful commands:"
    echo "  View logs:     docker logs $CONTAINER_NAME"
    echo "  Stop service:  docker stop $CONTAINER_NAME"
    echo "  Start service: docker start $CONTAINER_NAME"
    echo "  Remove:        docker rm -f $CONTAINER_NAME"
    echo
    echo "📖 For more information, visit: https://github.com/your-repo/github-env-manager"
    echo
}

# Function to handle script interruption
cleanup_on_exit() {
    print_warning "Installation interrupted. Cleaning up..."
    if docker ps -a --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        docker rm -f $CONTAINER_NAME > /dev/null 2>&1 || true
    fi
    exit 1
}

# Set up signal handlers
trap cleanup_on_exit INT TERM

# Main installation process
main() {
    echo "🚀 GitHub Environment Manager Installation Script"
    echo "=================================================="
    echo
    
    # Check if running as root (not recommended for Docker)
    if [[ $EUID -eq 0 ]]; then
        print_warning "Running as root is not recommended for Docker operations"
        read -p "Do you want to continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status "Installation cancelled"
            exit 1
        fi
    fi
    
    # Run installation steps
    check_docker
    check_port
    cleanup_existing_container
    pull_image
    run_container
    verify_container
    show_instructions
    
    # Wait a moment for the service to be fully ready before opening browser
    print_status "Waiting for service to be fully ready..."
    sleep 2
    open_browser
}

# Run main function
main "$@"
