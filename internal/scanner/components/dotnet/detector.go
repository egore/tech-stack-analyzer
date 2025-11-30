package dotnet

import (
	"path/filepath"
	"regexp"

	"github.com/petrarca/tech-stack-analyzer/internal/scanner/components"
	"github.com/petrarca/tech-stack-analyzer/internal/scanner/parsers"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

type Detector struct{}

func (d *Detector) Name() string {
	return "dotnet"
}

func (d *Detector) Detect(files []types.File, currentPath, basePath string, provider types.Provider, depDetector components.DependencyDetector) []*types.Payload {
	var results []*types.Payload

	// Check for .csproj files
	csprojRegex := regexp.MustCompile(`\.csproj$`)
	for _, file := range files {
		if csprojRegex.MatchString(file.Name) {
			payload := d.detectDotNetProject(file, currentPath, basePath, provider, depDetector)
			if payload != nil {
				results = append(results, payload)
			}
		}
	}

	return results
}

func (d *Detector) detectDotNetProject(file types.File, currentPath, basePath string, provider types.Provider, depDetector components.DependencyDetector) *types.Payload {
	content, err := provider.ReadFile(filepath.Join(currentPath, file.Name))
	if err != nil {
		return nil
	}

	// Parse .csproj file using parser
	dotnetParser := parsers.NewDotNetParser()
	project := dotnetParser.ParseCsproj(string(content))

	if project.Name == "" {
		return nil
	}

	// Create component payload
	relativeFilePath, _ := filepath.Rel(basePath, filepath.Join(currentPath, file.Name))
	if relativeFilePath == "." {
		relativeFilePath = "/"
	} else {
		relativeFilePath = "/" + relativeFilePath
	}

	payload := types.NewPayloadWithPath(project.Name, relativeFilePath)

	// Set tech to dotnet
	tech := "dotnet"
	payload.Tech = &tech

	// Add framework info to techs
	if project.Framework != "" {
		payload.AddTech(tech, "framework: "+project.Framework)
	} else {
		payload.AddTech(tech, "matched file: "+file.Name)
	}

	// Create dependencies list
	var dependencies []types.Dependency

	// Add NuGet package dependencies
	for _, pkg := range project.Packages {
		dep := types.Dependency{
			Type:    "nuget",
			Name:    pkg.Name,
			Example: pkg.Version,
		}
		dependencies = append(dependencies, dep)

		// Match package name against dependency rules
		matchedTechs := depDetector.MatchDependencies([]string{pkg.Name}, "nuget")

		// Determine tech and reasons for child components
		var childTech string
		var reasons []string
		for t, r := range matchedTechs {
			childTech = t
			reasons = r
			break // Take first match
		}

		if childTech != "" && childTech != tech { // Only create child if different tech
			if len(reasons) == 0 {
				reasons = []string{"matched: " + pkg.Name}
			}

			// Create child component for matched tech
			childPayload := types.NewPayloadWithPath(pkg.Name, relativeFilePath)
			childPayload.Tech = &childTech
			childPayload.Dependencies = []types.Dependency{dep}

			// Add techs and reasons to child
			for _, reason := range reasons {
				childPayload.AddTech(childTech, reason)
			}

			// Add child to parent payload
			payload.AddChild(childPayload)
		}
	}

	// Set dependencies on parent
	payload.Dependencies = dependencies

	return payload
}

func init() {
	components.Register(&Detector{})
}
