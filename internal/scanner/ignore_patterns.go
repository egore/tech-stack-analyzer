package scanner

// IgnorePatterns defines directories and files that should be excluded from scanning
// organized by ecosystem/language for better maintainability

// GetIgnorePatterns returns all ignore patterns organized by category
func GetIgnorePatterns() map[string][]string {
	return map[string][]string{
		"common": {
			"__fixtures__", "__snapshots__",
		},
		"git": {
			".git", ".gitlab", ".svn",
			// Note: .github is NOT ignored - needed to detect GitHub Actions
		},
		"python": {
			"venv", ".venv", "__pycache__",
			".pytest_cache", ".ruff_cache", ".mypy_cache",
			".tox", ".coverage", ".hypothesis", ".eggs",
			"*.egg-info", ".pytype",
		},
		"nodejs": {
			"node_modules", ".npm", ".yarn", ".pnp",
			".next", ".nuxt", ".vuepress",
		},
		"java": {
			".gradle", ".m2",
		},
		"dotnet": {
			"obj", "packages",
		},
		"ruby": {
			".bundle",
		},
		"rust": {
			"target", // Rust also uses target directory
		},
		"go": {
			// Go doesn't have many standard ignore patterns
		},
		"terraform": {
			".terraform", "terraform.tfstate.d",
		},
		"docker": {
			".docker",
		},
		"cloud": {
			".azure", ".azure-pipelines", ".vercel",
		},
		"cache": {
			".cache", ".artifacts", ".assets",
		},
		"ide": {
			".vscode", ".idea", ".devcontainer",
		},
		"ci": {
			".semgrep", ".release", ".changelog",
		},
		"serverless": {
			".serverless", ".fusebox",
		},
		"database": {
			".dynamodb",
		},
		"misc": {
			".log", ".metadata", ".react-email",
		},
	}
}

// GetFlatIgnoreList returns a flattened list of all ignore patterns
func GetFlatIgnoreList() []string {
	patterns := GetIgnorePatterns()
	var flat []string

	// Flatten all patterns into a single list
	for _, categoryPatterns := range patterns {
		flat = append(flat, categoryPatterns...)
	}

	return flat
}

// GetIgnorePatternsForLanguage returns ignore patterns for a specific language/ecosystem
func GetIgnorePatternsForLanguage(language string) []string {
	patterns := GetIgnorePatterns()

	// Always include common patterns
	result := make([]string, len(patterns["common"]))
	copy(result, patterns["common"])

	// Add language-specific patterns
	if langPatterns, exists := patterns[language]; exists {
		result = append(result, langPatterns...)
	}

	// Add git patterns (always useful)
	result = append(result, patterns["git"]...)

	return result
}
