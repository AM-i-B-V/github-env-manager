# Installation Guide

This guide provides multiple installation methods for GitHub Environment Manager across different platforms.

## 🍎 macOS (Intel & Apple Silicon)

### Option 1: Homebrew (Recommended)

```bash
# Add the tap (one-time setup)
brew tap AM-i-B-V/github-env-manager

# Install the tool
brew install github-env-manager
```

### Option 2: Universal Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/AM-i-B-V/github-env-manager/main/scripts/install.sh | bash
```

### Option 3: Manual Installation

```bash
# Detect your architecture
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    PLATFORM="darwin_amd64"
elif [ "$ARCH" = "arm64" ]; then
    PLATFORM="darwin_arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Download and install
curl -fsSL https://github.com/your-username/github-env-manager/releases/latest/download/github-env-manager_${PLATFORM}.tar.gz | tar -xz
sudo mv github-env-manager /usr/local/bin/
```

## 🐧 Ubuntu/Debian Linux

### Option 1: Universal Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/AM-i-B-V/github-env-manager/main/scripts/install.sh | bash
```

### Option 2: Manual Installation

```bash
# Detect your architecture
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    PLATFORM="linux_amd64"
elif [ "$ARCH" = "aarch64" ]; then
    PLATFORM="linux_arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Download and install
curl -fsSL https://github.com/AM-i-B-V/github-env-manager/releases/latest/download/github-env-manager_${PLATFORM}.tar.gz | tar -xz
sudo mv github-env-manager /usr/local/bin/
```

### Option 3: Using Snap (if available)

```bash
sudo snap install github-env-manager
```

## 🪟 Windows

### Option 1: PowerShell Script

```powershell
# Run in PowerShell as Administrator
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/AM-i-B-V/github-env-manager/main/scripts/install.ps1" -UseBasicParsing).Content
```

### Option 2: Manual Installation

1. **Download the latest release**

   - Go to [GitHub Releases](https://github.com/AM-i-B-V/github-env-manager/releases)
   - Download `github-env-manager_windows_amd64.zip`

2. **Extract and install**

   ```powershell
   # Extract to a directory
   Expand-Archive -Path github-env-manager_windows_amd64.zip -DestinationPath C:\github-env-manager

   # Add to PATH (run as Administrator)
   [Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";C:\github-env-manager", "User")
   ```

### Option 3: Using Chocolatey (if available)

```powershell
choco install github-env-manager
```

### Option 4: Using Scoop (if available)

```powershell
scoop install github-env-manager
```

## 🚀 Usage

After installation, you can run the tool from anywhere:

```bash
# Start the application
github-env-manager

# Run on a specific port
github-env-manager --port 3000

# Run on all interfaces
github-env-manager --host 0.0.0.0

# Get help
github-env-manager --help
```

The application will automatically open in your default browser at `http://localhost:8080`.

## 🔧 Configuration

The tool uses minimal configuration and works out of the box. You can optionally set:

```bash
# Environment variables (optional)
export HOST="0.0.0.0"  # Bind to all interfaces
export PORT="3000"     # Use custom port
```

## 🆘 Troubleshooting

### Permission Denied

```bash
# Make sure the binary is executable
chmod +x /usr/local/bin/github-env-manager
```

### Command Not Found

```bash
# Check if the binary is in your PATH
which github-env-manager

# If not found, add to PATH manually
export PATH="/usr/local/bin:$PATH"
```

### Port Already in Use

```bash
# Use a different port
github-env-manager --port 3000
```

### Update the Tool

```bash
# Using the install script
curl -fsSL https://raw.githubusercontent.com/AM-i-B-V/github-env-manager/main/scripts/install.sh | bash

# Using Homebrew
brew upgrade github-env-manager
```

## 🧪 Verify Installation

```bash
# Check version
github-env-manager --version

# Check help
github-env-manager --help

# Test the application
github-env-manager &
curl http://localhost:8080/health
```

## 📦 Uninstall

### Homebrew

```bash
brew uninstall github-env-manager
```

### Manual Installation

```bash
# Remove binary
sudo rm /usr/local/bin/github-env-manager

# Remove from PATH (edit your shell config file)
# Remove the line: export PATH="/usr/local/bin:$PATH"
```

### Windows

```powershell
# Remove binary
Remove-Item C:\github-env-manager\github-env-manager.exe

# Remove from PATH
# Edit Environment Variables in System Properties
```
