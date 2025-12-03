package cmd

import (
	"fmt"
	"io"
	"log"

	"github.com/petrarca/tech-stack-analyzer/internal/rules"
	"github.com/petrarca/tech-stack-analyzer/internal/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var ruleFormat string

var ruleCmd = &cobra.Command{
	Use:   "rule [tech-name]",
	Short: "Show rule details for a specific technology",
	Long:  `Display the complete rule definition for a given technology name.`,
	Args:  cobra.ExactArgs(1),
	Run:   runRule,
}

func init() {
	setupFormatFlag(ruleCmd, &ruleFormat)
}

// RuleResult wraps a rule for output
type RuleResult struct {
	Rule *types.Rule
}

func (r *RuleResult) ToJSON() interface{} {
	return r.Rule
}

func (r *RuleResult) ToText(w io.Writer) {
	// For text, use YAML as it's more readable
	data, err := yaml.Marshal(r.Rule)
	if err != nil {
		log.Fatalf("Failed to marshal rule: %v", err)
	}
	fmt.Fprint(w, string(data))
}

func runRule(cmd *cobra.Command, args []string) {
	techName := args[0]

	allRules, err := rules.LoadEmbeddedRules()
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

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

	result := &RuleResult{Rule: foundRule}
	Output(result, ruleFormat)
}
