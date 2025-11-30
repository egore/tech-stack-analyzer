package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/petrarca/tech-stack-analyzer/internal/aggregator"
	"github.com/petrarca/tech-stack-analyzer/internal/scanner"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
	"github.com/spf13/cobra"
)

var (
	outputFile  string
	excludeDirs string
	aggregate   string
	prettyPrint bool
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
  stack-analyzer scan --exclude-dir vendor,node_modules /path/to/project`,
	Args: cobra.MaximumNArgs(1),
	Run:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	scanCmd.Flags().StringVar(&excludeDirs, "exclude-dir", "", "Comma-separated directories to exclude")
	scanCmd.Flags().StringVar(&aggregate, "aggregate", "", "Aggregate fields: tech,techs,languages,licenses")
	scanCmd.Flags().BoolVar(&prettyPrint, "pretty", true, "Pretty print JSON output")
}

func runScan(cmd *cobra.Command, args []string) {
	// Get path from args or use current directory
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// Clean and validate project path
	path = strings.TrimSpace(path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Invalid path: %v", err)
	}

	// Check if path exists and determine if it's a file or directory
	fileInfo, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		log.Fatalf("Path does not exist: %s", absPath)
	}
	isFile := !fileInfo.IsDir()

	// Parse exclude dirs
	var excludeList []string
	if excludeDirs != "" {
		excludeList = strings.Split(excludeDirs, ",")
		for i, dir := range excludeList {
			excludeList[i] = strings.TrimSpace(dir)
		}
	}

	// Initialize scanner
	scannerPath := absPath
	if isFile {
		scannerPath = filepath.Dir(absPath)
	}

	s, err := scanner.NewScannerWithExcludes(scannerPath, excludeList)
	if err != nil {
		log.Fatalf("Failed to create scanner: %v", err)
	}

	// Scan project or file
	var payload interface{}
	if isFile {
		log.Printf("Scanning file %s...", absPath)
		payload, err = s.ScanFile(filepath.Base(absPath))
	} else {
		log.Printf("Scanning %s...", absPath)
		payload, err = s.Scan()
	}

	if err != nil {
		log.Fatalf("Failed to scan: %v", err)
	}

	// Generate output (aggregated or full payload)
	jsonData, err := generateOutput(payload, aggregate, prettyPrint)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Write output
	if outputFile != "" {
		err = os.WriteFile(outputFile, jsonData, 0644)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		log.Printf("Results written to %s", outputFile)
	} else {
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

		// Validate fields
		validFields := map[string]bool{"tech": true, "techs": true, "languages": true, "licenses": true}
		for _, field := range fields {
			if !validFields[field] {
				return nil, fmt.Errorf("invalid aggregate field: %s. Valid fields: tech, techs, languages, licenses", field)
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
