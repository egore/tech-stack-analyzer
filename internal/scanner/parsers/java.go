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

	// Pattern for implementation 'group:artifact:version'
	// Pattern for implementation("group:artifact:version")
	// Pattern for compile 'group:artifact:version'
	// Pattern for testImplementation 'group:artifact:version'
	depRegex := regexp.MustCompile(`(?:implementation|compile|testImplementation|api|compileOnly|runtimeOnly|testRuntimeOnly)\s*\(?\s*['"]([^:]+):([^:]+)(?::([^'"]+))?['"]\s*\)?`)

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
			continue
		}

		matches := depRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		group := matches[1]
		artifact := matches[2]
		version := matches[3]

		if group != "" && artifact != "" {
			dependencyName := group + ":" + artifact
			if version == "" {
				version = "latest"
			}

			dependencies = append(dependencies, types.Dependency{
				Type:    "gradle",
				Name:    dependencyName,
				Example: version,
			})
		}
	}

	return dependencies
}
