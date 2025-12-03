// Package codestats provides code statistics analysis (lines of code, comments, blanks, complexity)
package codestats

import (
	"math"
	"sort"
	"sync"

	"github.com/go-enry/go-enry/v2"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

// round2 rounds a float to 2 decimal places
func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

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

// TopLanguage represents a top programming language in Metrics
type TopLanguage struct {
	Language string  `json:"language"`
	Pct      float64 `json:"pct"`
}

// Metrics holds derived code metrics (programming languages only)
type Metrics struct {
	CommentRatio      float64       `json:"comment_ratio"`       // comments / code (documentation level)
	CodeDensity       float64       `json:"code_density"`        // code / lines (actual code vs whitespace/comments)
	AvgFileSize       float64       `json:"avg_file_size"`       // lines / files
	ComplexityPerKLOC float64       `json:"complexity_per_kloc"` // complexity / (code / 1000)
	AvgComplexity     float64       `json:"avg_complexity"`      // complexity / files
	TopLanguages      []TopLanguage `json:"top_languages"`       // top 5 programming languages (≥1%)
}

// TypeBucket holds stats for a language type (programming, data, markup, prose)
type TypeBucket struct {
	Total     Stats    `json:"total"`             // Aggregated stats for this type
	Metrics   *Metrics `json:"metrics,omitempty"` // Derived metrics (programming languages only)
	Languages []string `json:"languages"`         // Languages in this type (sorted by lines desc)
}

// ByType holds stats grouped by GitHub Linguist language type
type ByType struct {
	Programming *TypeBucket `json:"programming,omitempty"`
	Data        *TypeBucket `json:"data,omitempty"`
	Markup      *TypeBucket `json:"markup,omitempty"`
	Prose       *TypeBucket `json:"prose,omitempty"`
}

// CodeStats holds aggregated code statistics
type CodeStats struct {
	Total      Stats            `json:"total"`      // Grand total (analyzed only)
	ByType     ByType           `json:"by_type"`    // Stats grouped by language type (metrics in programming section)
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
	// By type aggregation (programming, data, markup, prose)
	byType map[string]*Stats // key: "programming", "data", "markup", "prose"
}

func newSCCAnalyzer() *sccAnalyzer {
	return &sccAnalyzer{
		codeByLanguage:  make(map[string]*Stats),
		otherByLanguage: make(map[string]*OtherStats),
		byType:          make(map[string]*Stats),
	}
}

func (a *sccAnalyzer) IsEnabled() bool {
	return true
}

// buildAnalyzedSlice converts codeByLanguage map to sorted slice
func (a *sccAnalyzer) buildAnalyzedSlice() []LanguageStats {
	result := make([]LanguageStats, 0, len(a.codeByLanguage))
	for lang, stats := range a.codeByLanguage {
		result = append(result, LanguageStats{
			Language: lang, Lines: stats.Lines, Code: stats.Code,
			Comments: stats.Comments, Blanks: stats.Blanks,
			Complexity: stats.Complexity, Files: stats.Files,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Lines > result[j].Lines })
	return result
}

// buildUnanalyzedSlice converts otherByLanguage map to sorted slice
func (a *sccAnalyzer) buildUnanalyzedSlice() []OtherLanguageStats {
	result := make([]OtherLanguageStats, 0, len(a.otherByLanguage))
	for lang, stats := range a.otherByLanguage {
		result = append(result, OtherLanguageStats{Language: lang, Lines: stats.Lines, Files: stats.Files})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Lines > result[j].Lines })
	return result
}

// buildByType creates the ByType structure from language stats
func (a *sccAnalyzer) buildByType(analyzed []LanguageStats, unanalyzed []OtherLanguageStats, metrics Metrics) ByType {
	typeLanguages := make(map[string][]string)

	// Collect languages by type (analyzed)
	for _, ls := range analyzed {
		typeName := types.LanguageTypeToString(enry.GetLanguageType(ls.Language))
		if typeName != "unknown" {
			typeLanguages[typeName] = append(typeLanguages[typeName], ls.Language)
		}
	}
	// Collect languages by type (unanalyzed)
	for _, ls := range unanalyzed {
		typeName := types.LanguageTypeToString(enry.GetLanguageType(ls.Language))
		if typeName != "unknown" {
			typeLanguages[typeName] = append(typeLanguages[typeName], ls.Language)
		}
	}

	byType := ByType{}
	for typeName, langs := range typeLanguages {
		if stats, ok := a.byType[typeName]; ok {
			bucket := &TypeBucket{Total: *stats, Languages: langs}

			// Only add metrics for programming type (since they're calculated from programming languages only)
			if typeName == "programming" && stats.Code > 0 {
				bucket.Metrics = &metrics
			}

			switch typeName {
			case "programming":
				byType.Programming = bucket
			case "data":
				byType.Data = bucket
			case "markup":
				byType.Markup = bucket
			case "prose":
				byType.Prose = bucket
			}
		}
	}
	return byType
}

// calculateMetrics computes Metrics from programming language stats
func (a *sccAnalyzer) calculateMetrics(analyzed []LanguageStats) Metrics {
	kpis := Metrics{}
	progStats := a.byType["programming"]
	if progStats == nil {
		return kpis
	}

	if progStats.Code > 0 {
		kpis.CommentRatio = round2(float64(progStats.Comments) / float64(progStats.Code))
		kpis.ComplexityPerKLOC = round2(float64(progStats.Complexity) / (float64(progStats.Code) / 1000))
	}
	if progStats.Lines > 0 {
		kpis.CodeDensity = round2(float64(progStats.Code) / float64(progStats.Lines))
	}
	if progStats.Files > 0 {
		kpis.AvgFileSize = round2(float64(progStats.Lines) / float64(progStats.Files))
		kpis.AvgComplexity = round2(float64(progStats.Complexity) / float64(progStats.Files))
	}

	// Top programming languages
	var progLangs []LanguageStats
	for _, ls := range analyzed {
		if enry.GetLanguageType(ls.Language) == enry.Programming {
			progLangs = append(progLangs, ls)
		}
	}
	a.setTopLanguages(&kpis, progLangs, progStats.Lines)
	return kpis
}

// setTopLanguages sets top programming languages (max 5, ≥1% threshold)
func (a *sccAnalyzer) setTopLanguages(kpis *Metrics, progLangs []LanguageStats, totalLines int64) {
	if totalLines == 0 || len(progLangs) == 0 {
		return
	}
	const maxLangs = 5
	const minPct = 0.01 // 1% threshold

	for i, ls := range progLangs {
		if i >= maxLangs {
			break
		}
		pct := round2(float64(ls.Lines) / float64(totalLines))
		if pct < minPct {
			break // Languages are sorted by lines, so remaining will be smaller
		}
		kpis.TopLanguages = append(kpis.TopLanguages, TopLanguage{
			Language: ls.Language,
			Pct:      pct,
		})
	}
}

func (a *sccAnalyzer) GetStats() interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()

	analyzed := a.buildAnalyzedSlice()
	unanalyzed := a.buildUnanalyzedSlice()
	metrics := a.calculateMetrics(analyzed)

	return &CodeStats{
		Total:      a.total,
		ByType:     a.buildByType(analyzed, unanalyzed, metrics),
		Analyzed:   AnalyzedBucket{Total: a.total, ByLanguage: analyzed},
		Unanalyzed: UnanalyzedBucket{Total: a.otherTotal, ByLanguage: unanalyzed},
	}
}
