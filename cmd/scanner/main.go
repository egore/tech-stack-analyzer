package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/petrarca/tech-stack-analyzer/internal/aggregator"
	"github.com/petrarca/tech-stack-analyzer/internal/scanner"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
)

// generateOutput creates JSON output, either aggregated or full payload
func generateOutput(payload *types.Payload, aggregateFields string, prettyPrint bool) ([]byte, error) {
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
		result = agg.Aggregate(payload)
	} else {
		result = payload
	}

	// Marshal to JSON
	if prettyPrint {
		return json.MarshalIndent(result, "", "  ")
	}
	return json.Marshal(result)
}

func main() {
	var (
		outputFile    = flag.String("output", "", "Output file (default: stdout)")
		externalRules = flag.String("rules", "", "External rules directory (not implemented)")
		excludeDirs   = flag.String("exclude-dir", "", "Comma-separated directories to exclude")
		aggregate     = flag.String("aggregate", "", "Aggregate and rollup fields (comma-separated): tech,techs,languages,licenses")
		prettyPrint   = flag.Bool("pretty", true, "Pretty print JSON output")
		validateRules = flag.Bool("validate", false, "Validate rules and exit")
		version       = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	// Get optional path from positional arguments
	var path string
	if len(flag.Args()) > 0 {
		path = flag.Args()[0]
	} else {
		path = "."
	}

	if *version {
		log.Printf("Stack Analyzer v1.0.0")
		log.Printf("Go implementation of the technology stack analyzer")
		os.Exit(0)
	}

	if *validateRules {
		if *externalRules == "" {
			log.Fatal("External rules directory required for validation")
		}
		// TODO: Implement rule validation
		log.Printf("Rule validation not yet implemented")
		os.Exit(0)
	}

	// Clean and validate project path
	path = strings.TrimSpace(path)
	if path == "" {
		path = "."
	}

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
	if *excludeDirs != "" {
		excludeList = strings.Split(*excludeDirs, ",")
		// Trim whitespace from each directory
		for i, dir := range excludeList {
			excludeList[i] = strings.TrimSpace(dir)
		}
	}

	// Initialize scanner with exclude directories
	// For single file scanning, use the parent directory as the base
	scannerPath := absPath
	if isFile {
		scannerPath = filepath.Dir(absPath)
	}

	s, err := scanner.NewScannerWithExcludes(scannerPath, excludeList)
	if err != nil {
		log.Fatalf("Failed to create scanner: %v", err)
	}

	// Scan project or file
	var payload *types.Payload
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
	jsonData, err := generateOutput(payload, *aggregate, *prettyPrint)

	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Write output
	if *outputFile != "" {
		err = os.WriteFile(*outputFile, jsonData, 0644)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		log.Printf("Results written to %s", *outputFile)
	} else {
		fmt.Println(string(jsonData))
	}
}
