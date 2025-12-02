package scanner

import "github.com/petrarca/tech-stack-analyzer/internal/types"

// ComponentTypes defines which technology types should create components vs just be listed as dependencies
// This classification determines whether a detected technology appears in the 'tech' field (primary technologies)
// or only in the 'techs' array (all technologies including tools/libraries)

var typesConfig *types.TypesConfig

// SetTypesConfig sets the global types configuration
func SetTypesConfig(config *types.TypesConfig) {
	typesConfig = config
}

// ShouldCreateComponent determines if a rule should create a component
// Returns true if component should be created, false otherwise
func ShouldCreateComponent(rule types.Rule) bool {
	// Priority 1: If is_component is explicitly set in rule, use that value
	if rule.IsComponent != nil {
		return *rule.IsComponent
	}

	// Priority 2: Check types configuration from _types.yaml
	if typesConfig != nil {
		if typeDef, exists := typesConfig.Types[rule.Type]; exists {
			return typeDef.IsComponent
		}
	}

	// Default: If type not found in _types.yaml or config not loaded, default to false
	// This ensures all types must be explicitly defined in _types.yaml
	return false
}

// ShouldAddPrimaryTech determines if a rule should add primary tech when component is created
// Returns true if primary tech should be added, false otherwise
func ShouldAddPrimaryTech(rule types.Rule) bool {
	// Priority 1: If is_primary_tech is explicitly set in rule, use that value
	if rule.IsPrimaryTech != nil {
		return *rule.IsPrimaryTech
	}

	// Priority 2: Use current logic - if component is created, add primary tech
	return ShouldCreateComponent(rule)
}
