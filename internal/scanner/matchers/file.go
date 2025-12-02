package matchers

import (
	"regexp"
	"strings"

	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

// FileMatcher is a function that matches files and returns the matched tech and file
type FileMatcher func(files []types.File, currentPath, basePath string) (tech string, matchedFile string, matched bool)

// fileMatcherRegistry holds all registered file matchers
var fileMatchers []FileMatcher

// RegisterFileMatcher adds a file matcher to the registry
func RegisterFileMatcher(matcher FileMatcher) {
	fileMatchers = append(fileMatchers, matcher)
}

// GetFileMatchers returns all registered file matchers
func GetFileMatchers() []FileMatcher {
	return fileMatchers
}

// ClearFileMatchers clears all registered file matchers (useful for testing)
func ClearFileMatchers() {
	fileMatchers = nil
}

// BuildFileMatchersFromRules creates file matchers from rules
func BuildFileMatchersFromRules(rules []types.Rule) {
	for _, rule := range rules {
		if len(rule.Files) == 0 {
			continue
		}

		// Skip package managers - they should not be promoted to main techs
		if rule.Type == "package_manager" {
			continue
		}

		// Create matcher for this rule
		RegisterFileMatcher(createFileMatcherForRule(rule))
	}
}

// createFileMatcherForRule creates a file matcher function for a specific rule
func createFileMatcherForRule(rule types.Rule) FileMatcher {
	tech := rule.Tech
	patterns := rule.Files

	return func(fileList []types.File, currentPath, basePath string) (string, string, bool) {
		for _, pattern := range patterns {
			if matched, matchedPath := matchPattern(pattern, fileList, currentPath); matched {
				return tech, matchedPath, true
			}
		}
		return "", "", false
	}
}

func matchPattern(pattern string, fileList []types.File, currentPath string) (bool, string) {
	if isDirectoryPattern(pattern) {
		return matchDirectoryPattern(pattern, currentPath)
	}
	return matchFilePattern(pattern, fileList)
}

func isDirectoryPattern(pattern string) bool {
	return strings.Contains(pattern, "/")
}

func matchDirectoryPattern(pattern, currentPath string) (bool, string) {
	if strings.HasSuffix(currentPath, pattern) {
		return true, pattern
	}
	return false, ""
}

func matchFilePattern(pattern string, fileList []types.File) (bool, string) {
	for _, file := range fileList {
		if matched := matchFileName(pattern, file.Name); matched {
			return true, file.Name
		}
	}
	return false, ""
}

func matchFileName(pattern, fileName string) bool {
	if isRegexPattern(pattern) {
		return matchRegex(pattern, fileName)
	}
	return fileName == pattern
}

func matchRegex(pattern, fileName string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(fileName)
}

// isRegexPattern checks if a string contains regex special characters
func isRegexPattern(pattern string) bool {
	regexChars := []string{"*", "+", "?", "[", "]", "(", ")", "|", "^", "$", "\\", "."}
	for _, char := range regexChars {
		if strings.Contains(pattern, char) {
			return true
		}
	}
	return false
}

// MatchFiles runs all file matchers and returns matched techs
// Returns a map of tech -> reasons (like original TypeScript)
func MatchFiles(files []types.File, currentPath, basePath string) map[string][]string {
	matched := make(map[string][]string)

	for _, matcher := range fileMatchers {
		if tech, file, ok := matcher(files, currentPath, basePath); ok {
			// Only add if not already matched (like original: if (matched.has(res[0].tech)) { continue; })
			if _, exists := matched[tech]; !exists {
				matched[tech] = []string{"matched file: " + file}
			}
		}
	}

	return matched
}
