# GitHub Environment Manager

A modern, web-based tool for managing GitHub repository variables and secrets across environments. Built with Go and featuring a beautiful, responsive UI.

## 🏗️ Project Structure

The project follows Go best practices and clean architecture principles:

```
github-env-manager/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go            # Configuration management
│   │   ├── errors.go            # Configuration errors
│   │   └── config_test.go       # Configuration tests
│   ├── handlers/
│   │   ├── auth.go              # Authentication handlers
│   │   └── handlers.go          # Placeholder handlers
│   ├── middleware/
│   │   └── middleware.go        # HTTP middleware
│   ├── models/
│   │   └── models.go            # Data models
│   ├── server/
│   │   └── server.go            # HTTP server setup
│   └── services/
│       └── auth.go              # Business logic
├── static/                      # Static assets
├── templates/                   # HTML templates
├── .gitignore                   # Git ignore rules
├── Dockerfile                   # Container configuration
├── Makefile                     # Build automation
├── go.mod                       # Go module dependencies
└── README.md                    # This file
```

## 🚀 Quick Start

### Prerequisites

- Go 1.25 or higher (only for development)
- Git (only for development)

### Easy Installation (Recommended)

#### macOS (Intel & Apple Silicon)

```bash
brew install AM-i-B-V/github-env-manager/github-env-manager
```

#### Ubuntu/Debian Linux

```bash
curl -fsSL https://raw.githubusercontent.com/AM-i-B-V/github-env-manager/main/scripts/install.sh | bash
```


#### Windows

**Option 1: PowerShell Script**

```powershell
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/AM-i-B-V/github-env-manager/main/scripts/install.ps1" -UseBasicParsing).Content
```

**Option 2: Manual Download**

1. Download from [GitHub Releases](https://github.com/AM-i-B-V/github-env-manager/releases)
2. Extract the ZIP file
3. Add the extracted directory to your PATH

### Run the Application

After installation, simply run:

```bash
github-env-manager
```

The application will automatically open at `http://localhost:8080`

### Development Installation

If you want to contribute or run from source:

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd github-env-manager
   ```

2. **Install dependencies**

   ```bash
   make deps
   # or
   go mod tidy
   ```

3. **Run the application**
   ```bash
   make run
   # or
   go run cmd/server/main.go
   ```

## 🛠️ Development

### Available Make Commands

```bash
make help          # Show available commands
make build         # Build the application
make run           # Run the application
make test          # Run tests
make test-coverage # Run tests with coverage
make clean         # Clean build artifacts
make deps          # Download dependencies
make lint          # Run linter
make format        # Format code
make docker-build  # Build Docker image
make docker-run    # Run Docker container
make install-tools # Install development tools
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/config/...

# Run benchmarks
make bench
```

### Code Quality

```bash
# Format code
make format

# Run linter
make lint

# Security scan
make security
```

## 🐳 Docker

### Build and Run with Docker

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Or manually
docker build -t github-env-manager .
docker run -p 8080:8080 github-env-manager
```

## 📁 Package Structure

### `cmd/server/`

Contains the main application entry point. This follows the Go standard layout where the `cmd` directory contains the main applications.

### `internal/`

Contains all internal packages that are not meant to be imported by other projects.

#### `internal/config/`

- **config.go**: Configuration structures and validation
- **errors.go**: Configuration-specific error definitions
- **config_test.go**: Unit tests for configuration

#### `internal/handlers/`

- **auth.go**: Authentication-related HTTP handlers
- **handlers.go**: Placeholder handlers for other endpoints

#### `internal/middleware/`

- **middleware.go**: HTTP middleware for logging, CORS, authentication, etc.

#### `internal/models/`

- **models.go**: Data structures used throughout the application

#### `internal/server/`

- **server.go**: HTTP server setup and route configuration

#### `internal/services/`

- **auth.go**: Business logic for authentication and session management

## 🔧 Configuration

The application uses environment variables for configuration. **All variables are optional** - the app runs with placeholder handlers by default.

| Variable | Description | Default   | Required |
| -------- | ----------- | --------- | -------- |
| `HOST`   | Server host | localhost | No       |
| `PORT`   | Server port | 8080      | No       |

### Authentication

The application currently supports Personal Access Token (PAT) authentication, which doesn't require OAuth client credentials. Users can authenticate by providing their own GitHub PAT directly through the UI.

## 🧪 Testing

The project includes a comprehensive testing structure:

- **Unit tests**: Test individual functions and methods
- **Integration tests**: Test component interactions
- **Coverage reports**: Track test coverage
- **Benchmarks**: Performance testing

### Test Structure

```
internal/
├── config/
│   └── config_test.go       # Configuration tests
├── handlers/
│   └── handlers_test.go     # Handler tests (TODO)
├── middleware/
│   └── middleware_test.go   # Middleware tests (TODO)
├── services/
│   └── auth_test.go         # Service tests (TODO)
└── server/
    └── server_test.go       # Server tests (TODO)
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write tests for new functionality
- Update documentation as needed
- Use conventional commit messages
- Ensure code passes linting and formatting

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Gin](https://github.com/gin-gonic/gin) web framework
- Styled with [Tailwind CSS](https://tailwindcss.com/)
- Icons from [Font Awesome](https://fontawesome.com/)
- GitHub API integration with [go-github](https://github.com/google/go-github)
