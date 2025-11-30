package scanner

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/petrarca/tech-stack-analyzer/internal/provider"
	"github.com/petrarca/tech-stack-analyzer/internal/rules"
	"github.com/petrarca/tech-stack-analyzer/internal/scanner/components"
	"github.com/petrarca/tech-stack-analyzer/internal/scanner/matchers"
	"github.com/petrarca/tech-stack-analyzer/internal/scanner/parsers"

	// Import component detectors to trigger init() registration
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/deno"
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/docker"
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/dotnet"
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/golang"
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/java"
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/nodejs"
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/php"
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/python"
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/ruby"
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/rust"
	_ "github.com/petrarca/tech-stack-analyzer/internal/scanner/components/terraform"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

// Scanner handles the scanning logic (like TypeScript's Payload.recurse)
type Scanner struct {
	provider       types.Provider
	rules          []types.Rule
	depDetector    *DependencyDetector
	compDetector   *ComponentDetector
	jsonDetector   *JSONSchemaDetector
	dotenvDetector *parsers.DotenvDetector
	excludeDirs    []string
}

// NewScanner creates a new scanner (mirroring TypeScript's analyser function)
func NewScanner(path string) (*Scanner, error) {
	// Create provider for the target path (like TypeScript's FSProvider)
	provider := provider.NewFSProvider(path)

	// Load rules (simple, not lazy loaded - like TypeScript's loadAllRules)
	rules, err := rules.LoadEmbeddedRules()
	if err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}

	// Initialize dependency detector
	depDetector := NewDependencyDetector(rules)

	// Initialize component detector
	compDetector := NewComponentDetector(depDetector, provider, rules)

	// Initialize JSON schema detector
	jsonDetector := NewJSONSchemaDetector(provider, rules)

	// Initialize dotenv detector
	dotenvDetector := parsers.NewDotenvDetector(provider, rules)

	// Build matchers from rules (like TypeScript's loadAllRules)
	matchers.BuildFileMatchersFromRules(rules)
	matchers.BuildExtensionMatchersFromRules(rules)

	return &Scanner{
		provider:       provider,
		rules:          rules,
		depDetector:    depDetector,
		compDetector:   compDetector,
		jsonDetector:   jsonDetector,
		dotenvDetector: dotenvDetector,
		excludeDirs:    nil,
	}, nil
}

// NewScannerWithExcludes creates a new scanner with directory exclusions
func NewScannerWithExcludes(path string, excludeDirs []string) (*Scanner, error) {
	// Create provider for the target path (like TypeScript's FSProvider)
	provider := provider.NewFSProvider(path)

	// Load rules (simple, not lazy loaded - like TypeScript's loadAllRules)
	rules, err := rules.LoadEmbeddedRules()
	if err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}

	// Initialize dependency detector
	depDetector := NewDependencyDetector(rules)

	// Initialize component detector
	compDetector := NewComponentDetector(depDetector, provider, rules)

	// Initialize JSON schema detector
	jsonDetector := NewJSONSchemaDetector(provider, rules)

	// Initialize dotenv detector
	dotenvDetector := parsers.NewDotenvDetector(provider, rules)

	// Build matchers from rules (like TypeScript's loadAllRules)
	matchers.BuildFileMatchersFromRules(rules)
	matchers.BuildExtensionMatchersFromRules(rules)

	return &Scanner{
		provider:       provider,
		rules:          rules,
		depDetector:    depDetector,
		compDetector:   compDetector,
		jsonDetector:   jsonDetector,
		dotenvDetector: dotenvDetector,
		excludeDirs:    excludeDirs,
	}, nil
}

