package aggregator

import (
	"sort"

	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

// AggregateOutput represents aggregated/rolled-up data from the scan
type AggregateOutput struct {
	Tech      []string       `json:"tech,omitempty"`  // Primary/main technologies
	Techs     []string       `json:"techs,omitempty"` // All detected technologies
	Languages map[string]int `json:"languages,omitempty"`
	Licenses  []string       `json:"licenses,omitempty"`
}

// Aggregator handles aggregation of scan results
type Aggregator struct {
	fields map[string]bool
}

// NewAggregator creates a new aggregator with specified fields
func NewAggregator(fields []string) *Aggregator {
	fieldMap := make(map[string]bool)
	for _, field := range fields {
		fieldMap[field] = true
	}
	return &Aggregator{
		fields: fieldMap,
	}
}

// Aggregate processes a payload and returns aggregated data
func (a *Aggregator) Aggregate(payload *types.Payload) *AggregateOutput {
	output := &AggregateOutput{}

	if a.fields["tech"] {
		output.Tech = a.collectPrimaryTechs(payload)
	}

	if a.fields["techs"] {
		output.Techs = a.collectTechs(payload)
	}

	if a.fields["languages"] {
		output.Languages = a.collectLanguages(payload)
	}

	if a.fields["licenses"] {
		output.Licenses = a.collectLicenses(payload)
	}

	return output
}

// collectPrimaryTechs recursively collects all unique primary techs (tech field) from payload and children
func (a *Aggregator) collectPrimaryTechs(payload *types.Payload) []string {
	techSet := make(map[string]bool)
	a.collectPrimaryTechsRecursive(payload, techSet)

	// Convert to sorted slice
	techs := make([]string, 0, len(techSet))
	for tech := range techSet {
		techs = append(techs, tech)
	}

	return sortStrings(techs)
}

// collectPrimaryTechsRecursive helper function
func (a *Aggregator) collectPrimaryTechsRecursive(payload *types.Payload, techSet map[string]bool) {
	// Add primary tech from current payload (if set)
	if payload.Tech != nil && *payload.Tech != "" {
		techSet[*payload.Tech] = true
	}

	// Recursively process children
	for _, child := range payload.Childs {
		a.collectPrimaryTechsRecursive(child, techSet)
	}
}

// collectTechs recursively collects all unique techs from payload and children
func (a *Aggregator) collectTechs(payload *types.Payload) []string {
	techSet := make(map[string]bool)
	a.collectTechsRecursive(payload, techSet)

	// Convert to sorted slice
	techs := make([]string, 0, len(techSet))
	for tech := range techSet {
		techs = append(techs, tech)
	}

	// Sort for consistent output
	return sortStrings(techs)
}

// collectTechsRecursive helper function
func (a *Aggregator) collectTechsRecursive(payload *types.Payload, techSet map[string]bool) {
	// Add techs from current payload
	for _, tech := range payload.Techs {
		techSet[tech] = true
	}

	// Recursively process children
	for _, child := range payload.Childs {
		a.collectTechsRecursive(child, techSet)
	}
}

// collectLanguages recursively collects and sums all languages
func (a *Aggregator) collectLanguages(payload *types.Payload) map[string]int {
	languages := make(map[string]int)
	a.collectLanguagesRecursive(payload, languages)
	return languages
}

// collectLanguagesRecursive helper function
func (a *Aggregator) collectLanguagesRecursive(payload *types.Payload, languages map[string]int) {
	// Add languages from current payload
	for lang, count := range payload.Languages {
		languages[lang] += count
	}

	// Recursively process children
	for _, child := range payload.Childs {
		a.collectLanguagesRecursive(child, languages)
	}
}

// collectLicenses recursively collects all unique licenses
func (a *Aggregator) collectLicenses(payload *types.Payload) []string {
	licenseSet := make(map[string]bool)
	a.collectLicensesRecursive(payload, licenseSet)

	// Convert to sorted slice
	licenses := make([]string, 0, len(licenseSet))
	for license := range licenseSet {
		licenses = append(licenses, license)
	}

	return sortStrings(licenses)
}

// collectLicensesRecursive helper function
func (a *Aggregator) collectLicensesRecursive(payload *types.Payload, licenseSet map[string]bool) {
	// Add licenses from current payload
	for _, license := range payload.Licenses {
		licenseSet[license] = true
	}

	// Recursively process children
	for _, child := range payload.Childs {
		a.collectLicensesRecursive(child, licenseSet)
	}
}

// sortStrings sorts a slice of strings in place and returns it
func sortStrings(s []string) []string {
	sort.Strings(s)
	return s
}
