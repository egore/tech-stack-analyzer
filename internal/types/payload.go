package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

// Payload represents the analysis result for a directory or component
type Payload struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Path         []string       `json:"path"`
	Tech         *string        `json:"tech"`
	Techs        []string       `json:"techs"`
	Languages    map[string]int `json:"languages"`
	Dependencies []Dependency   `json:"dependencies"`
	Childs       []*Payload     `json:"childs"` // Changed from Children to Childs
	Edges        []Edge         `json:"edges"`
	InComponent  *Payload       `json:"inComponent"` // Added missing field
	Licenses     []string       `json:"licenses"`    // Added missing field
	Reason       []string       `json:"reason"`
}

// Edge represents a relationship between components
type Edge struct {
	Target *Payload `json:"target"`
	Read   bool     `json:"read"`
	Write  bool     `json:"write"`
}

// MarshalJSON customizes Edge JSON serialization to match TypeScript (target as ID string)
func (e Edge) MarshalJSON() ([]byte, error) {
	// Like TypeScript: serialize target as just the ID string
	targetID := ""
	if e.Target != nil {
		targetID = e.Target.ID
	}

	// Create a map to match the exact TypeScript format
	edgeMap := map[string]interface{}{
		"target": targetID,
		"read":   e.Read,
		"write":  e.Write,
	}

	return json.Marshal(edgeMap)
}

// NewPayload creates a new payload
func NewPayload(name string, paths []string) *Payload {
	return &Payload{
		ID:           GenerateID(), // Generate unique ID like TypeScript nid()
		Name:         name,
		Path:         paths, // Store paths as array (like TypeScript Set)
		Techs:        make([]string, 0),
		Languages:    make(map[string]int),
		Dependencies: make([]Dependency, 0),
		Childs:       make([]*Payload, 0),
		Edges:        make([]Edge, 0),
		Licenses:     make([]string, 0),
		Reason:       make([]string, 0),
	}
}

// NewPayloadWithPath creates a new payload with a single path (convenience function)
func NewPayloadWithPath(name, path string) *Payload {
	return NewPayload(name, []string{path})
}

// AddChild adds a child payload with deduplication (following TypeScript's addChild logic exactly)
func (p *Payload) AddChild(service *Payload) *Payload {
	// Check for existing component to merge (like TypeScript lines 130-138)
	var exist *Payload
	for _, child := range p.Childs {
		// we only merge if a tech is similar otherwise it's too easy to get a false-positive
		if (child.Tech == nil || *child.Tech == "") && (service.Tech == nil || *service.Tech == "") {
			continue
		}
		if child.Name == service.Name {
			exist = child
			break
		}
		// REMOVED: Don't merge by technology type, only by name
		// This was causing all Node.js components to be merged together
	}

	// Handle hosting/cloud techs with dots (like TypeScript lines 140-153)
	// TODO: Implement hosting/cloud tech logic with listIndexed
	// For now, skip this complex hosting logic
	_ = service.Tech // Suppress unused warning

	if exist != nil {
		// Merge with existing component (like TypeScript lines 155-175)
		// Log all paths where it was found
		for _, path := range service.Path {
			exist.AddPath(path)
		}

		// Update edges to point to the initial component (simplified for Go)
		// This would need parent reference which we don't track in edges

		// Merge dependencies
		for _, dep := range service.Dependencies {
			exist.AddDependency(dep)
		}

		return exist
	}

	// Add new child if no duplicate found
	p.Childs = append(p.Childs, service)
	return service
}

// AddPath adds a path to the payload (like TypeScript Set.add)
func (p *Payload) AddPath(path string) {
	// Check for duplicate (like TypeScript Set behavior)
	for _, existing := range p.Path {
		if existing == path {
			return // Already exists, don't add duplicate
		}
	}
	p.Path = append(p.Path, path)
}

// AddLanguageWithCount increments the count for a language
func (p *Payload) AddLanguageWithCount(language string, count int) {
	p.Languages[language] += count
}

// Combine merges another payload into this one (following TypeScript's combine method)
func (p *Payload) Combine(other *Payload) {
	p.mergePaths(other.Path)
	p.mergeLanguages(other.Languages)
	p.mergeTechs(other.Techs)
	p.mergeTechField(other.Tech)
	p.mergeDependencies(other.Dependencies)
	p.mergeLicenses(other.Licenses)
}