// Scan performs analysis following the original TypeScript pattern
func (s *Scanner) Scan() (*types.Payload, error) {
	// Create main payload like in TypeScript: new Payload({ name: 'main', folderPath: '/' })
	payload := types.NewPayloadWithPath("main", "/")

	// Start recursion from base path (like TypeScript's payload.recurse(provider, provider.basePath))
	err := s.recurse(payload, s.provider.GetBasePath())
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// recurse follows the exact TypeScript implementation pattern
func (s *Scanner) recurse(payload *types.Payload, filePath string) error {
	// Get files in current directory (like TypeScript's provider.listDir)
	files, err := s.provider.ListDir(filePath)
	if err != nil {
		return err
	}

	// Apply rules to detect technologies (like TypeScript's ruleComponents loop)
	// This might return a different context if a component was detected
	ctx := s.applyRules(payload, files, filePath)

	// Process each file/directory (exactly like TypeScript's loop)
	for _, file := range files {
		if file.Type == "file" {
			// Detect language from file name (like TypeScript's detectLang)
			// Languages go into the current context (might be a component)
			ctx.DetectLanguage(file.Name)
			continue
		}

		// Skip ignored directories (like TypeScript's IGNORED_DIVE_PATHS)
		if s.shouldIgnoreDirectory(file.Name) {
			continue
		}

		// Recurse into subdirectory (like TypeScript's await ctx.recurse(provider, fp))
		// Important: We recurse with the CURRENT CONTEXT (ctx), not the original payload
		// This matches TypeScript's behavior where ctx might be a component
		childPath := filepath.Join(filePath, file.Name)
		err := s.recurse(ctx, childPath)
		if err != nil {
			return err
		}
	}

	// Note: Do NOT combine ctx back to payload
	// Components should remain separate with their own dependencies
	// Extension reasons are handled separately by the AddTech fix

	return nil
}

// applyRules applies rules to detect technologies (following TypeScript's pattern exactly)
func (s *Scanner) applyRules(payload *types.Payload, files []types.File, currentPath string) *types.Payload {
	ctx := payload

	// 1. Component-based detection
	ctx = s.detectComponents(payload, ctx, files, currentPath)

	// 2. GitHub Actions detection
	s.detectGitHubActions(payload, files, currentPath)

	// 3. Dotenv detection
	s.detectDotenv(ctx, files, currentPath)

	// 4. JSON schema detection
	s.detectJSONSchemas(payload, ctx, files, currentPath)

	// 5. File and extension-based detection
	matchedTechs := s.detectByFilesAndExtensions(ctx, files, currentPath)

	// 6. Legacy file-based detection
	s.detectLegacyFiles(ctx, files, matchedTechs)

	return ctx
}

func (s *Scanner) detectComponents(payload, ctx *types.Payload, files []types.File, currentPath string) *types.Payload {
	for _, detector := range components.GetDetectors() {
		detectedComponents := detector.Detect(files, currentPath, s.provider.GetBasePath(), s.provider, s.depDetector)
		for _, component := range detectedComponents {
			if component.Name == "virtual" {
				s.mergeVirtualPayload(payload, component, currentPath)
			} else {
				ctx = s.addNamedComponent(payload, component, currentPath)
			}
		}
	}
	return ctx
}

func (s *Scanner) mergeVirtualPayload(target, virtual *types.Payload, currentPath string) {
	for _, child := range virtual.Childs {
		target.AddChild(child)
	}
	target.Combine(virtual)
	for _, tech := range virtual.Techs {
		s.findImplicitComponentByTech(target, tech, currentPath, false)
	}
}

func (s *Scanner) addNamedComponent(payload, component *types.Payload, currentPath string) *types.Payload {
	payload.AddChild(component)
	for _, tech := range component.Techs {
		s.findImplicitComponentByTech(component, tech, currentPath, true)
	}
	return component
}

func (s *Scanner) detectGitHubActions(payload *types.Payload, files []types.File, currentPath string) {
	githubActionsComponents := s.compDetector.DetectGitHubActionsComponent(files, currentPath, s.provider.GetBasePath())
	if githubActionsComponents == nil {
		return
	}

	if githubActionsComponents.Name == "virtual" {
		s.mergeVirtualPayload(payload, githubActionsComponents, currentPath)
	} else {
		payload.AddChild(githubActionsComponents)
	}
}

func (s *Scanner) detectDotenv(ctx *types.Payload, files []types.File, currentPath string) {
	dotenvPayload := s.dotenvDetector.DetectInDotEnv(files, currentPath, s.provider.GetBasePath())
	if dotenvPayload != nil {
		s.mergeVirtualPayload(ctx, dotenvPayload, currentPath)
	}
}

func (s *Scanner) detectJSONSchemas(payload, ctx *types.Payload, files []types.File, currentPath string) {
	jsonComponents := s.jsonDetector.DetectJSONSchemaComponents(files, currentPath, s.provider.GetBasePath())
	for _, jsonComponent := range jsonComponents {
		if jsonComponent.Name == "virtual" {
			s.mergeVirtualPayload(ctx, jsonComponent, currentPath)
		} else {
			payload.AddChild(jsonComponent)
		}
	}
}

func (s *Scanner) detectByFilesAndExtensions(ctx *types.Payload, files []types.File, currentPath string) map[string]bool {
	matchedTechs := make(map[string]bool)

	// File-based detection
	fileMatches := matchers.MatchFiles(files, currentPath, s.provider.GetBasePath())
	s.processTechMatches(ctx, fileMatches, matchedTechs, currentPath, true)

	// Extension-based detection
	extensionMatches := matchers.MatchExtensions(files)
	s.processTechMatches(ctx, extensionMatches, matchedTechs, currentPath, false)

	return matchedTechs
}

func (s *Scanner) processTechMatches(ctx *types.Payload, matches map[string][]string, matchedTechs map[string]bool, currentPath string, addEdges bool) {
	for tech, reasons := range matches {
		if matchedTechs[tech] {
			continue
		}
		for _, reason := range reasons {
			ctx.AddTech(tech, reason)
		}
		matchedTechs[tech] = true
		s.findImplicitComponentByTech(ctx, tech, currentPath, addEdges)
	}
}

func (s *Scanner) detectLegacyFiles(ctx *types.Payload, files []types.File, matchedTechs map[string]bool) {
	for _, rule := range s.rules {
		if len(rule.Files) == 0 || matchedTechs[rule.Tech] {
			continue
		}
		if s.matchRuleFiles(rule, files) {
			ctx.AddTech(rule.Tech, fmt.Sprintf("matched file: %s", rule.Files[0]))
			matchedTechs[rule.Tech] = true
		}
	}
}

func (s *Scanner) matchRuleFiles(rule types.Rule, files []types.File) bool {
	for _, requiredFile := range rule.Files {
		for _, file := range files {
			if file.Name == requiredFile {
				return true
			}
		}
	}
	return false
}

// findImplicitComponentByTech finds the rule for a tech and creates an implicit component
func (s *Scanner) findImplicitComponentByTech(payload *types.Payload, tech string, currentPath string, addEdges bool) {
	// Find the rule for this tech
	for _, rule := range s.rules {
		if rule.Tech == tech {
			s.findImplicitComponent(payload, rule, currentPath, addEdges)
			return
		}
	}
}

// findImplicitComponent creates a child component for technologies that are not in the notAComponent set
// This replicates the TypeScript findImplicitComponent logic
func (s *Scanner) findImplicitComponent(payload *types.Payload, rule types.Rule, currentPath string, addEdges bool) {
	// notAComponent set from TypeScript (lines 6-22 in helpers.ts)
	notAComponent := map[string]bool{
		"ci":              true,
		"language":        true,
		"runtime":         true,
		"tool":            true,
		"framework":       true,
		"validation":      true,
		"builder":         true,
		"linter":          true,
		"test":            true,
		"orm":             true,
		"package_manager": true,
		"ui":              true,
		"ui_framework":    true,
		"iac":             true,
	}

	// If this tech type is in the notAComponent set, don't create a component
	if notAComponent[rule.Type] {
		return
	}

	// Create a new child component (like TypeScript lines 47-54)
	// CRITICAL FIX: Use parent's path, not currentPath (like TypeScript: folderPath: pl.path)
	component := types.NewPayload(rule.Name, payload.Path)
	component.Tech = &rule.Tech
	component.Reason = append(component.Reason, fmt.Sprintf("matched file: %s", currentPath))

	// Add the component as a child
	payload.AddChild(component)

	// Add edges for non-hosting/cloud types if requested (like TypeScript: if (ref.type !== 'hosting' && ref.type !== 'cloud'))
	if addEdges && rule.Type != "hosting" && rule.Type != "cloud" {
		payload.AddEdges(component)
	}
}

// shouldIgnoreDirectory mirrors TypeScript's IGNORED_DIVE_PATHS exactly
// plus user-specified exclude directories
func (s *Scanner) shouldIgnoreDirectory(name string) bool {
	// Check user-specified exclude directories first
	if s.excludeDirs != nil {
		lowerName := strings.ToLower(name)
		for _, excludeDir := range s.excludeDirs {
			// Support both exact match and path suffix match
			// e.g., "internal/rules" matches "internal/rules" and "rules" matches "rules"
			if lowerName == strings.ToLower(excludeDir) {
				return true
			}
			// Also check if the directory path contains the exclude pattern
			if strings.Contains(lowerName, strings.ToLower(excludeDir)) {
				return true
			}
		}
	}

	// Exact same ignore list as TypeScript implementation
	ignored := []string{
		"node_modules", "dist", "build", "bin", "static", "public", "vendor",
		"terraform.tfstate.d", "migrations", "tests", "e2e", "__fixtures__",
		"__snapshots__", "tmp",
		// -- Dot folder
		".artifacts", ".assets", ".azure", ".azure-pipelines", ".bundle",
		".cache", ".changelog", ".devcontainer", ".docker", ".dynamodb",
		".fusebox", ".git", // needed to detect github actions - '.github' is NOT ignored
		".gitlab", ".gradle", ".log", ".metadata", ".npm", ".nuxt",
		".react-email", ".release", ".semgrep", ".serverless", ".svn",
		".terraform", ".vercel", ".venv", ".vscode", ".vuepress",
		// Python virtual environments
		"venv", "__pycache__",
	}

	lowerName := strings.ToLower(name)
	for _, pattern := range ignored {
		// Use exact match to avoid false positives (e.g., .github matching .git)
		if lowerName == strings.ToLower(pattern) {
			return true
		}
	}

	return false
}
