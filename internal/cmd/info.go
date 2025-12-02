package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/petrarca/tech-stack-analyzer/internal/config"
	"github.com/petrarca/tech-stack-analyzer/internal/rules"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
	"github.com/petrarca/tech-stack-analyzer/internal/util"
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

	// Add format flag to all info subcommands with separate variables and validation
	setupFormatFlag(componentTypesCmd, "text", runComponentTypes)
	setupFormatFlag(techsCmd, "text", runTechs)
	setupFormatFlag(ruleCmd, "yaml", runRule)
}

// setupFormatFlag configures format flag and validation for a command
func setupFormatFlag(cmd *cobra.Command, defaultFormat string, runFunc func(*cobra.Command, []string)) {
	var format string
	cmd.Flags().StringVarP(&format, "format", "f", defaultFormat, "Output format: text, yaml, or json")
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		format = util.NormalizeFormat(format)
		if err := util.ValidateOutputFormat(format); err != nil {
			log.Fatalf("Invalid format: %v", err)
		}
		outputFormat = format
	}
	cmd.Run = runFunc
}

func runComponentTypes(cmd *cobra.Command, args []string) {
	// Load types configuration
	typesConfig, err := config.LoadTypesConfig()
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

	// Output based on format using validation utility
	switch util.NormalizeFormat(outputFormat) {
	case "json":
		result := map[string]interface{}{
			"component_types":     componentTypes,
			"non_component_types": nonComponentTypes,
		}
		outputAndMarshal(result, "json")
	case "yaml":
		result := map[string]interface{}{
			"component_types":     componentTypes,
			"non_component_types": nonComponentTypes,
		}
		outputAndMarshal(result, "yaml")
	default: // text format
		printTypeList(typesConfig, componentTypes, "Component Types (create components)")
		printTypeList(typesConfig, nonComponentTypes, "Non-Component Types (tools/libraries only)")
	}
}

// printTypeList prints a formatted list of types with descriptions
func printTypeList(typesConfig *types.TypesConfig, types []string, title string) {
	fmt.Printf("=== %s ===\n", title)
	for _, t := range types {
		if desc, ok := typesConfig.Types[t]; ok && desc.Description != "" {
			fmt.Printf("%s - %s\n", t, desc.Description)
		} else {
			fmt.Println(t)
		}
	}
	if title == "Component Types (create components)" {
		fmt.Println()
	}
}

func runTechs(cmd *cobra.Command, args []string) {
	// Load rules
	allRules, err := rules.LoadEmbeddedRules()
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

	// Create a map of tech key to rule for quick lookup
	ruleMap := make(map[string]*types.Rule)
	for i := range allRules {
		ruleMap[allRules[i].Tech] = &allRules[i]
	}

	// Extract and sort tech keys
	techKeys := make([]string, 0, len(ruleMap))
	for tech := range ruleMap {
		techKeys = append(techKeys, tech)
	}
	sort.Strings(techKeys)

	// Output based on format using validation utility
	switch util.NormalizeFormat(outputFormat) {
	case "json":
		technologies := buildTechList(ruleMap, techKeys)
		outputAndMarshal(map[string]interface{}{"technologies": technologies}, "json")
	case "yaml":
		technologies := buildTechList(ruleMap, techKeys)
		outputAndMarshal(map[string]interface{}{"technologies": technologies}, "yaml")
	default: // text format
		for _, tech := range techKeys {
			fmt.Println(tech)
		}
		fmt.Fprintf(cmd.OutOrStderr(), "\nTotal: %d technologies\n", len(techKeys))
	}
}

// buildTechList creates a slice of tech info maps for JSON/YAML output
func buildTechList(ruleMap map[string]*types.Rule, techKeys []string) []map[string]interface{} {
	technologies := make([]map[string]interface{}, 0, len(techKeys))
	for _, techKey := range techKeys {
		rule := ruleMap[techKey]
		techInfo := map[string]interface{}{
			"name": rule.Name,
			"tech": techKey,
			"type": rule.Type,
		}
		// Only include description if it's not empty
		if rule.Description != "" {
			techInfo["description"] = rule.Description
		}
		// Only include properties if they exist and are not empty
		if len(rule.Properties) > 0 {
			techInfo["properties"] = rule.Properties
		}
		technologies = append(technologies, techInfo)
	}
	return technologies
}

// outputAndMarshal handles common marshaling and output logic
func outputAndMarshal(data interface{}, format string) {
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(data, "", "  ")
	case "yaml":
		output, err = yaml.Marshal(data)
	}

	if err != nil {
		log.Fatalf("Failed to marshal data: %v", err)
	}

	fmt.Println(string(output))
}

func runRule(cmd *cobra.Command, args []string) {
	techName := args[0]

	// Load rules
	allRules, err := rules.LoadEmbeddedRules()
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

	// Find the rule
	foundRule := findRuleByTech(allRules, techName)
	if foundRule == nil {
		log.Fatalf("Rule not found: %s", techName)
	}

	// Output based on format using validation utility
	outputAndMarshal(foundRule, util.NormalizeFormat(outputFormat))
}

// findRuleByTech searches for a rule by tech name
func findRuleByTech(allRules []types.Rule, techName string) *types.Rule {
	for i := range allRules {
		if allRules[i].Tech == techName {
			return &allRules[i]
		}
	}
	return nil
}
