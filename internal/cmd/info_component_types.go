package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/petrarca/tech-stack-analyzer/internal/config"
	"github.com/spf13/cobra"
)

var componentTypesFormat string

var componentTypesCmd = &cobra.Command{
	Use:   "component-types",
	Short: "List all component types",
	Long:  `List all technology types that create components vs those that don't.`,
	Run:   runComponentTypes,
}

func init() {
	setupFormatFlag(componentTypesCmd, &componentTypesFormat)
}

// ComponentTypesResult is the output for the component-types command
type ComponentTypesResult struct {
	ComponentTypes    []string `json:"component_types"`
	NonComponentTypes []string `json:"non_component_types"`
}

func (r *ComponentTypesResult) ToJSON() interface{} {
	return r
}

func (r *ComponentTypesResult) ToText(w io.Writer) {
	fmt.Fprintln(w, "=== Component Types (create components) ===")
	for _, t := range r.ComponentTypes {
		fmt.Fprintln(w, t)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "=== Non-Component Types (tools/libraries only) ===")
	for _, t := range r.NonComponentTypes {
		fmt.Fprintln(w, t)
	}
}

func runComponentTypes(cmd *cobra.Command, args []string) {
	typesConfig, err := config.LoadTypesConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading types config: %v\n", err)
		os.Exit(1)
	}

	var componentTypes, nonComponentTypes []string
	for typeName, typeDef := range typesConfig.Types {
		if typeDef.IsComponent {
			componentTypes = append(componentTypes, typeName)
		} else {
			nonComponentTypes = append(nonComponentTypes, typeName)
		}
	}

	sort.Strings(componentTypes)
	sort.Strings(nonComponentTypes)

	result := &ComponentTypesResult{
		ComponentTypes:    componentTypes,
		NonComponentTypes: nonComponentTypes,
	}
	Output(result, componentTypesFormat)
}
