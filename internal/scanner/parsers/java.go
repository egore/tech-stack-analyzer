package parsers

import (
	"encoding/xml"
	"regexp"
	"strings"

	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

// JavaParser handles Java/Kotlin-specific file parsing (pom.xml, build.gradle, build.gradle.kts)
type JavaParser struct{}

// NewJavaParser creates a new Java parser
func NewJavaParser() *JavaParser {
	return &JavaParser{}
}

// ParsePomXML parses pom.xml and extracts Maven dependencies
func (p *JavaParser) ParsePomXML(content string) []types.Dependency {
	var dependencies []types.Dependency

	type MavenDependency struct {
		GroupId    string `xml:"groupId"`
		ArtifactId string `xml:"artifactId"`
		Version    string `xml:"version"`
	}

	type MavenProject struct {
		Dependencies struct {
			Dependencies []MavenDependency `xml:"dependency"`
		} `xml:"dependencies"`
	}

	var mavenProject MavenProject
	if err := xml.Unmarshal([]byte(content), &mavenProject); err != nil {
		return dependencies
	}

	for _, dep := range mavenProject.Dependencies.Dependencies {
		if dep.GroupId != "" && dep.ArtifactId != "" {
			dependencyName := dep.GroupId + ":" + dep.ArtifactId
			version := dep.Version
			if version == "" {
				version = "latest"
			}

			dependencies = append(dependencies, types.Dependency{
				Type:    "maven",
				Name:    dependencyName,
				Example: version,
			})
		}
	}

	return dependencies
}

// ParseGradle parses build.gradle or build.gradle.kts and extracts Gradle dependencies
func (p *JavaParser) ParseGradle(content string) []types.Dependency {
	var dependencies []types.Dependency

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if p.shouldSkipLine(line) {
			continue
		}

		// Quick validation - is this even a dependency line?
		if !p.isPotentialDependencyLine(line) {
			continue
		}

		gradleDep := p.parseGradleDependency(line)
		if gradleDep != nil {
			dependencies = append(dependencies, *gradleDep)
		}
	}

	return dependencies
}

// GradleDependency represents a parsed Gradle dependency
type GradleDependency struct {
	Type     string
	Group    string
	Artifact string
	Version  string
}

// shouldSkipLine checks if a line should be skipped during parsing
func (p *JavaParser) shouldSkipLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*")
}

// isPotentialDependencyLine does quick validation before expensive regex matching
func (p *JavaParser) isPotentialDependencyLine(line string) bool {
	// Must contain a dependency type and quoted content with colon
	hasDepType := strings.Contains(line, "implementation") ||
		strings.Contains(line, "compile") ||
		strings.Contains(line, "api") ||
		strings.Contains(line, "runtimeOnly") ||
		strings.Contains(line, "compileOnly") ||
		strings.Contains(line, "annotationProcessor") ||
		strings.Contains(line, "testImplementation") ||
		strings.Contains(line, "testRuntimeOnly")

	hasQuotedContent := (strings.Contains(line, "'") || strings.Contains(line, `"`)) && strings.Contains(line, ":")

	return hasDepType && hasQuotedContent
}

// parseGradleDependency parses a single Gradle dependency line
func (p *JavaParser) parseGradleDependency(line string) *types.Dependency {
	// Supported dependency types
	depTypes := []string{
		"implementation", "compile", "testImplementation", "api",
		"compileOnly", "runtimeOnly", "testRuntimeOnly", "annotationProcessor",
	}

	// Extract dependency type
	depTypeRegex := regexp.MustCompile(`^\s*(` + strings.Join(depTypes, "|") + `)`)
	depTypeMatch := depTypeRegex.FindStringSubmatch(line)
	if len(depTypeMatch) < 2 {
		return nil
	}

	// Extract the quoted dependency string
	quotedRegex := regexp.MustCompile(`['"]([^'"]+)['"]`)
	quotedMatch := quotedRegex.FindStringSubmatch(line)
	if len(quotedMatch) < 2 {
		return nil
	}

	// Parse the dependency parts
	depString := quotedMatch[1]
	parts := strings.Split(depString, ":")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return nil
	}

	group := parts[0]
	artifact := parts[1]
	version := "latest"

	if len(parts) >= 3 && parts[2] != "" {
		version = parts[2]
	}

	dependencyName := group + ":" + artifact

	return &types.Dependency{
		Type:    "gradle",
		Name:    dependencyName,
		Example: version,
	}
}
