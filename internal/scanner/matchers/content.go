package matchers

import (
	"regexp"

	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

// ContentMatcher handles content-based pattern matching for technology detection
type ContentMatcher struct {
	Tech    string
	Pattern *regexp.Regexp
}

// ContentMatcherRegistry manages compiled content matchers
type ContentMatcherRegistry struct {
	matchers map[string][]*ContentMatcher // keyed by extension (e.g., ".cpp", ".h")
}

// NewContentMatcherRegistry creates a new content matcher registry
func NewContentMatcherRegistry() *ContentMatcherRegistry {
	return &ContentMatcherRegistry{
		matchers: make(map[string][]*ContentMatcher),
	}
}

// BuildFromRules compiles content patterns from rules
func (r *ContentMatcherRegistry) BuildFromRules(rules []types.Rule) error {
	for _, rule := range rules {
		// Skip rules with dependencies - they don't need content validation
		if len(rule.Dependencies) > 0 {
			continue
		}

		// Skip rules without content patterns
		if len(rule.Content) == 0 {
			continue
		}

		// Skip rules without extensions - content matching requires extension context
		if len(rule.Extensions) == 0 {
			continue
		}

		// Compile content patterns for each extension
		for _, ext := range rule.Extensions {
			for _, contentRule := range rule.Content {
				// Compile regex pattern
				pattern, err := regexp.Compile(contentRule.Pattern)
				if err != nil {
					// Skip invalid patterns
					continue
				}

				matcher := &ContentMatcher{
					Tech:    rule.Tech,
					Pattern: pattern,
				}

				r.matchers[ext] = append(r.matchers[ext], matcher)
			}
		}
	}

	// Patterns are processed in the order they appear in the rule

	return nil
}

// MatchContent checks if content matches any patterns for the given extension
// Returns map of tech -> reasons
// Stops after first match per tech (rule is satisfied with one pattern match)
func (r *ContentMatcherRegistry) MatchContent(extension string, content string) map[string][]string {
	results := make(map[string][]string)

	matchers, exists := r.matchers[extension]
	if !exists {
		return results
	}

	// Check patterns in order - stop after first match per tech
	for _, matcher := range matchers {
		// Skip if we already matched this tech
		if _, alreadyMatched := results[matcher.Tech]; alreadyMatched {
			continue
		}

		if matcher.Pattern.MatchString(content) {
			results[matcher.Tech] = []string{
				"content matched: " + matcher.Pattern.String(),
			}
		}
	}

	return results
}

// HasContentMatchers checks if there are any content matchers for the given extension
func (r *ContentMatcherRegistry) HasContentMatchers(extension string) bool {
	_, exists := r.matchers[extension]
	return exists
}
