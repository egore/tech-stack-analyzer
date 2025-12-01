package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/petrarca/tech-stack-analyzer/internal/rules"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	outputFormat string
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display information about rules and types",
	Long:  `Display information about component types, available technologies, and rule details.`,
}

var componentTypesCmd = &cobra.Command{
	Use:   "component-types",
	Short: "List all component types",
	Long:  `List all technology types that create components vs those that don't.`,
	Run:   runComponentTypes,
}

var techsCmd = &cobra.Command{
	Use:   "techs",
	Short: "List all available technologies",
	Long:  `List all technology names from the embedded rules.`,
	Run:   runTechs,
}

var ruleCmd = &cobra.Command{
	Use:   "rule [tech-name]",
	Short: "Show rule details for a specific technology",
	Long:  `Display the complete rule definition for a given technology name.`,
	Args:  cobra.ExactArgs(1),
	Run:   runRule,
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.AddCommand(componentTypesCmd)
	infoCmd.AddCommand(techsCmd)
	infoCmd.AddCommand(ruleCmd)

	// Add format flag to rule command
	ruleCmd.Flags().StringVarP(&outputFormat, "format", "f", "yaml", "Output format: yaml or json")
}

func runComponentTypes(cmd *cobra.Command, args []string) {
	// Load types configuration
	typesConfig, err := rules.LoadTypesConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading types config: %v\n", err)
		os.Exit(1)
	}

	// Separate component and non-component types
	var componentTypes []string
	var nonComponentTypes []string

	for typeName, typeDef := range typesConfig.Types {
		if typeDef.IsComponent {
			componentTypes = append(componentTypes, typeName)
		} else {
			nonComponentTypes = append(nonComponentTypes, typeName)
		}
	}

	// Sort for consistent output
	sort.Strings(componentTypes)
	sort.Strings(nonComponentTypes)

	fmt.Println("=== Component Types (create components) ===")
	for _, t := range componentTypes {
		if desc, ok := typesConfig.Types[t]; ok && desc.Description != "" {
			fmt.Printf("%s - %s\n", t, desc.Description)
		} else {
			fmt.Println(t)
		}
	}

	fmt.Println("\n=== Non-Component Types (tools/libraries only) ===")
	for _, t := range nonComponentTypes {
		if desc, ok := typesConfig.Types[t]; ok && desc.Description != "" {
			fmt.Printf("%s - %s\n", t, desc.Description)
		} else {
			fmt.Println(t)
		}
	}
}

func runTechs(cmd *cobra.Command, args []string) {
	// Load rules
	allRules, err := rules.LoadEmbeddedRules()
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

	// Extract unique tech names
	techSet := make(map[string]bool)
	for _, rule := range allRules {
		techSet[rule.Tech] = true
	}

	// Sort and print
	techs := make([]string, 0, len(techSet))
	for tech := range techSet {
		techs = append(techs, tech)
	}
	sort.Strings(techs)

	for _, tech := range techs {
		fmt.Println(tech)
	}

	fmt.Fprintf(cmd.OutOrStderr(), "\nTotal: %d technologies\n", len(techs))
}

func runRule(cmd *cobra.Command, args []string) {
	techName := args[0]

	// Load rules
	allRules, err := rules.LoadEmbeddedRules()
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

	// Find the rule
	var foundRule *types.Rule
	for i := range allRules {
		if allRules[i].Tech == techName {
			foundRule = &allRules[i]
			break
		}
	}

	if foundRule == nil {
		log.Fatalf("Rule not found: %s", techName)
	}

	// Output based on format
	var output []byte
	switch outputFormat {
	case "json":
		output, err = json.MarshalIndent(foundRule, "", "  ")
	case "yaml":
		output, err = yaml.Marshal(foundRule)
	default:
		log.Fatalf("Invalid format: %s. Use 'yaml' or 'json'", outputFormat)
	}

	if err != nil {
		log.Fatalf("Failed to marshal rule: %v", err)
	}

	fmt.Println(string(output))
}
