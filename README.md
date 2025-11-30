# Tech Stack Analyzer

A high-performance technology stack analyzer written in Go, re-implementing [specfy/stack-analyser](https://github.com/specfy/stack-analyser) with  improvements and extended technology support.

## What This Project Does

The Tech Stack Analyzer automatically detects technologies, frameworks, databases, and tools used in codebases by analyzing files, dependencies, and configurations. It provides comprehensive insights into:

- **Programming Languages** - Detects source code languages and versions
- **Package Managers** - Identifies npm, pip, cargo, composer, nuget, maven dependencies
- **Frameworks** - Detects .NET, Spring Boot, Angular, React, Django frameworks
- **Databases** - Identifies PostgreSQL, MySQL, MongoDB, Redis, Oracle, SQL Server
- **Infrastructure** - Detects Docker, Kubernetes, Terraform, GitLab configurations
- **DevOps Tools** - Identifies CI/CD pipelines, monitoring, and deployment tools

## How to Use It

### Prerequisites

- **Go 1.19+** - For building from source
- **[Task](https://taskfile.dev)** (optional) - Task runner for build automation (see installation below)
- **Docker** (optional) - For containerized deployment

### Installation

#### Option 1: Build from Source
```bash
# Clone the repository
git clone https://github.com/petrarca/tech-stack-analyzer.git
cd tech-stack-analyzer

# Build stack-analyzer
go build -o bin/stack-analyzer ./cmd/scanner

# Or use Task (recommended)
task build
```

#### Option 2: Install Directly
```bash
go install github.com/petrarca/tech-stack-analyzer/cmd/scanner@latest
```

### Basic Usage

```bash
# Scan current directory
./bin/stack-analyzer

# Scan specific directory
./bin/stack-analyzer /path/to/project

# Save results to file
./bin/stack-analyzer /path/to/project --output results.json

# Pretty print JSON output (enabled by default)
./bin/stack-analyzer /path/to/project --pretty

# Compact JSON output
./bin/stack-analyzer /path/to/project --pretty=false

# Exclude specific directories
./bin/stack-analyzer /path/to/project --exclude-dir "vendor,node_modules,bin"
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `path` | Project path to analyze (positional argument) | `.` |
| `--output` | Output file path | stdout |
| `--pretty` | Pretty print JSON output | `true` |
| `--exclude-dir` | Comma-separated directories to exclude | none |
| `--validate` | Validate rules and exit | `false` |
| `--version` | Show version information | `false` |

### Detection Approach

The scanner uses a **two-tier detection system** to provide comprehensive technology analysis:

#### **Universal Technology Detection**
- **Works with all technologies** through pattern matching and dependency analysis
- **700+ technology rules** covering databases, frameworks, tools, cloud services
- **Multi-language support** - detects npm packages, Python libraries, Java dependencies, etc.
- **File-based detection** - identifies technologies through configuration files, environment variables, and file patterns

#### **Specialized Component Detectors**
For major technology stacks, the scanner provides **deep analysis** through dedicated component detectors:

- **Python Projects**: Parses `pyproject.toml`, `requirements.txt`, `setup.py` to extract package versions, dependencies, and project metadata
- **Node.js Applications**: Analyzes `package.json`, `package-lock.json`, `yarn.lock` for npm/yarn packages, scripts, and project configuration
- **Java/Kotlin Projects**: Parses `pom.xml` (Maven) and `build.gradle` (Gradle) for dependencies, plugins, and project structure
- **.NET Applications**: Extracts NuGet packages, framework versions, and project references from `.csproj` files
- **Docker Environments**: Analyzes `docker-compose.yml` and `Dockerfile` to identify services, networks, and container configurations
- **Terraform Infrastructure**: Parses HCL files to detect cloud providers, resources, and infrastructure components
- **Ruby Projects**: Extracts gems and dependencies from `Gemfile` and `gemspec` files
- **Rust Projects**: Analyzes `Cargo.toml` for crate dependencies and project metadata
- **PHP Projects**: Parses `composer.json` for PHP packages and project configuration
- **Deno Projects**: Analyzes `deno.json` and import maps for Deno-specific dependencies

#### **What This Means**

**Universal Detection** ensures that even technologies without specialized detectors are identified through:
- Package manager dependencies (npm, pip, cargo, composer, nuget, maven)
- Docker image references
- Configuration file patterns
- Environment variable prefixes
- File extensions and naming conventions

**Specialized Detection** provides enhanced capabilities for major stacks:
- **Version extraction** - Exact package versions, not just presence
- **Dependency relationships** - Maps how components depend on each other
- **Configuration analysis** - Extracts meaningful settings and metadata
- **Project structure** - Identifies sub-projects, modules, and services
- **Edge relationships** - Builds dependency graphs between components

This hybrid approach ensures **broad coverage** (detects almost any technology) while providing **deep insights** for the most common technology stacks.

### Output Structure

The scanner outputs a hierarchical JSON structure representing the detected technologies:

- **id**: Unique identifier for each component
- **name**: Component name (e.g., "main", "frontend", "backend")
- **path**: File system path relative to the project root
- **tech**: Primary technology detected for this component
- **techs**: Array of all technologies detected in this component
- **languages**: Object mapping programming languages to file counts
- **dependencies**: Array of detected dependencies with format `[type, name, version]`
- **childs**: Array of nested components (sub-projects, services, etc.)
- **edges**: Array of relationships to other components with read/write access
- **inComponent**: Reference to parent component if this is a nested component
- **licenses**: Array of detected licenses in this component
- **reason**: Array explaining why technologies were detected

### Example Output

```json
{
  "id": "abc123",
  "name": "main",
  "path": ["/"],
  "tech": "nodejs",
  "techs": ["nodejs", "express", "postgresql"],
  "languages": {
    "TypeScript": 45,
    "JavaScript": 12
  },
  "dependencies": [
    ["npm", "express", "^4.18.0"],
    ["npm", "pg", "^8.8.0"]
  ],
  "childs": [
    {
      "id": "def456",
      "name": "frontend",
      "tech": "react",
      "dependencies": [["npm", "react", "^18.2.0"]]
    }
  ]
}
```

## How to Build It

### Using Task (Recommended)

**Task** is a modern task runner that simplifies common development operations. The `Taskfile.yml` defines reusable commands for building, testing, and maintaining the project.

```bash
# Install Task (if not already installed)
# macOS
brew install go-task

# Or install directly with Go
go install github.com/go-task/task/v3/cmd/task@latest

# Build the project
task build

# Run all quality checks (format, check, test)
task fct

# Clean build artifacts
task clean

# Run the scanner (use -- <path>)
task run -- /path/to/project
```

#### Available Tasks

| Task | Description |
|------|-------------|
| `task build` | Compile the stack-analyzer binary |
| `task format` | Format Go code using gofmt |
| `task check` | Run go vet and golangci-lint |
| `task test` | Run all tests |
| `task fct` | Run format, check, and test in sequence |
| `task clean` | Clean up build artifacts and caches |
| `task run` | Run stack-analyzer on a directory |
| `task run:help` | Show stack-analyzer help message |
| `task pre-commit:setup` | Install pre-commit tool |
| `task pre-commit:install` | Install pre-commit git hooks |
| `task pre-commit:run` | Run pre-commit on all files |

### Using Go Commands

```bash
# Build stack-analyzer
go build -o bin/stack-analyzer ./cmd/scanner

# Run tests
go test -v ./...

# Run with race detection
go test -race ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/stack-analyzer-linux ./cmd/scanner
GOOS=windows GOARCH=amd64 go build -o bin/stack-analyzer-windows.exe ./cmd/scanner
```

### Docker Build

```bash
# Build Docker image
docker build -t tech-stack-analyzer .

# Run in container
docker run --rm -v /path/to/project:/app tech-stack-analyzer /app
```

## Architecture Overview

### Project Structure

```
tech-stack-analyzer/
├── cmd/
│   ├── scanner/           # CLI application entry point
│   └── convert-rules/     # Rules conversion utilities
├── internal/
│   ├── provider/          # File system abstraction layer
│   ├── rules/             # Rule loading and validation
│   │   └── core/          # Embedded technology rules (700+ rules in 32 categories)
│   ├── scanner/           # Core scanning engine
│   │   ├── components/    # Component detectors (11 detectors)
│   │   ├── matchers/      # File and extension matchers
│   │   └── parsers/       # Specialized file parsers
│   └── types/             # Core data structures
├── docs/                  # Documentation
└── Taskfile.yml           # Task automation
```

### Core Components

#### 1. Scanner Engine (`internal/scanner/`)
- **Main orchestrator** that coordinates all detection phases
- **Sequential processing** with efficient recursive traversal
- **Component detection** through modular detector system

#### 2. Component Detectors (`internal/scanner/components/`)
Each detector handles specific project types:
- **Node.js** - package.json, npm/yarn detection
- **Python** - pyproject.toml, pip detection  
- **.NET** - .csproj files, NuGet packages
- **Java/Kotlin** - Maven/Gradle detection
- **Docker** - docker-compose.yml services
- **Terraform** - HCL file parsing
- **And more...**

#### 3. Rule System (`internal/rules/`)
- **700+ technology rules** covering enterprise stacks
- **YAML-based DSL** for easy extension
- **Multi-language support** (npm, pip, cargo, etc.)

#### 4. Language Detection (`github.com/go-enry/go-enry/v2`)
- **GitHub Linguist integration** for comprehensive language detection
- **1500+ languages** supported through open-source language database
- **Smart detection** by file extension and filename patterns
- **Handles special files** like Makefile, Dockerfile, etc.

#### 5. Parser System (`internal/scanner/parsers/`)
Specialized parsers for complex file formats:
- **HCL parser** for Terraform files
- **XML parser** for .csproj files
- **JSON parser** for package.json files
- **TOML parser** for pyproject.toml files

### Detection Pipeline

1. **File Discovery** - Recursive file system scanning
2. **Language Detection** - GitHub Linguist (go-enry) identification by extension and filename
3. **Component Detection** - Project-specific analysis
4. **Dependency Matching** - Pattern matching against rules
5. **Result Assembly** - Hierarchical payload construction

## How to Extend It

### Adding New Technology Rules

#### 1. Create a New Rule File

```yaml
# internal/rules/core/database/newtech.yaml
tech: newtech
name: New Technology
type: db
dotenv:
  - NEWTECH_
dependencies:
  - type: npm
    name: newtech-driver
    example: newtech-driver
  - type: python
    name: newtech-client
    example: newtech-client
files:
  - newtech.conf
detect:
  type: terraform
  file: "*.tf"
```

#### 2. Rule Categories

The rules are organized into 32 categories:

```
internal/rules/core/
├── ai/                   # AI/ML technologies
├── analytics/            # Analytics platforms
├── application/          # Application frameworks
├── automation/           # Automation tools
├── build/                # Build systems
├── ci/                   # CI/CD systems
├── cloud/                # Cloud providers
├── cms/                  # Content management systems
├── collaboration/        # Collaboration tools
├── communication/        # Communication services
├── crm/                  # CRM systems
├── database/             # Database systems
├── etl/                  # ETL tools
├── framework/            # Application frameworks
├── hosting/              # Hosting services
├── infrastructure/       # Infrastructure tools
├── language/             # Programming languages
├── monitoring/           # Monitoring and observability
├── network/              # Network tools
├── notification/         # Notification services
├── payment/              # Payment processors
├── queue/                # Message queues
├── runtime/              # Runtime environments
├── saas/                 # SaaS platforms
├── security/             # Security tools
├── ssg/                  # Static site generators
├── storage/              # Storage services
├── test/                 # Testing frameworks
├── tool/                 # Development tools
├── ui/                   # UI libraries and frameworks
└── validation/           # Validation libraries
```

### Adding New Component Detectors

#### 1. Create Detector Structure

```go
// internal/scanner/components/newtech/detector.go
package newtech

import (
    "tech-stack-analyzer/internal/scanner/components"
    "tech-stack-analyzer/internal/types"
)

type Detector struct{}

func (d *Detector) Name() string {
    return "newtech"
}

func (d *Detector) Detect(files []types.File, ...) []*types.Payload {
    // Implementation here
}

func init() {
    components.Register(&Detector{})
}
```

#### 2. Create Parser (if needed)

```go
// internal/scanner/parsers/newtech.go
package parsers

type NewTechParser struct{}

func (p *NewTechParser) ParseConfig(content string) NewTechConfig {
    // Parse configuration files
}
```

#### 3. Register in Scanner

```go
// internal/scanner/scanner.go
import (
    _ "tech-stack-analyzer/internal/scanner/components/newtech"
)
```

### Adding New File Matchers

```go
// internal/scanner/matchers/newmatcher.go
func registerNewMatcher() {
    components.RegisterFileMatcher(&matcher.FileMatcher{
        Tech:       "newtech",
        Extensions: []string{".newext"},
        Pattern:    regexp.MustCompile(`newtech\.config`),
    })
}
```

### Custom Rule Directories

> **Note**: External rules support is planned but not yet implemented. Currently, the scanner uses embedded rules only.

## How to Contribute

We welcome contributions! Please follow these guidelines:

### Getting Started

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Set up development environment**
   ```bash
   # Install dependencies
   go mod download
   
   # Install Task for automation (optional)
   # macOS
   brew install go-task
   
   # Or install directly with Go
   go install github.com/go-task/task/v3/cmd/task@latest
   
   # Install pre-commit hooks (recommended)
   task pre-commit:setup
   task pre-commit:install
   ```

### Development Workflow

```bash
# Make your changes
# ... (edit files)

# Run quality checks
task fct    # Format, Check, Test

# Run specific tests
go test -v ./internal/scanner/...

# Run benchmarks
go test -bench=. ./internal/scanner/...

# Build to verify
task build
```

### Pre-commit Hooks

The project uses [pre-commit](https://pre-commit.com/) to automatically run quality checks before commits and pushes:

**Setup:**
```bash
# Install pre-commit tool (one-time setup)
task pre-commit:setup

# Install git hooks (one-time per clone)
task pre-commit:install
```

**Behavior:**
- **On commit**: Runs `task fct` (format + check + test) on changed Go files
- **On push**: Runs `task fct` + race detection tests for extra safety

**Useful commands:**
```bash
# Run hooks manually on all files
task pre-commit:run

# Skip hooks for a specific commit (use sparingly)
git commit -m "msg" --no-verify

# Validate configuration
task pre-commit:validate

# Update hook versions
task pre-commit:update
```

### Contribution Types

#### 1. Bug Fixes
- Create an issue describing the bug
- Add tests that reproduce the issue
- Fix the bug with failing tests passing
- Ensure all existing tests still pass

#### 2. New Technology Rules
- Add rules in appropriate category directories
- Follow existing rule structure and naming
- Test against real projects using the technology
- Update documentation if needed

#### 3. New Component Detectors
- Follow existing detector patterns
- Add comprehensive tests
- Update architecture documentation
- Register in scanner initialization

#### 4. Performance Improvements
- Add benchmarks for performance measurement
- Ensure no regression in functionality
- Document performance gains

### Code Style Guidelines

- **Go Formatting**: Use `gofmt` and `goimports`
- **Linting**: Pass `golangci-lint` with no issues
- **Testing**: Maintain >90% test coverage
- **Documentation**: Add comments for public functions
- **Commit Messages**: Use conventional commit format
  ```
  feat: add new Oracle database detector
  fix: resolve memory leak in file processing
  docs: update README with installation instructions
  ```

### Submitting Changes

1. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create Pull Request**
   - Use descriptive title and description
   - Link to relevant issues
   - Include screenshots for UI changes
   - Add performance benchmarks for optimizations

3. **Code Review Process**
   - Address all review feedback
   - Ensure CI checks pass
   - Update documentation as needed

### Reporting Issues

- **Bug Reports**: Use issue template with reproduction steps
- **Feature Requests**: Describe use case and expected behavior
- **Performance Issues**: Include benchmarks and system specs

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

### Original Project
This is a Go re-implementation of [specfy/stack-analyser](https://github.com/specfy/stack-analyser) by the original author. The original TypeScript implementation provided the foundation and inspiration for this project.

### Extensions and Enhancements
This Go implementation provides practical improvements focused on deployment simplicity:

- **Zero Dependencies**: Single executable binary with no Node.js runtime or package management required
- **Extended Technology Support**: Added Java/Kotlin and .NET component detectors alongside existing Node.js, Python, Docker, Terraform, Ruby, Rust, PHP, Deno, and Go support
- **Enhanced Database Coverage**: Improved detection for Oracle, MongoDB, Redis, and other enterprise databases
- **Modular Architecture**: Clean component detector system for easier maintenance and extension
- **Comprehensive Rules**: 700+ technology rules across 32 categories covering modern enterprise stacks

### Contributors
Thank you to all contributors who help improve this project.

---

Built with Go - Delivering fast technology stack analysis for modern development teams.
