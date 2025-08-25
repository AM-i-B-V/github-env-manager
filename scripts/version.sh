#!/bin/bash

# Version management script for GitHub Environment Manager
# This script handles version bumping and release preparation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to get current version from git tags
get_current_version() {
    local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    echo "${latest_tag#v}"
}

# Function to bump version
bump_version() {
    local current_version=$1
    local bump_type=$2
    
    IFS='.' read -ra VERSION_PARTS <<< "$current_version"
    local major=${VERSION_PARTS[0]}
    local minor=${VERSION_PARTS[1]}
    local patch=${VERSION_PARTS[2]}
    
    case $bump_type in
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "patch")
            patch=$((patch + 1))
            ;;
        *)
            print_error "Invalid bump type: $bump_type. Use major, minor, or patch"
            exit 1
            ;;
    esac
    
    echo "$major.$minor.$patch"
}

# Function to analyze commit messages and determine bump type
analyze_commits() {
    local since_tag=$1
    local major_count=0
    local minor_count=0
    local patch_count=0
    
    # Get commits since last tag
    local commits=$(git log --pretty=format:"%s" ${since_tag}..HEAD 2>/dev/null || git log --pretty=format:"%s")
    
    while IFS= read -r commit; do
        if [[ "$commit" == *"BREAKING CHANGE"* ]] || [[ "$commit" == *"!:"* ]]; then
            major_count=$((major_count + 1))
        elif [[ "$commit" == feat:* ]] || [[ "$commit" == *"feat:"* ]]; then
            minor_count=$((minor_count + 1))
        elif [[ "$commit" == fix:* ]] || [[ "$commit" == *"fix:"* ]]; then
            patch_count=$((patch_count + 1))
        else
            # Default to minor for other commits
            minor_count=$((minor_count + 1))
        fi
    done <<< "$commits"
    
    # Determine highest priority bump
    if [ $major_count -gt 0 ]; then
        echo "major"
    elif [ $minor_count -gt 0 ]; then
        echo "minor"
    elif [ $patch_count -gt 0 ]; then
        echo "patch"
    else
        echo "minor"  # Default to minor
    fi
}

# Function to create release
create_release() {
    local version=$1
    local bump_type=${2:-minor}
    
    print_status "Creating release for version $version"
    
    # Get current version
    local current_version=$(get_current_version)
    print_status "Current version: $current_version"
    
    # Bump version if not provided
    if [ "$version" = "auto" ]; then
        version=$(bump_version "$current_version" "$bump_type")
    fi
    
    print_status "New version: $version"
    
    # Check if tag already exists
    if git rev-parse "v$version" >/dev/null 2>&1; then
        print_error "Tag v$version already exists"
        exit 1
    fi
    
    # Create and push tag
    print_status "Creating tag v$version..."
    git tag "v$version"
    git push origin "v$version"
    
    print_success "Release v$version created and pushed!"
    print_status "Release Drafter will automatically update the release draft"
}

# Function to auto-bump based on commit messages
auto_bump() {
    local current_version=$(get_current_version)
    local current_tag="v$current_version"
    
    print_status "Analyzing commits since $current_tag..."
    
    # Determine bump type from commits
    local bump_type=$(analyze_commits "$current_tag")
    print_status "Detected bump type: $bump_type"
    
    # Create release
    create_release "auto" "$bump_type"
}

# Function to show current version
show_version() {
    local current_version=$(get_current_version)
    print_status "Current version: $current_version"
    
    # Show next versions
    local next_patch=$(bump_version "$current_version" "patch")
    local next_minor=$(bump_version "$current_version" "minor")
    local next_major=$(bump_version "$current_version" "major")
    
    echo "Next versions:"
    echo "  Patch: $next_patch"
    echo "  Minor: $next_minor"
    echo "  Major: $next_major"
}

# Function to show help
show_help() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  release [VERSION] [TYPE]  Create a new release"
    echo "  auto-bump                 Auto-bump version based on commit messages"
    echo "  version                   Show current version"
    echo "  help                      Show this help message"
    echo ""
    echo "Options:"
    echo "  VERSION                   Version number (e.g., 1.2.3) or 'auto'"
    echo "  TYPE                      Bump type: major, minor, patch (default: minor)"
    echo ""
    echo "Examples:"
    echo "  $0 auto-bump              # Auto-bump based on commits"
    echo "  $0 release auto patch     # Auto-bump patch version"
    echo "  $0 release auto minor     # Auto-bump minor version"
    echo "  $0 release 1.2.3          # Create specific version"
    echo "  $0 version                # Show current version"
    echo ""
    echo "Commit Message Conventions:"
    echo "  feat: new feature         -> minor bump"
    echo "  fix: bug fix              -> patch bump"
    echo "  BREAKING CHANGE           -> major bump"
    echo "  other commits             -> minor bump (default)"
}

# Main script logic
case "${1:-help}" in
    "release")
        version=${2:-auto}
        bump_type=${3:-minor}
        create_release "$version" "$bump_type"
        ;;
    "auto-bump")
        auto_bump
        ;;
    "version")
        show_version
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
