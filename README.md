# GitHub Environment Manager

A modern, web-based tool for managing GitHub environment variables and secrets across multiple repositories and environments. Built with Go and featuring a beautiful, responsive UI.

## Features

- 🔐 **GitHub Authentication** - Secure authentication using Personal Access Tokens (PAT)
- 🏗️ **Environment Management** - Create and manage GitHub environments
- 🔑 **Variables & Secrets** - Manage repository and environment-level variables and secrets
- 🔄 **Sync & Compare** - Sync variables between environments and compare configurations
- 📤 **Export/Import** - Export variables as .env files and import from existing configurations
- 🌐 **Real-time Updates** - WebSocket support for live updates
- 📱 **Responsive UI** - Modern, mobile-friendly interface built with Tailwind CSS
- 🔍 **Search & Filter** - Easy repository and environment discovery

## Installation

### Prerequisites

- Go 1.25.0 or higher
- GitHub Personal Access Token with appropriate permissions

### Quick Start

1. **Clone the repository**

   ```bash
   git clone https://github.com/your-username/github-env-manager.git
   cd github-env-manager
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Run the application**

   ```bash
   go run main.go
   ```

4. **Access the web interface**
   - The application will automatically open your browser to `http://localhost:8005`
   - If it doesn't open automatically, navigate to the URL manually

### Docker (Alternative)

```bash
# Build the Docker image
docker build -t github-env-manager .

# Run the container
docker run -p 8005:8005 github-env-manager
```

## Usage

### Authentication

1. Click "Connect with Token" in the web interface
2. Enter your GitHub Personal Access Token
3. The token should have the following permissions:
   - `repo` - Full control of private repositories
   - `admin:org` - Full control of organizations and teams
   - `workflow` - Update GitHub Action workflows

### Managing Environments

1. **Select a Repository**: Choose from your accessible GitHub repositories
2. **View Environments**: See existing environments or create new ones
3. **Manage Variables**: Add, edit, or delete environment variables
4. **Manage Secrets**: Handle sensitive data with GitHub secrets

### Key Operations

- **Create Environment**: Set up new deployment environments
- **Sync Variables**: Copy variables between environments
- **Export Configuration**: Download variables as .env files
- **Compare Environments**: View differences between environment configurations

## Configuration

### Command Line Options

```bash
# Run on custom port
go run main.go --port 8080

# Run on custom host
go run main.go --host 0.0.0.0

# Run with both custom host and port
go run main.go --host 0.0.0.0 --port 8080
```

## Development

### Project Structure

```
github-env-manager/
├── main.go              # Main application entry point
├── server.go            # Server implementation
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
├── Dockerfile           # Docker configuration
├── static/              # Static assets
│   ├── css/
│   │   └── style.css    # Custom styles
│   └── js/
│       └── app.js       # Frontend JavaScript
└── templates/
    └── index.html       # Main HTML template
```
