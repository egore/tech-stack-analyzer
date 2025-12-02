# Scanner Architecture

## Overview

The scanner uses a modular, registry-based architecture inspired by the original TypeScript implementation. It features:
- **Component detector registry** for tech stacks (Node.js, Python, etc.)
- **Matcher registries** for file-based and extension-based detection
- **DependencyDetector class** for dependency matching (regex-based)
- **go-enry integration** for comprehensive language detection (GitHub Linguist)
- **Implicit component creation** for third-party technologies

## Directory Structure

```
internal/scanner/
├── scanner.go                    # Main scanning logic & orchestration
├── components/
│   ├── detector.go              # Detector interface
│   ├── registry.go              # Component detector registry
│   ├── nodejs/
│   │   └── detector.go          # Node.js detector
│   ├── python/
│   │   └── detector.go          # Python detector
│   └── [additional detectors...]
├── matchers/
│   ├── file.go                  # File matcher registry
│   └── extension.go             # Extension matcher registry
├── dependencies.go              # Dependency matching (DependencyDetector class)
├── github_actions.go            # GitHub Actions detection
├── dotenv.go                    # Dotenv detection
└── jsonschema.go                # JSON schema detection
```

## Scanning Flow

The scanner follows this sequence for each directory:

1. **List files** in current directory
2. **Apply rules** (detect components, match files/extensions, create implicit components)
3. **Process files**:
   - Detect language for each file
   - Skip ignored directories (`.venv`, `node_modules`, etc.)
4. **Recurse** into subdirectories with current context

```go
func (s *Scanner) recurse(payload *types.Payload, filePath string) error {
    files, _ := s.provider.ListDir(filePath)
    
    // Apply rules (component detection, file/extension matching)
    ctx := s.applyRules(payload, files, filePath)
    
    // Process each file/directory
    for _, file := range files {
        if file.Type == "file" {
            ctx.DetectLanguage(file.Name)
            continue
        }
        
        if s.shouldIgnoreDirectory(file.Name) {
            continue
        }
        
        // Recurse with current context
        s.recurse(ctx, filepath.Join(filePath, file.Name))
    }
    
    return nil
}
```

## How It Works

### 1. Detector Interface

All component detectors implement this interface:

```go
type Detector interface {
    Name() string
    Detect(files []types.File, currentPath, basePath string, 
           provider types.Provider, depDetector DependencyDetector) []*types.Payload
}
```

### 2. Auto-Registration

Each detector registers itself via `init()`:

```go
// internal/scanner/components/nodejs/detector.go
func init() {
    components.Register(&Detector{})
}
```

When the package is imported, `init()` runs automatically and registers the detector.

### 3. Matcher Registries

The scanner uses matcher registries for rule-based detection:

#### File Matcher Registry
Detects technologies based on file/directory presence:
```go
// Built at initialization from rules
matchers.BuildFileMatchersFromRules(rules)

// Used during scanning
fileMatches := matchers.MatchFiles(files, currentPath, basePath)
```
Examples: `package.json` → nodejs, `.github/workflows` → github.actions

#### Extension Matcher Registry
Detects technologies based on file extensions:
```go
// Built at initialization from rules
matchers.BuildExtensionMatchersFromRules(rules)

// Used during scanning
extensionMatches := matchers.MatchExtensions(files)
```
Examples: `.ts` → typescript, `.py` → python, `.css` → css

**Benefits of Matcher Registries:**
- **Performance**: Pre-built matchers, no rule iteration during scan
- **Modularity**: Matchers separated from scanner logic
- **Extensibility**: Easy to add new matcher types

### 4. Dependency Matching

Dependency matching uses a **DependencyDetector class** instead of a registry pattern:

```go
type DependencyDetector struct {
    matchers map[string][]*DependencyMatcher // keyed by type (npm, python, etc.)
}

func (d *DependencyDetector) MatchDependencies(packages []string, depType string) map[string][]string {
    // Match packages against compiled regex patterns
}
```

