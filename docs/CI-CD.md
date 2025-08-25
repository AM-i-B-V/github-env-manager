# CI/CD Pipeline Documentation

This document describes the automated CI/CD pipeline for GitHub Environment Manager.

## 🚀 Overview

The CI/CD pipeline is fully automated and powered by [Release Drafter](https://github.com/release-drafter/release-drafter). It handles:

- **Release Note Generation**: Automatic release notes from commit messages and PRs
- **Smart Version Bumping**: Based on conventional commit messages
- **Draft Releases**: Created automatically on every push to main
- **Published Releases**: Built and released only when drafts are published
- **Cross-Platform Builds**: Builds for Linux, macOS, and Windows
- **File Updates**: Automatic updates to Homebrew formula and install scripts
- **Documentation**: Automatic documentation updates

## 📋 Workflows

### 1. Release Drafter Workflow (`.github/workflows/release-drafter.yml`)

**Triggers:**

- Push to main branch

**What it does:**

1. **Version Bumping**: Analyzes commit messages and bumps version automatically
2. **Tag Creation**: Creates and pushes new version tag
3. **Release Note Generation**: Creates/updates draft releases with notes from commits
4. **Draft Management**: Maintains draft releases for review

### 2. Build & Release Workflow (`.github/workflows/release.yml`)

**Triggers:**

- Release published event (when draft is published)
- Manual workflow dispatch

**What it does:**

1. **Version Detection**: Extracts version from published release
2. **Cross-Platform Build**: Builds binaries for all platforms
3. **SHA256 Calculation**: Calculates checksums for all binaries
4. **File Updates**: Updates Homebrew formula and install scripts
5. **Asset Upload**: Uploads binaries to the published release

## 🔧 Version Management

### Commit Message Conventions

The system uses conventional commit messages to determine version bumps:

```bash
feat: add new feature          # → minor version bump
fix: resolve bug              # → patch version bump
BREAKING CHANGE: major change # → major version bump
docs: update documentation    # → minor version bump (default)
chore: maintenance tasks      # → minor version bump (default)
```

### Version Script (`scripts/version.sh`)

The version management script provides several commands:

```bash
# Show current version and next versions
./scripts/version.sh version

# Auto-bump based on commit messages
./scripts/version.sh auto-bump

# Create patch release (0.0.x)
./scripts/version.sh release auto patch

# Create minor release (0.x.0)
./scripts/version.sh release auto minor

# Create major release (x.0.0)
./scripts/version.sh release auto major

# Create specific version
./scripts/version.sh release 1.2.3
```

### Makefile Commands

```bash
# Show current version and next versions
make version

# Auto-bump based on commit messages
make auto-bump

# Manual releases
make release-patch  # 0.0.x
make release-minor  # 0.x.0 (default)
make release-major  # x.0.0
```

## 🏗️ Build Process

### Supported Platforms

| Platform | Architecture | Binary Name              | Archive                                  |
| -------- | ------------ | ------------------------ | ---------------------------------------- |
| Linux    | AMD64        | `github-env-manager`     | `github-env-manager_linux_amd64.tar.gz`  |
| Linux    | ARM64        | `github-env-manager`     | `github-env-manager_linux_arm64.tar.gz`  |
| macOS    | AMD64        | `github-env-manager`     | `github-env-manager_darwin_amd64.tar.gz` |
| macOS    | ARM64        | `github-env-manager`     | `github-env-manager_darwin_arm64.tar.gz` |
| Windows  | AMD64        | `github-env-manager.exe` | `github-env-manager_windows_amd64.zip`   |

### Build Configuration

- **Go Version**: 1.25
- **Build Flags**: `-ldflags="-s -w -X main.version=$VERSION"`
- **CGO**: Disabled for static binaries
- **Optimization**: Stripped binaries for smaller size

## 📦 Release Process

### Automatic Updates

The release process automatically updates:

1. **Homebrew Formula** (`Formula/github-env-manager.rb`)

   - Version number
   - Download URLs
   - SHA256 checksums

2. **Install Scripts**

   - `scripts/install.sh` - Default version
   - `scripts/install.ps1` - Default version

3. **Documentation**
   - `README.md` - Version references
   - `INSTALL.md` - Version references

### Release Artifacts

Each release includes:

- Binary files for all platforms
- SHA256 checksums
- Release notes with installation instructions
- Automatic changelog generation

## 🔄 Workflow

### Creating a Release

#### Option 1: Automatic (Recommended)

```bash
# Push commits with conventional messages
git commit -m "feat: add new feature"
git push origin main

# Version bump and draft release created automatically
# Review and publish draft when ready
```

#### Option 3: Manual Release

```bash
# Using the version script
./scripts/version.sh release auto minor

# Using Makefile
make release-minor
```

#### Option 4: GitHub Actions

1. Go to Actions → Release workflow
2. Click "Run workflow"
3. Select version and release type
4. Click "Run workflow"

### Pull Request Workflow

1. **Create Feature Branch**

   ```bash
   git checkout -b feature/new-feature
   # Make changes
   git commit -m "feat: add new feature"
   git push origin feature/new-feature
   ```

2. **Create Pull Request**

   - Use conventional commit messages:
     - `feat:` for new features (minor bump)
     - `fix:` for bug fixes (patch bump)
     - `BREAKING CHANGE:` for breaking changes (major bump)

3. **Automated Checks**

   - Tests run automatically
   - Release Drafter updates draft release
   - Build verification

4. **Merge and Release**
   - Merge to main
   - Version bump and draft release created automatically
   - Review and publish draft when ready

## 🔐 Security

### Secrets Required

The workflows use the following secrets:

- `GITHUB_TOKEN` - Automatically provided by GitHub Actions

### Security Features

- **Non-root Docker builds**: Secure container builds
- **SHA256 Verification**: All binaries are checksummed
- **Signed Commits**: All automated commits are signed
- **Branch Protection**: Main branch is protected

## 📊 Monitoring

### Workflow Status

Monitor workflow status at:

- GitHub Actions tab in repository
- Release page for successful releases
- Issues for failed builds

### Notifications

- **Success**: Release created with binaries
- **Failure**: Issue created with error details
- **Manual**: Workflow dispatch for manual releases

## 🛠️ Troubleshooting

### Common Issues

#### Build Failures

```bash
# Check build locally
make build-all

# Check specific platform
GOOS=linux GOARCH=amd64 go build cmd/server/main.go
```

#### Version Issues

```bash
# Check current version
make version

# Reset version
git tag -d v1.2.3
git push origin :refs/tags/v1.2.3
```

#### Release Issues

```bash
# Check release workflow
# Go to Actions → Release workflow → View logs

# Manual release
# Go to Actions → Release workflow → Run workflow
```

### Debug Commands

```bash
# Check all tags
git tag -l

# Check workflow files
ls -la .github/workflows/

# Test version script
./scripts/version.sh version

# Test build process
make build-all
```

## 📈 Best Practices

### Version Management

- Use semantic versioning (MAJOR.MINOR.PATCH)
- Bump patch for bug fixes
- Bump minor for new features
- Bump major for breaking changes

### Release Process

- Always test locally before releasing
- Use descriptive release notes
- Include installation instructions
- Test installation on target platforms

### Development Workflow

- Use feature branches for development
- Add appropriate labels to PRs
- Test changes locally before pushing
- Review automated changes

## 🔮 Future Enhancements

### Planned Features

- [ ] Automated changelog generation
- [ ] Release notes from commit messages
- [ ] Automated dependency updates
- [ ] Security scanning integration
- [ ] Performance benchmarking
- [ ] Automated testing on all platforms

### Integration Opportunities

- [ ] Homebrew core submission
- [ ] Snap store integration
- [ ] Chocolatey package
- [ ] Scoop bucket
- [ ] Docker Hub publishing
