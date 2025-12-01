package config

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Settings holds all scanner configuration
type Settings struct {
	// Output settings
	OutputFile  string
	PrettyPrint bool

	// Scan behavior
	ExcludeDirs []string
	Aggregate   string

	// Logging
	LogLevel  logrus.Level
	LogFormat string // "text" or "json"
}

// DefaultSettings returns default configuration
func DefaultSettings() *Settings {
	return &Settings{
		OutputFile:  "",
		PrettyPrint: true,
		ExcludeDirs: []string{},
		Aggregate:   "",
		LogLevel:    logrus.InfoLevel,
		LogFormat:   "text",
	}
}

// LoadSettings creates settings from defaults and applies environment variable overrides
func LoadSettings() *Settings {
	settings := DefaultSettings()

	// Apply environment variable overrides
	if outputFile := os.Getenv("STACK_ANALYZER_OUTPUT"); outputFile != "" {
		settings.OutputFile = outputFile
	}

	if excludeDirs := os.Getenv("STACK_ANALYZER_EXCLUDE_DIRS"); excludeDirs != "" {
		settings.ExcludeDirs = strings.Split(excludeDirs, ",")
		for i, dir := range settings.ExcludeDirs {
			settings.ExcludeDirs[i] = strings.TrimSpace(dir)
		}
	}

	if aggregate := os.Getenv("STACK_ANALYZER_AGGREGATE"); aggregate != "" {
		settings.Aggregate = aggregate
	}

	if pretty := os.Getenv("STACK_ANALYZER_PRETTY"); pretty != "" {
		settings.PrettyPrint = strings.ToLower(pretty) == "true"
	}

	// Logging settings
	if logLevel := os.Getenv("STACK_ANALYZER_LOG_LEVEL"); logLevel != "" {
		if level, err := logrus.ParseLevel(logLevel); err == nil {
			settings.LogLevel = level
		}
	}

	if logFormat := os.Getenv("STACK_ANALYZER_LOG_FORMAT"); logFormat != "" {
		settings.LogFormat = logFormat
	}

	return settings
}

// ConfigureLogger sets up the global logger based on settings
func (s *Settings) ConfigureLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(s.LogLevel)

	// Set log format
	if s.LogFormat == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set output to stderr by default (can be changed to file later)
	logger.SetOutput(os.Stderr)

	return logger
}

// Validate checks if settings are valid
func (s *Settings) Validate() error {
	// TODO: Add validation logic
	// - Check if output directory exists/writable
	// - Validate aggregate fields
	// - Validate max depth is reasonable
	return nil
}
