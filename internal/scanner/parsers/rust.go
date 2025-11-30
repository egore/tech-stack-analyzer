package parsers

import (
	"strings"

	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

// RustParser handles Rust-specific file parsing (Cargo.toml)
type RustParser struct{}

// NewRustParser creates a new Rust parser
func NewRustParser() *RustParser {
	return &RustParser{}
}

// Dependency represents different Cargo dependency formats
type Dependency struct {
	Version string
	Path    string
	Git     string
	Branch  string
	Rev     string
}

// CargoToml represents the structure of Cargo.toml ( the parts we need)
type CargoToml struct {
	Package struct {
		Name    string `toml:"name"`
		License string `toml:"license"`
	} `toml:"package"`
	Dependencies      map[string]interface{} `toml:"dependencies"`
	DevDependencies   map[string]interface{} `toml:"dev-dependencies"`
	BuildDependencies map[string]interface{} `toml:"build-dependencies"`
	WorkspaceDeps     map[string]interface{} `toml:"workspace.dependencies"`
}

// ParseCargoToml parses Cargo.toml and extracts project info and dependencies
func (p *RustParser) ParseCargoToml(content string) (string, string, []types.Dependency, bool) {
	lines := strings.Split(content, "\n")

	var projectName, license string
	var dependencies []types.Dependency
	hasPackage := false

	// Parse the TOML manually to avoid external dependencies
	var currentSection string
	var inDependencies bool
	var inDevDependencies bool
	var inBuildDependencies bool
	var inWorkspaceDeps bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.Trim(line, "[]")
			currentSection = section
			inDependencies = (section == "dependencies")
			inDevDependencies = (section == "dev-dependencies")
			inBuildDependencies = (section == "build-dependencies")
			inWorkspaceDeps = (section == "workspace.dependencies")
			continue
		}

		// Parse package section
		if currentSection == "package" {
			if strings.HasPrefix(line, "name") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					projectName = strings.Trim(strings.Trim(parts[1], " "), `"`)
					hasPackage = true
				}
			} else if strings.HasPrefix(line, "license") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					license = strings.Trim(strings.Trim(parts[1], " "), `"`)
				}
			}
		}

		// Parse dependencies
		if inDependencies || inDevDependencies || inBuildDependencies || inWorkspaceDeps {
			dep := p.parseDependencyLine(line)
			if dep.Name != "" {
				dependencies = append(dependencies, dep)
			}
		}
	}

	return projectName, license, dependencies, hasPackage
}

// parseDependencyLine parses a single dependency line from Cargo.toml
func (p *RustParser) parseDependencyLine(line string) types.Dependency {
	// Remove comments
	if idx := strings.Index(line, "#"); idx != -1 {
		line = strings.TrimSpace(line[:idx])
	}

	if line == "" {
		return types.Dependency{}
	}

	// Split on = to get name and value
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return types.Dependency{}
	}

	name := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Handle different dependency formats
	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		// Simple string version: "serde = "1.0""
		version := strings.Trim(value, `"`)
		return types.Dependency{
			Type:    "rust",
			Name:    name,
			Example: version,
		}
	} else if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		// Object format: "serde = { version = "1.0", features = ["derive"] }"
		return p.parseObjectDependency(name, value)
	}

	return types.Dependency{}
}

// parseObjectDependency parses object-style dependencies
func (p *RustParser) parseObjectDependency(name, value string) types.Dependency {
	// Remove braces and split by lines
	content := strings.Trim(value, "{}")
	lines := strings.Split(content, ",")

	depInfo := &dependencyInfo{}

	for _, line := range lines {
		depInfo.parseLine(line)
	}

	return p.buildDependency(name, depInfo)
}

// dependencyInfo holds parsed dependency information
type dependencyInfo struct {
	version, path, git, branch, rev string
}

// parseLine extracts dependency information from a single line
func (d *dependencyInfo) parseLine(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return
	}

	value := strings.Trim(strings.Trim(parts[1], " "), `"`)
	key := strings.TrimSpace(parts[0])

	switch key {
	case "version":
		d.version = value
	case "path":
		d.path = value
	case "git":
		d.git = value
	case "branch":
		d.branch = value
	case "rev":
		d.rev = value
	}
}

// buildDependency creates a dependency from parsed information
func (p *RustParser) buildDependency(name string, info *dependencyInfo) types.Dependency {
	var example string

	switch {
	case info.path != "":
		example = p.buildPathExample(info)
	case info.git != "":
		example = p.buildGitExample(info)
	default:
		example = info.version
		if example == "" {
			example = "latest"
		}
	}

	return types.Dependency{
		Type:    "rust",
		Name:    name,
		Example: example,
	}
}

// buildPathExample creates a path-based example string
func (p *RustParser) buildPathExample(info *dependencyInfo) string {
	example := "path:" + info.path
	if info.version != "" {
		example += ":" + info.version
	}
	return example
}

// buildGitExample creates a git-based example string
func (p *RustParser) buildGitExample(info *dependencyInfo) string {
	example := "git:" + info.git

	ref := info.branch
	if ref == "" {
		ref = info.rev
	}

	if ref != "" {
		example += "#" + ref
	} else {
		example += "#latest"
	}

	return example
}
