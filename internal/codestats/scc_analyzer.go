package codestats

import (
	"os"
	"sync"

	"github.com/boyter/scc/v3/processor"
	"github.com/go-enry/go-enry/v2"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

var initOnce sync.Once

// ProcessFile analyzes a file using SCC and aggregates stats
// language is the go-enry detected language name (used for grouping in by_language)
// If content is provided (non-nil, non-empty), it will be used; otherwise the file will be read
func (a *sccAnalyzer) ProcessFile(filename string, language string, content []byte) {
	// Skip if no language detected by go-enry
	if language == "" {
		return
	}

	// Read file if content not provided
	if len(content) == 0 {
		var err error
		content, err = os.ReadFile(filename)
		if err != nil || len(content) == 0 {
			return
		}
	}

	// Initialize SCC language definitions once
	initOnce.Do(func() {
		processor.ProcessConstants()
	})

	// Detect SCC language for proper comment/code parsing
	// SCC needs its own language name to know comment syntax
	// If SCC doesn't recognize the file, use empty string (counts all as code)
	sccLangs, _ := processor.DetectLanguage(filename)
	sccLang := ""
	if len(sccLangs) > 0 {
		sccLang = sccLangs[0]
	}
	// Note: sccLang can be empty - SCC will treat as plain text (all lines = code)

	// Create file job for SCC
	filejob := &processor.FileJob{
		Filename: filename,
		Language: sccLang, // Use SCC's language for parsing
		Content:  content,
		Bytes:    int64(len(content)),
	}

	// Count stats
	processor.CountStats(filejob)

	// Aggregate results
	a.mu.Lock()
	defer a.mu.Unlock()

	// Determine if SCC recognized this file (can distinguish code/comments)
	sccRecognized := sccLang != ""

	// Get language type from go-enry (programming, data, markup, prose)
	langType := enry.GetLanguageType(language)
	typeName := types.LanguageTypeToString(langType)

	if sccRecognized {
		// Analyzed bucket: SCC-analyzed files with full stats
		a.total.Lines += filejob.Lines
		a.total.Code += filejob.Code
		a.total.Comments += filejob.Comment
		a.total.Blanks += filejob.Blank
		a.total.Complexity += filejob.Complexity
		a.total.Files++

		if _, ok := a.codeByLanguage[language]; !ok {
			a.codeByLanguage[language] = &Stats{}
		}
		a.codeByLanguage[language].Lines += filejob.Lines
		a.codeByLanguage[language].Code += filejob.Code
		a.codeByLanguage[language].Comments += filejob.Comment
		a.codeByLanguage[language].Blanks += filejob.Blank
		a.codeByLanguage[language].Complexity += filejob.Complexity
		a.codeByLanguage[language].Files++

		// Aggregate by language type
		if typeName != "unknown" {
			if _, ok := a.byType[typeName]; !ok {
				a.byType[typeName] = &Stats{}
			}
			a.byType[typeName].Lines += filejob.Lines
			a.byType[typeName].Code += filejob.Code
			a.byType[typeName].Comments += filejob.Comment
			a.byType[typeName].Blanks += filejob.Blank
			a.byType[typeName].Complexity += filejob.Complexity
			a.byType[typeName].Files++
		}
	} else {
		// Other bucket: files SCC can't analyze (just line counts)
		a.otherTotal.Lines += filejob.Lines
		a.otherTotal.Files++

		if _, ok := a.otherByLanguage[language]; !ok {
			a.otherByLanguage[language] = &OtherStats{}
		}
		a.otherByLanguage[language].Lines += filejob.Lines
		a.otherByLanguage[language].Files++

		// Also aggregate by type for unanalyzed files
		if typeName != "unknown" {
			if _, ok := a.byType[typeName]; !ok {
				a.byType[typeName] = &Stats{}
			}
			a.byType[typeName].Lines += filejob.Lines
			a.byType[typeName].Files++
		}
	}
}
