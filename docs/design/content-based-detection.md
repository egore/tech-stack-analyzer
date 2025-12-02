# Content-Based Detection

## Overview

Content-based detection adds pattern matching capabilities to the scanner, enabling technology detection based on file content rather than just filenames and extensions.

## Architecture

### Components

1. **ContentRule** (`internal/types/rule.go`)
   - Defines content patterns in YAML rules
   - Fields: `pattern` (regex), `priority` (optional)

2. **ContentMatcherRegistry** (`internal/scanner/matchers/content.go`)
   - Compiles and manages content patterns
   - Indexes patterns by file extension for fast lookup
   - Sorts patterns by priority (higher first)

3. **Scanner Integration** (`internal/scanner/scanner.go`)
   - `detectByContent()` method processes files with content patterns
   - Only reads files that have matching content rules
   - Integrates seamlessly with existing detection flow

## Rule Structure

```yaml
tech: stl
name: C++ Standard Template Library
type: library
extensions:
  - .cpp
  - .h
  - .hpp
content:
  - pattern: '#include\s+<vector>'
    priority: 10
  - pattern: 'std::vector'
    priority: 5
```

## Detection Flow

```
File → Extension Match → Content Pattern Check → Tech Detected
```

### Key Principles

1. **Extension-First**: Content matching only applies to files with matching extensions
2. **Dependency Skip**: Rules with `dependencies` skip content validation (already definitive)
3. **Optional Validation**: Content patterns are optional - rules without them work as before
4. **Performance**: Only reads files when necessary, early exit on extension mismatch

## When to Use Content Detection

### ✅ Good Use Cases

- **Languages with shared extensions**: Distinguish C++ from C in `.h` files
- **Libraries without package managers**: Detect OpenGL, STL through includes
- **Framework-specific patterns**: React hooks, Vue composition API
- **Code pattern detection**: Specific API usage, coding styles

### ❌ Don't Use For

- **Package-based detection**: npm, pip, maven dependencies are already definitive
- **Config files**: `webpack.config.js`, `tsconfig.json` are unique identifiers
- **Redundant validation**: If extension/filename is sufficient

## Performance Characteristics

- **O(1) extension lookup**: Fast check if content matching needed
- **Lazy file reading**: Only reads files with matching patterns
- **Compiled regex**: Patterns compiled once at startup
- **Priority ordering**: Stops after first match per tech
- **Zero overhead**: Rules without content patterns have no performance impact

## Example Rules

### C++ STL Detection
```yaml
tech: stl
name: C++ Standard Template Library
type: library
extensions: [.cpp, .h, .hpp]
content:
  - pattern: '#include\s+<(vector|string|map|set)'
    priority: 10
```

### OpenGL Detection
```yaml
tech: opengl
name: OpenGL
type: library
extensions: [.c, .cpp, .h]
content:
  - pattern: '#include\s+<GL/'
    priority: 10
  - pattern: 'glBegin|glEnd|glVertex'
    priority: 5
```

### React Hooks Detection
```yaml
tech: react-hooks
name: React Hooks
type: framework
extensions: [.js, .jsx, .ts, .tsx]
content:
  - pattern: 'useState|useEffect|useContext'
    priority: 10
```

## Testing

Comprehensive test coverage in `internal/scanner/matchers/content_test.go`:

- Pattern compilation and matching
- Extension-based filtering
- Priority ordering
- Rule validation (skips dependencies, requires extensions)
- Multiple pattern matching

Run tests:
```bash
go test ./internal/scanner/matchers/... -v
```

## Implementation Notes

1. **No File Caching**: Scanner reads each file once, no caching needed
2. **Sequential Processing**: Maintains existing streaming architecture
3. **Backward Compatible**: Existing rules work without modification
4. **Zero Dependencies**: Uses standard library `regexp` package

## Future Enhancements

Potential improvements (not implemented):

- **Size Limits**: Skip files larger than threshold
- **High-Performance Regex**: Consider re2go for 290x speedup
- **Parallel Processing**: Process multiple files concurrently
- **Content Sampling**: Check first N lines for performance