**Why not a registry?**
- Dependency matching is more complex (regex compilation, state management)
- Called by detectors with context (package type), not directly by scanner
- DependencyDetector provides better encapsulation than a global registry
- This is actually an improvement over the original TypeScript (which uses a global object)

### 5. Scanner Integration

The scanner imports detectors using blank imports to trigger registration:

```go
import (
    _ "github.com/stack-analyser/scanner/internal/scanner/components/nodejs"
    _ "github.com/stack-analyser/scanner/internal/scanner/components/python"
)
```

Then loops through all registered detectors:

```go
for _, detector := range components.GetDetectors() {
    detectedComponents := detector.Detect(files, currentPath, basePath, provider, depDetector)
    for _, component := range detectedComponents {
        ctx = component
        payload.AddChild(component)
        
        // Create implicit components for detected techs
        for _, tech := range component.Techs {
            s.findImplicitComponentByTech(component, tech, currentPath, true)
        }
    }
}
```

## Component Types

### Named Components
- Have a specific name extracted from their config file
- Become the context for nested detections
- Examples: Node.js projects (`@myapp/frontend`), Python projects (`mylib`)
- `tech` field is `null` (user projects)
- **Design choice**: Python as named (not virtual) for better structure

### Virtual Components
- Created for detection but merged into parent
- Don't appear as separate components in output
- Examples: GitHub Actions, Dotenv
- Used for path and tech accumulation
- Children are added to parent, data is combined

### Implicit Components
- Created automatically when a technology is detected
- Have a specific `tech` field matching the technology
- Examples: Nginx, Docker, OpenAI, MkDocs, Lucide Icons
- Represent third-party technologies/services
- Nested under the component that uses them
- Created with edges for non-hosting/cloud types

## Language Detection

The scanner uses **go-enry** (GitHub Linguist Go port) for comprehensive language detection:

```go
import "github.com/go-enry/go-enry/v2"

func (p *Payload) DetectLanguage(filename string) {
    // Try detection by extension first (fast path)
    lang, safe := enry.GetLanguageByExtension(filename)
    
    // Fallback to filename for special files (Makefile, Dockerfile)
    if !safe || lang == "" {
        lang, _ = enry.GetLanguageByFilename(filename)
    }
    
    if lang != "" {
        p.AddLanguage(lang)
    }
}
```

**Features:**
- Detects 1500+ languages from GitHub Linguist database
- Handles special files without extensions (Makefile, Dockerfile)
- Fast extension-based detection with filename fallback
- Language counts stored in `payload.Languages` map

## Supported Tech Stacks

### Currently Implemented

| Tech Stack | Detector File | Detection File | Component Type |
|------------|---------------|----------------|----------------|
| Node.js    | `nodejs/detector.go` | `package.json` | Named |
| Python     | `python/detector.go` | `pyproject.toml` | Named |
| GitHub Actions | `github_actions.go` | `.github/workflows/*.yml` | Virtual |
| Dotenv | `dotenv.go` | `.env*` files | Virtual |
| JSON Schema | `jsonschema.go` | `components.json` (Shadcn) | Virtual/Named |

### Future Tech Stacks

The following can be easily added using the same pattern:

- **Go** - `main.go` → Named component with `tech="golang"`
- **Ruby** - `Gemfile` → Virtual component
- **PHP** - `composer.json` → Virtual component
- **Rust** - `Cargo.toml` → Virtual component
- **Deno** - `deno.json` → Virtual component
- **Zig** - `build.zig` → Virtual component

## Adding a New Detector

### Step 1: Create Detector File

```go
// internal/scanner/components/golang/detector.go
package golang

import (
    "github.com/stack-analyser/scanner/internal/scanner/components"
    "github.com/stack-analyser/scanner/internal/types"
)

type Detector struct{}

func (d *Detector) Name() string {
    return "golang"
}

func (d *Detector) Detect(files []types.File, currentPath, basePath string, 
                          provider types.Provider, depDetector DependencyDetector) []*types.Payload {
    // Detection logic here
    return nil
}

func init() {
    components.Register(&Detector{})
}
```

