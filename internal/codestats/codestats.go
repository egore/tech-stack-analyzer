// Package codestats provides code statistics analysis (lines of code, comments, blanks, complexity)
package codestats

import (
	"sort"
	"sync"
)

// Stats holds code statistics for a language or total (SCC-analyzed files)
type Stats struct {
	Lines      int64 `json:"lines"`
	Code       int64 `json:"code"`
	Comments   int64 `json:"comments"`
	Blanks     int64 `json:"blanks"`
	Complexity int64 `json:"complexity"`
	Files      int   `json:"files"`
}

// LanguageStats holds stats for a specific language (includes language name for sorted output)
type LanguageStats struct {
	Language   string `json:"language"`
	Lines      int64  `json:"lines"`
	Code       int64  `json:"code"`
	Comments   int64  `json:"comments"`
	Blanks     int64  `json:"blanks"`
	Complexity int64  `json:"complexity"`
	Files      int    `json:"files"`
}

// OtherStats holds statistics for files SCC cannot analyze (just line counts)
type OtherStats struct {
	Lines int64 `json:"lines"`
	Files int   `json:"files"`
}

// OtherLanguageStats holds stats for unanalyzed language (includes language name)
type OtherLanguageStats struct {
	Language string `json:"language"`
	Lines    int64  `json:"lines"`
	Files    int    `json:"files"`
}

// AnalyzedBucket holds stats for SCC-analyzed files (full code/comments/blanks breakdown)
type AnalyzedBucket struct {
	Total      Stats           `json:"total"`
	ByLanguage []LanguageStats `json:"by_language"` // Sorted by lines descending
}

// UnanalyzedBucket holds stats for files SCC cannot analyze (just line counts)
type UnanalyzedBucket struct {
	Total      OtherStats           `json:"total"`
	ByLanguage []OtherLanguageStats `json:"by_language"` // Sorted by lines descending
}

// CodeStats holds aggregated code statistics
type CodeStats struct {
	Total      Stats            `json:"total"`      // Grand total (analyzed only)
	Analyzed   AnalyzedBucket   `json:"analyzed"`   // SCC-recognized languages
	Unanalyzed UnanalyzedBucket `json:"unanalyzed"` // Files SCC can't parse
}

// Analyzer interface for code statistics collection
type Analyzer interface {
	// ProcessFile analyzes a file and adds its stats
	// language is the go-enry detected language name (used for grouping)
	// If content is provided, it will be used; otherwise the file will be read
	ProcessFile(filename string, language string, content []byte)

	// GetStats returns the aggregated statistics
	GetStats() interface{}

	// IsEnabled returns whether code stats collection is enabled
	IsEnabled() bool
}

// NewAnalyzer creates an analyzer based on enabled flag
func NewAnalyzer(enabled bool) Analyzer {
	if enabled {
		return newSCCAnalyzer()
	}
	return &noopAnalyzer{}
}

// noopAnalyzer is a no-op implementation when code stats are disabled
type noopAnalyzer struct{}

func (n *noopAnalyzer) ProcessFile(filename string, language string, content []byte) {} // language and content optional
func (n *noopAnalyzer) GetStats() interface{}                                        { return nil }
func (n *noopAnalyzer) IsEnabled() bool                                              { return false }

// sccAnalyzer uses boyter/scc for code statistics
type sccAnalyzer struct {
	mu              sync.Mutex
	total           Stats                  // Grand total (code only)
	codeByLanguage  map[string]*Stats      // SCC-analyzed languages
	otherTotal      OtherStats             // Total for non-SCC files
	otherByLanguage map[string]*OtherStats // Non-SCC languages
}

func newSCCAnalyzer() *sccAnalyzer {
	return &sccAnalyzer{
		codeByLanguage:  make(map[string]*Stats),
		otherByLanguage: make(map[string]*OtherStats),
	}
}

func (a *sccAnalyzer) IsEnabled() bool {
	return true
}

func (a *sccAnalyzer) GetStats() interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Convert analyzed bucket to sorted slice
	analyzedByLang := make([]LanguageStats, 0, len(a.codeByLanguage))
	for lang, stats := range a.codeByLanguage {
		analyzedByLang = append(analyzedByLang, LanguageStats{
			Language:   lang,
			Lines:      stats.Lines,
			Code:       stats.Code,
			Comments:   stats.Comments,
			Blanks:     stats.Blanks,
			Complexity: stats.Complexity,
			Files:      stats.Files,
		})
	}
	// Sort by lines descending
	sort.Slice(analyzedByLang, func(i, j int) bool {
		return analyzedByLang[i].Lines > analyzedByLang[j].Lines
	})

	// Convert unanalyzed bucket to sorted slice
	unanalyzedByLang := make([]OtherLanguageStats, 0, len(a.otherByLanguage))
	for lang, stats := range a.otherByLanguage {
		unanalyzedByLang = append(unanalyzedByLang, OtherLanguageStats{
			Language: lang,
			Lines:    stats.Lines,
			Files:    stats.Files,
		})
	}
	// Sort by lines descending
	sort.Slice(unanalyzedByLang, func(i, j int) bool {
		return unanalyzedByLang[i].Lines > unanalyzedByLang[j].Lines
	})

	return &CodeStats{
		Total: a.total,
		Analyzed: AnalyzedBucket{
			Total:      a.total,
			ByLanguage: analyzedByLang,
		},
		Unanalyzed: UnanalyzedBucket{
			Total:      a.otherTotal,
			ByLanguage: unanalyzedByLang,
		},
	}
}
