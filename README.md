# Tech Stack Analyzer

> **Work in Progress**  
> This project is under active development. The output format, API, and configuration structure may change. While the core functionality is stable, breaking changes may occur as we refine the functionalities and add new features. Use in production environments at your own discretion.

A technology stack analyzer written in Go, re-implementing [specfy/stack-analyser](https://github.com/specfy/stack-analyser) with improvements and extended technology support.

## What This Project Does

The Tech Stack Analyzer automatically detects technologies, frameworks, databases, and tools used in codebases by analyzing files, dependencies, and configurations. It provides comprehensive insights into:

- **Programming Languages** - Detects source code languages and versions
- **Package Managers** - Identifies npm, pip, cargo, composer, nuget, maven dependencies
- **Frameworks** - Detects .NET, Spring Boot, Angular, React, Django frameworks
- **Databases** - Identifies PostgreSQL, MySQL, MongoDB, Redis, Oracle, SQL Server
- **Infrastructure** - Detects Docker, Kubernetes, Terraform, GitLab configurations
- **DevOps Tools** - Identifies CI/CD pipelines, monitoring, and deployment tools

## Key Features

- **700+ Technology Rules** - Comprehensive detection across databases, frameworks, tools, cloud services
- **Zero Dependencies** - Single binary deployment without Node.js runtime requirement
- **Project Configuration** - `.stack-analyzer.yml` for custom metadata, exclusions, and external dependencies
- **Scan Metadata** - Automatic tracking of scan execution (timestamp, duration, git info, file counts)
- **Glob Pattern Exclusions** - Flexible `--exclude` flag supporting `**`, `*`, `?` patterns for files and directories
- **Content-Based Detection** - Validates technologies through regex pattern matching in file contents for precise identification
- **Configurable Components** - Override default component classification per rule with `is_component` field
- **Externalized Configuration** - Type definitions and ignore patterns in YAML (no code changes needed)
- **Tech-Specific Metadata** - Structured properties for Docker (base images, ports) and Terraform (providers, resource counts)
- **Multi-Technology Components** - Detects hybrid projects with multiple primary technologies in the same directory
- **Professional Logging** - Structured logging with multiple levels (trace/debug/info/warn/error) and JSON/text formats
- **Hierarchical Output** - Component-based analysis with parent-child relationships
- **Aggregated Views** - Rollup summaries for quick technology stack overviews

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

The analyzer uses a command-based interface powered by [Cobra](https://github.com/spf13/cobra):

```bash
# Get help
./bin/stack-analyzer --help
./bin/stack-analyzer scan --help
./bin/stack-analyzer info --help

# Scan current directory
./bin/stack-analyzer scan

# Scan specific directory
./bin/stack-analyzer scan /path/to/project

# Save results to file
./bin/stack-analyzer scan /path/to/project --output results.json

# Exclude specific directories and files (supports glob patterns)
./bin/stack-analyzer scan /path/to/project --exclude "vendor" --exclude "node_modules" --exclude "*.log"

# Scan a single file (useful for quick testing)
./bin/stack-analyzer scan /path/to/pom.xml
./bin/stack-analyzer scan /path/to/package.json
./bin/stack-analyzer scan /path/to/pyproject.toml

# Aggregate output (rollup technologies, languages, licenses, dependencies)
./bin/stack-analyzer scan --aggregate tech,techs,languages,licenses,dependencies /path/to/project
./bin/stack-analyzer scan --aggregate techs /path/to/project
./bin/stack-analyzer scan --aggregate dependencies /path/to/project

# List all available technologies
./bin/stack-analyzer info techs

# Show rule details for a specific technology
./bin/stack-analyzer info rule postgresql
./bin/stack-analyzer info rule postgresql --format json

# List component types
./bin/stack-analyzer info component-types
```

### Configuration & Logging

The scanner supports configuration through command-line flags and environment variables. Environment variables provide defaults that can be overridden by flags.

#### Environment Variables

```bash
# Output configuration
export STACK_ANALYZER_OUTPUT=/tmp/scan-results.json
export STACK_ANALYZER_PRETTY=false

# Scan behavior
export STACK_ANALYZER_EXCLUDE_DIRS=vendor,node_modules,build
export STACK_ANALYZER_AGGREGATE=tech,techs,languages

# Logging
export STACK_ANALYZER_LOG_LEVEL=debug      # trace, debug, info, warn, error, fatal, panic
export STACK_ANALYZER_LOG_FORMAT=json      # text or json
```

#### Logging

The scanner provides professional logging with multiple levels and formats:

```bash
# Debug logging with structured fields
./bin/stack-analyzer scan /path --log-level debug

# JSON logging for automated processing
./bin/stack-analyzer scan /path --log-format json

# Trace level for maximum detail
./bin/stack-analyzer scan /path --log-level trace

# Environment variables for logging
STACK_ANALYZER_LOG_LEVEL=debug STACK_ANALYZER_LOG_FORMAT=json \
  ./bin/stack-analyzer scan /path
```

**Log Output Examples:**

Text format:
```
INFO[2025-12-01 11:24:14] Starting Tech Stack Analyzer command=scan version=1.0.0
INFO[2025-12-01 11:24:14] Initializing scanner exclude_dirs="[]" max_depth=-1 path=/path
DEBU[2025-12-01 11:24:14] Generating output aggregate=tech pretty_print=true
```

JSON format:
```json
{"command":"scan","level":"info","msg":"Starting Tech Stack Analyzer","time":"2025-12-01 11:24:16","version":"1.0.0"}
{"path":"/path","level":"info","msg":"Initializing scanner","time":"2025-12-01 11:24:16"}
{"aggregate":"tech","level":"debug","msg":"Generating output","pretty_print":true,"time":"2025-12-01 11:24:16"}
```

### Commands

#### `scan` - Analyze a project or file

Scans a project directory or single file to detect technologies, frameworks, databases, and services.

**Usage:**
```bash
stack-analyzer scan [path] [flags]
```

**Flags:**
- `--output, -o` - Output file path (default: stdout)
- `--aggregate` - Aggregate fields: `tech,techs,languages,licenses,dependencies`
- `--exclude` - Patterns to exclude (supports glob patterns like `**/__tests__/**`, `*.log`; can be specified multiple times)
- `--pretty` - Pretty print JSON output (default: true)
- `--log-level` - Log level: trace, debug, info, warn, error, fatal, panic (default: info)
- `--log-format` - Log format: text or json (default: text)

**Examples:**
```bash
# Basic usage
stack-analyzer scan .
stack-analyzer scan /path/to/project --output results.json
stack-analyzer scan --aggregate techs,languages,dependencies /path

# Exclude patterns (glob support)
stack-analyzer scan /path --exclude vendor --exclude node_modules
stack-analyzer scan /path --exclude "**/__tests__/**" --exclude "*.log"

# Logging examples
stack-analyzer scan /path --log-level debug --log-format json
stack-analyzer scan /path --log-level trace
```

#### `info` - Display information about rules and types

**Subcommands:**

**`info component-types`** - List all component types
```bash
stack-analyzer info component-types
```
Shows which technology types create components (appear in `tech` field) vs those that don't (only in `techs` array).

**`info techs`** - List all available technologies
```bash
stack-analyzer info techs
stack-analyzer info techs | grep postgres
```
Lists all technology names from the embedded rules.

**`info rule [tech-name]`** - Show rule details
```bash
stack-analyzer info rule postgresql
stack-analyzer info rule postgresql --format json
```
Displays the complete rule definition for a given technology.

**Flags:**
- `--format, -f` - Output format: `yaml` or `json` (default: yaml)

### Global Flags

- `--help, -h` - Help for any command
- `--version, -v` - Show version information

### Detection Approach

The scanner uses a **two-tier detection system**:

**1. Universal Technology Detection** - Works with all technologies through pattern matching:
- Comprehensive technology rules covering databases, frameworks, tools, cloud services
- Package manager dependencies (npm, pip, cargo, composer, nuget, maven)
- Configuration files, environment variables, and file patterns
- Docker image references and file extensions

**2. Specialized Component Detectors** - Deep analysis for major stacks (Python, Node.js, Java/Kotlin, .NET, Docker, Terraform, Ruby, Rust, PHP, Deno, Go):
- Exact package versions and dependency relationships
- Configuration analysis and project structure
- Sub-projects, modules, and service identification
- Dependency graph construction

This ensures broad coverage while providing detailed insights for common technology stacks.

### Component Classification

The scanner distinguishes between **architectural components** and **tools/libraries**. Technologies like databases, hosting services, and SaaS platforms create components (appear in `tech` field), while development tools, frameworks, and languages are listed only in the `techs` array.

This classification is fully configurable through type definitions and per-rule overrides. See the [Technology Type Configuration](#technology-type-configuration) section for details.

### Content-Based Detection

In addition to file names and extensions, the scanner can validate technology detection through **content pattern matching**. This enables precise identification of libraries and frameworks that share file extensions.

#### How It Works

1. **Extension Pre-filtering**: Content matching only runs on files with matching extensions
2. **Pattern Validation**: If a rule has `content` patterns, ALL must be satisfied:
   - Extension matches (pre-filter)
   - At least one content pattern matches (validation)
3. **First Match Wins**: Stops after the first pattern matches per technology
4. **Efficient**: Only reads files when necessary, skips files without matching extensions

#### Rule Example

```yaml
tech: mfc
name: Microsoft Foundation Class Library
type: ui_framework
extensions: [.cpp, .h, .hpp]
content:
  - pattern: '#include\s+<afx'
  - pattern: 'class\s+\w+\s*:\s*public\s+C(Wnd|FrameWnd|CDialog|...)'
  - pattern: '(BEGIN_MESSAGE_MAP|END_MESSAGE_MAP|DECLARE_MESSAGE_MAP)'
```

**Behavior:**
- `.cpp` files are checked (extension matches)
- If `#include <afx` is found → MFC detected (first pattern matched, validation passed)
- If no patterns match → MFC not detected (validation failed, tech removed)
- Rules without `content` → detected by extension alone (existing behavior)

#### Use Cases

- **Distinguish similar technologies**: C++ STL vs plain C in `.h` files
- **Library-specific detection**: MFC, Qt, Boost through include patterns
- **Framework patterns**: React hooks, Vue composition API through code signatures
- **Prevent false positives**: Only detect when actual usage is confirmed

### Technology Type Configuration

Technology types and their component behavior are defined in `internal/config/types.yaml`. This configuration file determines which technology types create architectural components versus being classified as tools/libraries.

#### Type Configuration File

```yaml
# internal/config/types.yaml
types:
  database:
    is_component: true
    description: "Database systems (PostgreSQL, MongoDB, Redis, etc.)"
  
  framework:
    is_component: false
    description: "Application frameworks (React, Django, Spring, etc.)"
```

**Adding New Technology Types:**

1. Add the type definition to `internal/config/types.yaml`:
   ```yaml
   my_new_type:
     is_component: true  # or false
     description: "Description of this type"
   ```

2. Use the type in your rules:
   ```yaml
   tech: my-tech
   name: My Technology
   type: my_new_type
   ```

**Benefits:**
- **No code changes required** - Edit YAML, no recompilation needed
- **Self-documenting** - Descriptions explain each type's purpose
- **Centralized** - All type definitions in one place
- **Discoverable** - Use `stack-analyzer info component-types` to list all types

#### Per-Rule Component Override

Individual rules can override the type's default behavior using the `is_component` field:

```yaml
tech: mfc
type: ui_framework  # Default: is_component: false
is_component: true  # Override: create component anyway
```

**Priority Order:**
1. Rule's `is_component` field (highest priority)
2. Type definition in `types.yaml`
3. Default to `false` if type not defined

**Example Use Cases:**
- **Promote to component**: `ui_framework` with `is_component: true` creates a component
- **Demote from component**: `database` with `is_component: false` doesn't create a component
- **New types**: Types not in `types.yaml` default to no component creation

This configuration-driven approach allows fine-grained control over which technologies appear as architectural components versus implementation details.

## Project Configuration

### `.stack-analyzer.yml` Configuration File

Place a `.stack-analyzer.yml` file in your project root to customize scan behavior, add metadata, and document external dependencies.

```yaml
# .stack-analyzer.yml - Tech Stack Analyzer Configuration

# Custom properties added to metadata.properties in scan output
properties:
  product: "My Product Name"
  team: "Platform Engineering"
  environment: "production"
  owner: "engineering@company.com"

# Files and directories to exclude from scanning
# Supports glob patterns (**, *, ?)
exclude:
  - "node_modules"
  - "vendor"
  - "*.log"
  - "**/__tests__/**"
  - "**/*.test.js"

# Technologies to add to scan results (even if not auto-detected)
techs:
  - tech: "aws"
    reason: "Deployed on AWS ECS"
  - tech: "datadog"
    reason: "Monitoring via Datadog"
```

**Configuration Options:**

- **`properties`** - Custom metadata added to `metadata.properties` in output
  - Document product context, ownership, deployment information
  - Any key-value pairs relevant to your project
  
- **`exclude`** - Patterns to exclude from scanning
  - Supports glob patterns: `**`, `*`, `?`
  - Matches files and directories
  - Merged with CLI `--exclude` flags
  
- **`techs`** - Technologies to force-add to scan results
  - Useful for external dependencies (AWS, SaaS services)
  - Each tech has optional `reason` field
  - Added to root payload's `techs` array

**Example Output with Configuration:**

```json
{
  "metadata": {
    "timestamp": "2025-12-01T14:45:35Z",
    "scan_path": "/path/to/project",
    "properties": {
      "product": "My Product Name",
      "team": "Platform Engineering",
      "environment": "production"
    }
  },
  "techs": ["nodejs", "aws", "datadog"],
  "reason": [
    "matched file: package.json",
    "Deployed on AWS ECS",
    "Monitoring via Datadog"
  ]
}
```

**Benefits:**
- **Version controlled** - Configuration lives with code
- **Team-shared** - Everyone uses same exclusions and metadata
- **Documented** - External dependencies explicitly listed
- **Flexible** - Custom metadata for any use case

See `.stack-analyzer.yml.example` for a complete configuration template.

### Output Structure

The scanner outputs a hierarchical JSON structure representing the detected technologies:

- **id**: Unique identifier for each component
- **name**: Component name (e.g., "main", "frontend", "backend")
- **path**: File system path relative to the project root
- **tech**: Array of primary technologies for this component (e.g., `["nodejs", "java"]` for hybrid projects)
- **techs**: Array of all technologies detected in this component (components + tools/libraries)
- **languages**: Object mapping programming languages to file counts
- **dependencies**: Array of detected dependencies with format `[type, name, version]`
- **childs**: Array of nested components (sub-projects, services, etc.)
- **edges**: Array of relationships between components (e.g., service → database connections); created for architectural components like databases, SaaS services, and monitoring tools, but not for hosting/cloud providers
- **inComponent**: Reference to parent component if this is a nested component
- **licenses**: Array of detected licenses in this component
- **reason**: Array explaining why technologies were detected
- **properties**: Object containing tech-specific metadata (Docker, Terraform, Kubernetes, etc.)
- **metadata**: Scan execution metadata (only in root payload)

#### Metadata Field

The `metadata` field (present only in the root payload) provides information about the scan execution:

```json
{
  "metadata": {
    "timestamp": "2025-12-01T14:45:35Z",
    "scan_path": "/absolute/path/to/project",
    "scanner_version": "1.0.0",
    "duration_ms": 1173,
    "file_count": 523,
    "directory_count": 87,
    "excluded_dirs": ["node_modules", "vendor"],
    "git": {
      "branch": "main",
      "commit": "a1b2c3d",
      "is_dirty": true,
      "remote_url": "https://github.com/user/repo.git"
    },
    "properties": {
      "product": "My Product",
      "team": "Engineering"
    }
  }
}
```

**Metadata Fields:**
- **timestamp**: ISO 8601 timestamp when scan was performed
- **scan_path**: Absolute path to scanned directory
- **scanner_version**: Version of the analyzer
- **duration_ms**: Scan duration in milliseconds
- **file_count**: Total files scanned
- **directory_count**: Total directories traversed
- **excluded_dirs**: List of excluded patterns
- **git**: Git repository information (if in a git repo)
  - **branch**: Current branch name
  - **commit**: Short commit hash (7 characters)
  - **is_dirty**: Whether there are uncommitted changes
  - **remote_url**: Origin remote URL
- **properties**: Custom properties from `.stack-analyzer.yml`

#### Properties Field

The `properties` field provides structured metadata about specific technologies detected in the project. This field uses an industry-standard format compatible with JSON Schema, OpenAPI, and SBOM tools.

**Supported Technologies:**

**Docker** - Extracts information from Dockerfiles:
```json
"properties": {
  "docker": [
    {
      "file": "/backend/Dockerfile",
      "base_images": ["python:3.13", "python:3.13-slim"],
      "exposed_ports": [8080],
      "multi_stage": true,
      "stages": ["builder"]
    },
    {
      "file": "/frontend/Dockerfile",
      "base_images": ["node:20-alpine", "nginx:alpine"],
      "exposed_ports": [80],
      "multi_stage": true,
      "stages": ["builder"]
    }
  ]
}
```

**Terraform** - Aggregates infrastructure resources:
```json
"properties": {
  "terraform": [
    {
      "file": "/infrastructure/main.tf",
      "providers": ["aws", "google"],
      "resources_by_provider": {
        "aws": 15,
        "google": 3
      },
      "resources_by_category": {
        "compute": 5,
        "storage": 8,
        "database": 3,
        "networking": 2
      },
      "total_resources": 18
    }
  ]
}
```

**Key Features:**
- **Array format**: Supports multiple files (multiple Dockerfiles, .tf files, etc.)
- **File tracking**: Each entry includes the source file path
- **Component-scoped**: Properties can appear at root or in child components
- **Tool-friendly**: Compatible with security scanners, SBOM generators, and CI/CD tools

#### Multi-Technology Components

When multiple technology stacks are detected in the same directory (e.g., a directory with both `package.json` and `pom.xml`), the scanner automatically merges them into a single component with multiple primary technologies. This accurately represents hybrid projects that combine different technology stacks:

```json
{
  "name": "hybrid-service",
  "tech": ["nodejs", "java"],
  "techs": ["nodejs", "java", "maven", "npm", "typescript"],
  "languages": {
    "TypeScript": 150,
    "Java": 45
  }
}
```

This is common in projects with:
- Node.js frontend + Java backend in the same module
- Integration tests (Playwright/TypeScript) alongside Java applications
- Build tools from multiple ecosystems

### Example Full Output

```json
{
  "id": "abc123",
  "name": "main",
  "path": ["/"],
  "tech": ["nodejs"],
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
      "tech": ["nodejs"],
      "dependencies": [["npm", "react", "^18.2.0"]]
    }
  ]
}
```

### Aggregated Output

Use the `--aggregate` flag to get a simplified, rolled-up view of your entire codebase:

```bash
./bin/stack-analyzer scan --aggregate tech,techs,languages,licenses,dependencies /path/to/project
```

**Output:**
```json
{
  "tech": ["nodejs", "python", "postgresql", "redis"],
  "techs": ["nodejs", "python", "postgresql", "redis", "react", "typescript", "docker", "eslint", "prettier"],
  "languages": {
    "Python": 130,
    "TypeScript": 89,
    "JavaScript": 45,
    "Go": 12
  },
  "licenses": ["MIT", "Apache-2.0"],
  "dependencies": [
    ["npm", "react", "^18.2.0"],
    ["npm", "express", "^4.18.0"],
    ["python", "fastapi", "0.118.2"],
    ["python", "pydantic", "latest"]
  ]
}
```

**Available fields:**
- `tech` - Primary technologies
- `techs` - All detected technologies (includes frameworks, tools, libraries)
- `languages` - Programming languages with file counts
- `licenses` - Detected licenses from LICENSE files and package manifests
- `dependencies` - All dependencies as `[type, name, version]` arrays

This is useful for:
- Quick technology stack overview
- Generating technology badges
- Dependency auditing and security scanning
- License compliance checking
- Counting dependencies: `jq '.dependencies | length'`

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
│   │   ├── components/    # Component detectors
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
- **Detection** by file extension and filename patterns
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

The rules are organized into 30+ categories:

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
- **Comprehensive Rules**: 700+ technology rules across 30+ categories covering modern enterprise stacks

### Contributors
Thank you to all contributors who help improve this project.

---

Built with Go - Delivering technology stack analysis for modern development teams.
