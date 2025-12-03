package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/petrarca/tech-stack-analyzer/internal/config"
	"github.com/spf13/cobra"
)

var typesFormat string

var typesCmd = &cobra.Command{
	Use:   "types",
	Short: "List all technology types",
	Long:  `List all technology types with their descriptions.`,
	Run:   runTypes,
}

func init() {
	setupFormatFlag(typesCmd, &typesFormat)
}

// TypeInfo represents a single type entry
type TypeInfo struct {
	Name        string `json:"name"`
	IsComponent bool   `json:"is_component"`
	Description string `json:"description"`
}

// TypesResult is the output for the types command
type TypesResult struct {
	Types []TypeInfo `json:"types"`
	Count int        `json:"count"`
}

func (r *TypesResult) ToJSON() interface{} {
	return r
}

func (r *TypesResult) ToText(w io.Writer) {
	fmt.Fprintf(w, "=== Technology Types (%d) ===\n\n", r.Count)
	for _, t := range r.Types {
		componentStr := ""
		if t.IsComponent {
			componentStr = " [component]"
		}
		fmt.Fprintf(w, "%-20s%s\n", t.Name, componentStr)
		if t.Description != "" {
			fmt.Fprintf(w, "  %s\n", t.Description)
		}
	}
}

func runTypes(cmd *cobra.Command, args []string) {
	typesConfig, err := config.LoadTypesConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading types config: %v\n", err)
		os.Exit(1)
	}

	var types []TypeInfo
	for typeName, typeDef := range typesConfig.Types {
		types = append(types, TypeInfo{
			Name:        typeName,
			IsComponent: typeDef.IsComponent,
			Description: typeDef.Description,
		})
	}

	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})

	result := &TypesResult{
		Types: types,
		Count: len(types),
	}
	Output(result, typesFormat)
}
