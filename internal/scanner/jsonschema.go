package scanner

import (
	"encoding/json"
	"path/filepath"

	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

// JSONSchemaDetector handles JSON schema-based detection (e.g., Shadcn)
type JSONSchemaDetector struct {
	provider types.Provider
	rules    []types.Rule
}

// NewJSONSchemaDetector creates a new JSON schema detector
func NewJSONSchemaDetector(provider types.Provider, rules []types.Rule) *JSONSchemaDetector {
	return &JSONSchemaDetector{
		provider: provider,
		rules:    rules,
	}
}

// DetectJSONSchemaComponents detects components from JSON schema files
func (d *JSONSchemaDetector) DetectJSONSchemaComponents(files []types.File, currentPath string, basePath string) []*types.Payload {
	var payloads []*types.Payload

	// Find all rules that have JSON schema detection
	for _, rule := range d.rules {
		if rule.Detect == nil || rule.Detect.Type != "json-schema" {
			continue
		}

		// Look for the specified file
		for _, file := range files {
			if file.Name != rule.Detect.File {
				continue
			}

			// Read and parse the JSON file
			content, err := d.provider.ReadFile(filepath.Join(currentPath, file.Name))
			if err != nil {
				continue
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(content, &parsed); err != nil {
				continue
			}

			// Check if $schema field exists and matches
			schema, ok := parsed["$schema"]
			if !ok {
				continue
			}

			schemaStr, ok := schema.(string)
			if !ok {
				continue
			}

			// Validate schema matches expected value
			if schemaStr == rule.Detect.Schema {
				// Create a virtual payload for this detection with relative path
				relativePath, _ := filepath.Rel(basePath, filepath.Join(currentPath, file.Name))
				if relativePath == "." {
					relativePath = "/"
				} else {
					relativePath = "/" + relativePath
				}
				payload := types.NewPayloadWithPath("virtual", relativePath)
				payload.AddTech(rule.Tech, "matched JSON schema: "+rule.Detect.Schema)
				payloads = append(payloads, payload)
			}
		}
	}

	return payloads
}
