# GitHub Environment Manager Installer for Windows
# Supports: Windows (x64)

param(
    [string]$Version = "latest",
    [string]$Repo = "AM-i-B-V/github-env-manager",
    [string]$InstallDir = "$env:USERPROFILE\.local\bin"
)

# Function to write colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Function to detect Windows architecture
function Get-Platform {
    $arch = (Get-WmiObject -Class Win32_Processor).Architecture
    
    switch ($arch) {
        0 { return "windows_amd64" }  # x86
        9 { return "windows_amd64" }  # x64
        default { 
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Function to download and install binary
function Install-Binary {
    param([string]$Platform)
    
    $binaryName = "github-env-manager.exe"
    $downloadUrl = ""
    
    if ($Version -eq "latest") {
        $downloadUrl = "https://github.com/$Repo/releases/latest/download/github-env-manager_${Platform}.zip"
    } else {
        $downloadUrl = "https://github.com/$Repo/releases/download/v$Version/github-env-manager_${Platform}.zip"
    }
    
    Write-Status "Downloading binary for $Platform..."
    
    # Create temporary directory
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    Set-Location $tempDir
    
    try {
        # Download binary
        $zipPath = Join-Path $tempDir "github-env-manager.zip"
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath
        
        # Extract binary
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
        
        # Create install directory if it doesn't exist
        if (!(Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        
        # Install binary
        $binaryPath = Join-Path $tempDir $binaryName
        $installPath = Join-Path $InstallDir $binaryName
        Copy-Item -Path $binaryPath -Destination $installPath -Force
        
        Write-Success "Binary installed to $installPath"
    }
    catch {
        Write-Error "Failed to download or install binary: $($_.Exception.Message)"
        exit 1
    }
    finally {
        # Clean up
        Set-Location $PWD
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Function to add to PATH
function Add-ToPath {
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
    if ($currentPath -notlike "*$InstallDir*") {
        $newPath = "$currentPath;$InstallDir"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Success "Added to PATH"
        Write-Warning "Please restart your terminal or run: refreshenv"
    } else {
        Write-Status "Already in PATH"
    }
}

# Function to verify installation
function Test-Installation {
    $binaryName = "github-env-manager.exe"
    $binaryPath = Join-Path $InstallDir $binaryName
    
    if (Test-Path $binaryPath) {
        Write-Success "Installation verified!"
        Write-Status "Run '$binaryName --help' to get started"
    } else {
        Write-Warning "Binary not found. Please restart your terminal or run:"
        Write-Host "  `$env:PATH += `";$InstallDir`""
    }
}

# Main installation process
function Main {
    Write-Status "Installing GitHub Environment Manager..."
    
    # Detect platform
    $platform = Get-Platform
    Write-Status "Detected platform: $platform"
    
    # Install binary
    Install-Binary -Platform $platform
    
    # Add to PATH
    Add-ToPath
    
    # Verify installation
    Test-Installation
    
    Write-Success "Installation complete! 🎉"
}

# Show help if requested
if ($args -contains "-h" -or $args -contains "--help") {
    Write-Host "Usage: .\install.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Version VERSION    Install specific version (default: latest)"
    Write-Host "  -Repo REPO          GitHub repository (default: your-username/github-env-manager)"
    Write-Host "  -InstallDir DIR     Installation directory (default: ~\.local\bin)"
    Write-Host "  -h, --help          Show this help message"
    exit 0
}

# Run main installation
Main
