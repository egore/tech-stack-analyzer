package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/petrarca/tech-stack-analyzer/internal/aggregator"
	"github.com/petrarca/tech-stack-analyzer/internal/config"
	"github.com/petrarca/tech-stack-analyzer/internal/scanner"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	settings *config.Settings
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a project or file for technology stack",
	Long: `Scan analyzes a project directory or single file to detect technologies,
frameworks, databases, and services used in your codebase.

Examples:
  stack-analyzer scan /path/to/project
  stack-analyzer scan /path/to/pom.xml
  stack-analyzer scan --aggregate techs,languages /path/to/project
  stack-analyzer scan --aggregate all /path/to/project
  stack-analyzer scan --exclude vendor,node_modules /path/to/project
  stack-analyzer scan --exclude "**/__tests__/**" --exclude "*.log" /path/to/project`,
	Args: cobra.MaximumNArgs(1),
	Run:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Initialize settings with defaults and environment variables
	settings = config.LoadSettings()

	// Store environment variable values for flag defaults
	outputFile := settings.OutputFile
	aggregate := settings.Aggregate
	prettyPrint := settings.PrettyPrint
	verbose := settings.Verbose
	logLevel := settings.LogLevel.String()
	logFormat := settings.LogFormat

	// Set up flags with defaults from environment variables
	scanCmd.Flags().StringVarP(&settings.OutputFile, "output", "o", outputFile, "Output file path (default: stdout)")
	scanCmd.Flags().StringVar(&settings.Aggregate, "aggregate", aggregate, "Aggregate fields: tech,techs,languages,licenses,dependencies,all")
	scanCmd.Flags().BoolVar(&settings.PrettyPrint, "pretty", prettyPrint, "Pretty print JSON output")
	scanCmd.Flags().BoolVarP(&settings.Verbose, "verbose", "v", verbose, "Show detailed progress information")

	// Exclude patterns - support multiple flags or comma-separated values
	scanCmd.Flags().StringSliceVar(&settings.ExcludeDirs, "exclude", settings.ExcludeDirs, "Patterns to exclude (supports glob patterns, can be specified multiple times)")

	// Logging flags - use defaults from environment variables
	scanCmd.Flags().String("log-level", logLevel, "Log level: trace, debug, info, warn, error, fatal, panic")
	scanCmd.Flags().String("log-format", logFormat, "Log format: text or json")
}

func runScan(cmd *cobra.Command, args []string) {
	// Get logging flags and configure logger
	logLevel, _ := cmd.Flags().GetString("log-level")
	logFormat, _ := cmd.Flags().GetString("log-format")

	// Update settings with flag values
	if level, err := logrus.ParseLevel(logLevel); err == nil {
		settings.LogLevel = level
	}
	settings.LogFormat = logFormat

	// Configure logger
	logger := settings.ConfigureLogger()
	logger.WithFields(logrus.Fields{
		"version": "1.0.0",
		"command": "scan",
	}).Info("Starting Tech Stack Analyzer")

	// Get path from args or use current directory
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// Clean and validate project path
	path = strings.TrimSpace(path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		logger.WithError(err).Fatal("Invalid path")
	}

	// Check if path exists and determine if it's a file or directory
	fileInfo, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		logger.WithField("path", absPath).Fatal("Path does not exist")
	}
	isFile := !fileInfo.IsDir()

	// Get the exclude flag value (already parsed as StringSlice by cobra)
	excludeList, _ := cmd.Flags().GetStringSlice("exclude")

	// Trim whitespace from each pattern
	for i, pattern := range excludeList {
		excludeList[i] = strings.TrimSpace(pattern)
	}

	// Update settings with actual flag values
	settings.ExcludeDirs = excludeList

	// Validate settings
	if err := settings.Validate(); err != nil {
		logger.WithError(err).Fatal("Invalid settings")
	}

	// Initialize scanner
	scannerPath := absPath
	if isFile {
		scannerPath = filepath.Dir(absPath)
	}

	logger.WithFields(logrus.Fields{
		"path":         scannerPath,
		"exclude_dirs": settings.ExcludeDirs,
		"verbose":      settings.Verbose,
	}).Info("Initializing scanner")

	s, err := scanner.NewScannerWithExcludes(scannerPath, settings.ExcludeDirs, settings.Verbose)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create scanner")
	}

	// Scan project or file
	var payload interface{}
	if isFile {
		logger.WithField("file", absPath).Info("Scanning file")
		payload, err = s.ScanFile(filepath.Base(absPath))
	} else {
		logger.WithField("directory", absPath).Info("Scanning directory")
		payload, err = s.Scan()
	}

	if err != nil {
		logger.WithError(err).Fatal("Failed to scan")
	}

	// Generate output (aggregated or full payload)
	logger.WithFields(logrus.Fields{
		"aggregate":    settings.Aggregate,
		"pretty_print": settings.PrettyPrint,
	}).Debug("Generating output")

	jsonData, err := generateOutput(payload, settings.Aggregate, settings.PrettyPrint)
	if err != nil {
		logger.WithError(err).Fatal("Failed to marshal JSON")
	}

	// Write output
	if settings.OutputFile != "" {
		logger.WithField("output_file", settings.OutputFile).Info("Writing results to file")
		err = os.WriteFile(settings.OutputFile, jsonData, 0644)
		if err != nil {
			logger.WithError(err).Fatal("Failed to write output file")
		}
		logger.WithField("file", settings.OutputFile).Info("Results written successfully")
	} else {
		logger.Debug("Outputting to stdout")
		fmt.Println(string(jsonData))
	}
}

func generateOutput(payload interface{}, aggregateFields string, prettyPrint bool) ([]byte, error) {
	var result interface{}

	if aggregateFields != "" {
		// Parse aggregate fields
		fields := strings.Split(aggregateFields, ",")
		for i, field := range fields {
			fields[i] = strings.TrimSpace(field)
		}

		// Handle "all" as special case - aggregate all available fields
		if len(fields) == 1 && fields[0] == "all" {
			fields = []string{"tech", "techs", "languages", "licenses", "dependencies"}
		}

		// Validate fields
		validFields := map[string]bool{"tech": true, "techs": true, "languages": true, "licenses": true, "dependencies": true}
		for _, field := range fields {
			if !validFields[field] {
				return nil, fmt.Errorf("invalid aggregate field: %s. Valid fields: tech, techs, languages, licenses, dependencies, all", field)
			}
		}

		// Create aggregator and aggregate
		agg := aggregator.NewAggregator(fields)
		result = agg.Aggregate(payload.(*types.Payload))
	} else {
		result = payload
	}

	// Marshal to JSON
	if prettyPrint {
		return json.MarshalIndent(result, "", "  ")
	}
	return json.Marshal(result)
}
