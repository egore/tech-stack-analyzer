package terraform

import (
	"path/filepath"
	"regexp"

	"github.com/petrarca/tech-stack-analyzer/internal/scanner/components"
	"github.com/petrarca/tech-stack-analyzer/internal/scanner/parsers"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

type Detector struct{}

func (d *Detector) Name() string {
	return "terraform"
}

func (d *Detector) Detect(files []types.File, currentPath, basePath string, provider types.Provider, depDetector components.DependencyDetector) []*types.Payload {
	var results []*types.Payload

	// Check for .terraform.lock.hcl
	for _, file := range files {
		if file.Name == ".terraform.lock.hcl" {
			payload := d.detectTerraformLock(file, currentPath, basePath, provider, depDetector)
			if payload != nil {
				results = append(results, payload)
			}
		}
	}

	// Check for .tf files (returns array of payloads)
	tfRegex := regexp.MustCompile(`\.tf$`)
	for _, file := range files {
		if tfRegex.MatchString(file.Name) {
			payloads := d.detectTerraformResource(file, currentPath, basePath, provider, depDetector)
			if len(payloads) > 0 {
				results = append(results, payloads...)
			}
		}
	}

	return results
}

func (d *Detector) detectTerraformLock(file types.File, currentPath, basePath string, provider types.Provider, depDetector components.DependencyDetector) *types.Payload {
	content, err := provider.ReadFile(filepath.Join(currentPath, file.Name))
	if err != nil {
		return nil
	}

	// Parse terraform lock file using parser
	terraformParser := parsers.NewTerraformParser()
	providers := terraformParser.ParseTerraformLock(string(content))

	if len(providers) == 0 {
		return nil
	}

	// Create virtual payload
	relativeFilePath, _ := filepath.Rel(basePath, filepath.Join(currentPath, file.Name))
	if relativeFilePath == "." {
		relativeFilePath = "/"
	} else {
		relativeFilePath = "/" + relativeFilePath
	}
	payload := types.NewPayloadWithPath("virtual", relativeFilePath)

	// Create dependencies list
	var dependencies []types.Dependency

	// Create child components for each provider
	for _, provider := range providers {
		// Add to dependencies list
		dependencies = append(dependencies, types.Dependency{
			Type:    "terraform",
			Name:    provider.Name,
			Example: provider.Version,
		})

		// Match provider name against dependency rules
		matchedTechs := depDetector.MatchDependencies([]string{provider.Name}, "terraform")

		// Determine tech and reasons
		var tech string
		var reasons []string
		for t, r := range matchedTechs {
			tech = t
			reasons = r
			break // Take first match
		}

		if tech == "" {
			continue // Skip providers that don't match known techs
		}

		if len(reasons) == 0 {
			reasons = []string{"matched: " + provider.Name}
		}

		// Create child component
		childPayload := types.NewPayloadWithPath(provider.Name, relativeFilePath)
		childPayload.AddPrimaryTech(tech)
		childPayload.Dependencies = []types.Dependency{
			{
				Type:    "terraform",
				Name:    provider.Name,
				Example: provider.Version,
			},
		}

		// Add techs and reasons to child
		for _, reason := range reasons {
			childPayload.AddTech(tech, reason)
		}

		// Add child to parent payload
		payload.AddChild(childPayload)
	}

	// Set dependencies on parent
	payload.Dependencies = dependencies

	return payload
}

func (d *Detector) detectTerraformResource(file types.File, currentPath, basePath string, provider types.Provider, depDetector components.DependencyDetector) []*types.Payload {
	content, err := provider.ReadFile(filepath.Join(currentPath, file.Name))
	if err != nil {
		return nil
	}

	// Skip files larger than 500KB
	if len(content) > 500_000 {
		return nil
	}

	// Parse terraform resource file using parser
	terraformParser := parsers.NewTerraformParser()
	resources := terraformParser.ParseTerraformResource(string(content))

	if len(resources) == 0 {
		return nil
	}

	// Create virtual payload
	relativeFilePath, _ := filepath.Rel(basePath, filepath.Join(currentPath, file.Name))
	if relativeFilePath == "." {
		relativeFilePath = "/"
	} else {
		relativeFilePath = "/" + relativeFilePath
	}
	payload := types.NewPayloadWithPath("virtual", relativeFilePath)

	// Create child components for each resource type
	for _, resource := range resources {
		// Match resource type against dependency rules
		matchedTechs := depDetector.MatchDependencies([]string{resource}, "terraform.resource")

		// Determine tech and reasons
		var tech string
		var reasons []string
		for t, r := range matchedTechs {
			tech = t
			reasons = r
			break // Take first match
		}

		if tech == "" {
			continue // Skip resources that don't match known techs
		}

		if len(reasons) == 0 {
			reasons = []string{"matched: " + resource}
		}

		// Create child component
		childPayload := types.NewPayloadWithPath(resource, relativeFilePath)
		childPayload.AddPrimaryTech(tech)
		childPayload.Dependencies = []types.Dependency{
			{
				Type:    "terraform.resource",
				Name:    resource,
				Example: "unknown",
			},
		}

		// Add techs and reasons to child
		for _, reason := range reasons {
			childPayload.AddTech(tech, reason)
		}

		// Add child to parent payload
		payload.AddChild(childPayload)
	}

	return []*types.Payload{payload}
}

func init() {
	components.Register(&Detector{})
}
