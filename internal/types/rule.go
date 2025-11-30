package types

import (
	"encoding/json"
	"regexp"
)

// Rule represents a technology detection rule
type Rule struct {
	Tech         string        `yaml:"tech" json:"tech"`
	Name         string        `yaml:"name" json:"name"`
	Type         string        `yaml:"type" json:"type"`
	DotEnv       []string      `yaml:"dotenv,omitempty" json:"dotenv,omitempty"`
	Dependencies []Dependency  `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Files        []string      `yaml:"files,omitempty" json:"files,omitempty"`
	Extensions   []string      `yaml:"extensions,omitempty" json:"extensions,omitempty"`
	Detect       *DetectConfig `yaml:"detect,omitempty" json:"detect,omitempty"`
}

// Dependency represents a dependency pattern (struct for YAML, but marshals as array for JSON)
type Dependency struct {
	Type    string `yaml:"type" json:"type"`
	Name    string `yaml:"name" json:"name"`
	Example string `yaml:"example,omitempty" json:"example,omitempty"`
}

// MarshalJSON converts Dependency struct to array format [type, name, version] to match TypeScript
func (d Dependency) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{d.Type, d.Name, d.Example})
}

// CompiledDependency is a pre-compiled dependency for performance
type CompiledDependency struct {
	Regex *regexp.Regexp
	Tech  string
	Name  string
	Type  string
}

// DetectConfig represents custom detection configuration
type DetectConfig struct {
	Type    string `yaml:"type" json:"type"` // json-schema, regex, yaml-path, package-json
	File    string `yaml:"file" json:"file"`
	Schema  string `yaml:"schema,omitempty" json:"schema,omitempty"`
	Pattern string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Path    string `yaml:"path,omitempty" json:"path,omitempty"`
	Extract bool   `yaml:"extract,omitempty" json:"extract,omitempty"`
}

// Compile compiles a dependency pattern to regex for performance
func (d *Dependency) Compile() (*CompiledDependency, error) {
	pattern := d.Name

	// Check if it's a regex pattern
	if len(pattern) > 2 && pattern[0] == '/' && pattern[len(pattern)-1] == '/' {
		regex, err := regexp.Compile(pattern[1 : len(pattern)-1])
		if err != nil {
			return nil, err
		}
		return &CompiledDependency{
			Regex: regex,
			Tech:  "", // Will be set by rule
			Name:  d.Name,
			Type:  d.Type,
		}, nil
	}

	// Simple string match - compile to exact regex
	regex := regexp.MustCompile("^" + regexp.QuoteMeta(pattern) + "$")
	return &CompiledDependency{
		Regex: regex,
		Tech:  "", // Will be set by rule
		Name:  d.Name,
		Type:  d.Type,
	}, nil
}

// Match checks if the dependency matches the given string
func (cd *CompiledDependency) Match(s string) bool {
	return cd.Regex.MatchString(s)
}