// Helper functions to reduce cognitive complexity

func (p *Payload) mergePaths(paths []string) {
	for _, path := range paths {
		if !p.containsString(p.Path, path) {
			p.Path = append(p.Path, path)
		}
	}
}

func (p *Payload) mergeLanguages(languages map[string]int) {
	for lang, count := range languages {
		p.Languages[lang] += count
	}
}

func (p *Payload) mergeTechs(techs []string) {
	for _, tech := range techs {
		if !p.containsString(p.Techs, tech) {
			p.Techs = append(p.Techs, tech)
		}
	}
}

func (p *Payload) mergeTechField(tech *string) {
	if tech != nil && *tech != "" && !p.containsString(p.Techs, *tech) {
		p.Techs = append(p.Techs, *tech)
	}
}

func (p *Payload) mergeDependencies(deps []Dependency) {
	for _, dep := range deps {
		if !p.containsDependency(dep) {
			p.Dependencies = append(p.Dependencies, dep)
		}
	}
}

func (p *Payload) mergeLicenses(licenses []string) {
	for _, license := range licenses {
		if !p.containsString(p.Licenses, license) {
			p.Licenses = append(p.Licenses, license)
		}
	}
}

func (p *Payload) containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func (p *Payload) containsDependency(dep Dependency) bool {
	for _, existing := range p.Dependencies {
		if existing.Type == dep.Type && existing.Name == dep.Name && existing.Example == dep.Example {
			return true
		}
	}
	return false
}

// AddTech adds a technology to the payload
func (p *Payload) AddTech(tech string, reason string) {
	// Avoid duplicates for techs, but still add reasons
	techExists := false
	for _, existing := range p.Techs {
		if existing == tech {
			techExists = true
			break
		}
	}

	if !techExists {
		p.Techs = append(p.Techs, tech)
		// NOTE: Don't set primary tech here like the original
		// The original's addTech method only adds to techs set, doesn't set this.tech
	}

	// CRITICAL FIX: Add reasons with deduplication
	// This matches TypeScript behavior where unique extension reasons are preserved
	if reason != "" {
		// Check if reason already exists to avoid duplicates
		reasonExists := false
		for _, existing := range p.Reason {
			if existing == reason {
				reasonExists = true
				break
			}
		}
		if !reasonExists {
			p.Reason = append(p.Reason, reason)
		}
	}
}

// AddTechs adds multiple technologies
func (p *Payload) AddTechs(techs map[string][]string) {
	for tech, reasons := range techs {
		for _, reason := range reasons {
			p.AddTech(tech, reason)
		}
	}
}

// AddLanguage increments the count for a language
func (p *Payload) AddLanguage(language string) {
	p.Languages[language]++
}

// AddLicense adds a license to the payload (like TypeScript addLicenses)
func (p *Payload) AddLicense(license string) {
	// Avoid duplicates (like TypeScript Set behavior)
	for _, existing := range p.Licenses {
		if existing == license {
			return
		}
	}

	p.Licenses = append(p.Licenses, license)
}

// AddEdges adds an edge to another payload (like TypeScript addEdges)
func (p *Payload) AddEdges(target *Payload) {
	edge := Edge{
		Target: target,
		Read:   true,
		Write:  true,
	}

	p.Edges = append(p.Edges, edge)
}

// AddDependency adds a dependency
func (p *Payload) AddDependency(dep Dependency) {
	p.Dependencies = append(p.Dependencies, dep)
}

// DetectLanguage detects the language from a file name using go-enry (GitHub Linguist)
func (p *Payload) DetectLanguage(filename string) {
	// Try detection by extension first (fast path)
	lang, safe := enry.GetLanguageByExtension(filename)

	// If not confident or no result, try by filename (handles special files like Makefile, Dockerfile)
	if !safe || lang == "" {
		lang, _ = enry.GetLanguageByFilename(filename)
	}

	// Add language if detected
	if lang != "" {
		p.AddLanguage(lang)
	}
}

// GetFullPath returns the full path as a string
func (p *Payload) GetFullPath() string {
	return strings.Join(p.Path, "/")
}

// String returns a string representation
func (p *Payload) String() string {
	techStr := "nil"
	if p.Tech != nil {
		techStr = *p.Tech
	}
	return fmt.Sprintf("Payload{id:%s, name:%s, tech:%s, techs:%v}",
		p.ID, p.Name, techStr, p.Techs)
}