### Step 2: Import in Scanner

```go
// internal/scanner/scanner.go
import (
    _ "github.com/stack-analyser/scanner/internal/scanner/components/golang"
)
```

That's it! The detector is now active.

## Comparison with Original TypeScript

### Architecture Similarities
- ✅ Modular structure (one directory per tech stack)
- ✅ Central registry pattern for components
- ✅ Auto-registration via init/import
- ✅ Loop through detectors during scan
- ✅ File-based and extension-based detection
- ✅ Dependency matching
- ✅ Implicit component creation
- ✅ Virtual component merging
- ✅ Edge creation

### Implementation Differences

| Feature | TypeScript | Go |
|---------|-----------|-----|
| Registration | `register()` calls | `init()` functions |
| Imports | Side-effect imports | Blank imports `_` |
| Detectors | Function types | Interface types |
| File matching | Inline loops | Matcher registry |
| Extension matching | Inline loops | Matcher registry |
| Dependency matching | Global object | DependencyDetector class |
| Language detection | Hardcoded list | go-enry (Linguist) |
| Python components | Virtual | Named (design choice) |

### Key Improvements & Differences

**Improvements in Go Version:**
1. **Matcher Registries**: Pre-built matchers for better performance
2. **DependencyDetector Class**: Better encapsulation than TypeScript's global object
3. **go-enry Integration**: 1500+ languages vs ~50 hardcoded
4. **Python as Named Components**: Better hierarchy and ownership (not virtual like TypeScript)
5. **Cleaner Ignore List**: Excludes `.venv` and `__pycache__` by default

**Intentional Differences:**
- `.venv` ignored by default - cleaner language stats
- Hierarchical component tree vs flat structure

**Known Limitations:**
- Tech bubbling not implemented - technologies stay in components, don't aggregate at root
- This results in fewer root techs but better component ownership

## Benefits

### Modularity
Each tech stack is self-contained in its own package. Changes to one detector don't affect others.

### Extensibility
Adding new tech stacks requires:
1. Create new detector file
2. Add blank import
No changes to core scanner logic.

### Testability
Each detector can be tested independently:

```go
func TestNodeJSDetector(t *testing.T) {
    detector := &nodejs.Detector{}
    // Test detection logic
}
```

### Maintainability
Clear separation of concerns:
- Scanner handles recursion and orchestration
- Detectors handle tech-specific logic
- Registry handles registration

## Performance

The registry pattern has minimal overhead:
- Detectors registered once at startup via `init()`
- Registry lookup is O(n) where n = number of detectors
- Typically n < 20, so overhead is negligible

## Thread Safety

The registry uses `sync.RWMutex` for thread-safe access:
- `Register()` uses write lock
- `GetDetectors()` uses read lock
- Safe for concurrent scanning (future feature)

## Implementation Status

### ✅ Fully Implemented
- Component detector registry (Node.js, Python)
- File matcher registry
- Extension matcher registry
- Language detection with go-enry (GitHub Linguist)
- Implicit component creation with edges
- Virtual component merging
- Dependency extraction
- GitHub Actions detection
- Dotenv detection
- JSON schema detection (Shadcn)
- Directory ignore list (`.venv`, `node_modules`, etc.)

### 📊 Test Coverage
Compare outputs with original TypeScript scanner:
```bash
# Run Go scanner
./bin/scanner /path/to/project > output-go.json

# Compare with original
diff output-ts.json output-go.json
```

**Expected differences:**
- Python components as named (hierarchical) vs virtual (flat)
- Fewer root techs (no tech bubbling)
- Different language counts (no `.venv` scanning)

All core functionality matches the original with intentional improvements! 🎉
