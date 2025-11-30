package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/petrarca/tech-stack-analyzer/internal/scanner"
)

func main() {
	var (
		outputFile    = flag.String("output", "", "Output file (default: stdout)")
		externalRules = flag.String("rules", "", "External rules directory (not implemented)")
		excludeDirs   = flag.String("exclude-dir", "", "Comma-separated directories to exclude")
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

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Fatalf("Path does not exist: %s", absPath)
	}

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
	s, err := scanner.NewScannerWithExcludes(absPath, excludeList)
	if err != nil {
		log.Fatalf("Failed to create scanner: %v", err)
	}

	// Scan project
	log.Printf("Scanning %s...", absPath)
	payload, err := s.Scan()
	if err != nil {
		log.Fatalf("Failed to scan: %v", err)
	}

	// Convert to JSON
	var jsonData []byte
	if *prettyPrint {
		jsonData, err = json.MarshalIndent(payload, "", "  ")
	} else {
		jsonData, err = json.Marshal(payload)
	}

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
