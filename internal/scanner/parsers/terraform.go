package parsers

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

// TerraformParser handles Terraform-specific file parsing (.tf and .terraform.lock.hcl)
type TerraformParser struct{}

// NewTerraformParser creates a new Terraform parser
func NewTerraformParser() *TerraformParser {
	return &TerraformParser{}
}

// TerraformProvider represents a provider in terraform.lock.hcl
type TerraformProvider struct {
	Name    string
	Version string
}

// ParseTerraformLock parses .terraform.lock.hcl and extracts providers
func (p *TerraformParser) ParseTerraformLock(content string) []TerraformProvider {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(content), "terraform.lock.hcl")
	if diags.HasErrors() {
		return nil
	}

	contentBody := file.Body
	providers := []TerraformProvider{}

	// Extract provider blocks
	if contentBody != nil {
		// Try to parse as blocks
		content, _ := contentBody.Content(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "provider",
					LabelNames: []string{"name"},
				},
			},
		})

		for _, block := range content.Blocks.OfType("provider") {
			if len(block.Labels) > 0 {
				providerName := block.Labels[0]
				version := "latest"

				// Extract version from attributes
				attrs, _ := block.Body.JustAttributes()
				if versionAttr, exists := attrs["version"]; exists {
					if val, diags := versionAttr.Expr.Value(nil); !diags.HasErrors() && val.Type() == cty.String {
						version = val.AsString()
					}
				}

				providers = append(providers, TerraformProvider{
					Name:    providerName,
					Version: version,
				})
			}
		}
	}

	return providers
}

// ParseTerraformResource parses .tf files and extracts resource types
func (p *TerraformParser) ParseTerraformResource(content string) []string {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(content), "resource.tf")
	if diags.HasErrors() {
		return nil
	}

	contentBody := file.Body
	resourceSet := make(map[string]bool)
	var resources []string

	// Extract resource blocks
	if contentBody != nil {
		content, _ := contentBody.Content(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "resource",
					LabelNames: []string{"type", "name"},
				},
			},
		})

		for _, block := range content.Blocks.OfType("resource") {
			if len(block.Labels) > 0 {
				resourceType := block.Labels[0]
				if !resourceSet[resourceType] {
					resources = append(resources, resourceType)
					resourceSet[resourceType] = true
				}
			}
		}
	}

	return resources
}
