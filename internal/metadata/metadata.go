package metadata

import (
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ScanMetadata contains information about the scan execution
type ScanMetadata struct {
	Timestamp      string                 `json:"timestamp"`
	ScanPath       string                 `json:"scan_path"`
	SpecVersion    string                 `json:"specVersion"` // Output format specification version
	DurationMs     int64                  `json:"duration_ms,omitempty"`
	FileCount      int                    `json:"file_count,omitempty"`
	DirectoryCount int                    `json:"directory_count,omitempty"`
	ExcludedDirs   []string               `json:"excluded_dirs,omitempty"`
	Git            *GitInfo               `json:"git,omitempty"`
	Properties     map[string]interface{} `json:"properties,omitempty"`
}

// GitInfo contains git repository information
type GitInfo struct {
	Branch    string `json:"branch,omitempty"`
	Commit    string `json:"commit,omitempty"`
	IsDirty   bool   `json:"is_dirty"`
	RemoteURL string `json:"remote_url,omitempty"`
}

// NewScanMetadata creates a new scan metadata instance
func NewScanMetadata(scanPath string, version string, excludedDirs []string) *ScanMetadata {
	absPath, _ := filepath.Abs(scanPath)

	return &ScanMetadata{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		ScanPath:     absPath,
		SpecVersion:  version,
		ExcludedDirs: excludedDirs,
		Git:          GetGitInfo(scanPath),
	}
}

// SetDuration sets the scan duration in milliseconds
func (m *ScanMetadata) SetDuration(duration time.Duration) {
	m.DurationMs = duration.Milliseconds()
}

// SetFileCounts sets the file and directory counts
func (m *ScanMetadata) SetFileCounts(fileCount, dirCount int) {
	m.FileCount = fileCount
	m.DirectoryCount = dirCount
}

// SetProperties sets custom properties from configuration
func (m *ScanMetadata) SetProperties(properties map[string]interface{}) {
	if len(properties) > 0 {
		m.Properties = properties
	}
}

// GetGitInfo retrieves git repository information for the given path
func GetGitInfo(path string) *GitInfo {
	// Check if path is in a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		// Not a git repository
		return nil
	}

	gitInfo := &GitInfo{}

	// Get current branch
	if branch, err := runGitCommand(path, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		gitInfo.Branch = strings.TrimSpace(branch)
	}

	// Get current commit hash
	if commit, err := runGitCommand(path, "rev-parse", "HEAD"); err == nil {
		gitInfo.Commit = strings.TrimSpace(commit)
		// Use short hash (first 7 characters)
		if len(gitInfo.Commit) > 7 {
			gitInfo.Commit = gitInfo.Commit[:7]
		}
	}

	// Check if working directory is dirty
	if status, err := runGitCommand(path, "status", "--porcelain"); err == nil {
		gitInfo.IsDirty = len(strings.TrimSpace(status)) > 0
	}

	// Get remote URL
	if remoteURL, err := runGitCommand(path, "config", "--get", "remote.origin.url"); err == nil {
		gitInfo.RemoteURL = strings.TrimSpace(remoteURL)
	}

	return gitInfo
}

// runGitCommand executes a git command and returns the output
func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
