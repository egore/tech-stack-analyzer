package python

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/petrarca/tech-stack-analyzer/internal/scanner/components"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

// Detector implements Python component detection
type Detector struct{}

// Name returns the detector name
func (d *Detector) Name() string {
	return "python"
}

// Detect scans for Python projects (pyproject.toml)
func (d *Detector) Detect(files []types.File, currentPath, basePath string, provider types.Provider, depDetector components.DependencyDetector) []*types.Payload {
	var payloads []*types.Payload

	for _, file := range files {
		if file.Name != "pyproject.toml" {
			continue
		}

		// Read pyproject.toml
		content, err := provider.ReadFile(filepath.Join(currentPath, file.Name))
		if err != nil {
			continue
		}

		// Extract project name
		projectName := extractProjectName(string(content))
		if projectName == "" {
			continue
		}

		// Create payload with specific file path
		relativeFilePath, _ := filepath.Rel(basePath, filepath.Join(currentPath, file.Name))
		if relativeFilePath == "." {
			relativeFilePath = "/"
		} else {
			relativeFilePath = "/" + relativeFilePath
		}

		payload := types.NewPayloadWithPath(projectName, relativeFilePath)

		// Set tech field to python
		tech := "python"
		payload.Tech = &tech

		// Parse dependencies
		dependencies := parseDependencies(string(content))

		// Extract dependency names for tech matching
		var depNames []string
		for _, dep := range dependencies {
			depNames = append(depNames, dep.Name)
		}

		// Match dependencies against rules
		if len(dependencies) > 0 {
			matchedTechs := depDetector.MatchDependencies(depNames, "python")
			for tech, reasons := range matchedTechs {
				for _, reason := range reasons {
					payload.AddTech(tech, reason)
				}
			}

			payload.Dependencies = dependencies
		}

		// Detect license
		detectLicense(string(content), payload)

		payloads = append(payloads, payload)
	}

	return payloads
}

// extractProjectName extracts the project name from pyproject.toml
func extractProjectName(content string) string {
	lines := strings.Split(content, "\n")
	inProjectSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[project]" {
			inProjectSection = true
			continue
		}

		if strings.HasPrefix(line, "[") && line != "[project]" {
			inProjectSection = false
			continue
		}

		if inProjectSection && strings.HasPrefix(line, "name") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[1])
				name = strings.Trim(name, `"'`)
				if name != "" {
					return name
				}
			}
		}
	}

	return ""
}

// parseDependencies parses dependencies from pyproject.toml
func parseDependencies(content string) []types.Dependency {
	var dependencies []types.Dependency
	lineReg := regexp.MustCompile(`(^([a-zA-Z0-9._-]+)$|^([a-zA-Z0-9._-]+)(([>=]+)([0-9.]+)))`)

	lines := strings.Split(content, "\n")
	inProjectSection := false
	inDependenciesSection := false
	expectingDependencies := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[project]" {
			inProjectSection = true
			continue
		}

		if line == "[project.dependencies]" || line == "[tool.poetry.dependencies]" {
			inDependenciesSection = true
			continue
		}

		if strings.HasPrefix(line, "[") {
			inProjectSection = false
			inDependenciesSection = false
			expectingDependencies = false
			continue
		}

		if inProjectSection && strings.HasPrefix(line, "dependencies") {
			expectingDependencies = true
			continue
		}

		if (inDependenciesSection || expectingDependencies) && line != "" && !strings.HasPrefix(line, "#") {
			line = strings.TrimSpace(line)
			line = strings.TrimPrefix(line, `"`)
			line = strings.TrimSuffix(line, `",`)
			line = strings.TrimSuffix(line, `"`)

			match := lineReg.FindStringSubmatch(line)
			if match == nil {
				continue
			}

			name := match[2]
			if name == "" {
				name = match[3]
			}
			version := match[6]
			if version == "" {
				version = "latest"
			}

			if name != "" {
				dependencies = append(dependencies, types.Dependency{
					Type:    "python",
					Name:    name,
					Example: version,
				})
			}
		}
	}

	return dependencies
}

// detectLicense detects license from pyproject.toml
func detectLicense(content string, payload *types.Payload) {
	lines := strings.Split(content, "\n")
	inProjectSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[project]" {
			inProjectSection = true
			continue
		}

		if strings.HasPrefix(line, "[") && line != "[project]" {
			inProjectSection = false
			continue
		}

		if inProjectSection && strings.HasPrefix(line, "license") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				license := strings.TrimSpace(parts[1])
				license = strings.Trim(license, `"'`)
				if license != "" {
					payload.Licenses = append(payload.Licenses, license)
				}
			}
		}
	}
}

func init() {
	// Auto-register this detector
	components.Register(&Detector{})
}
