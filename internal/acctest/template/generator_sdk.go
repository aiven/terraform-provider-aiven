package template

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Ensure SDKTemplateGenerator implements TemplateGenerator
var _ TemplateGenerator = &SDKTemplateGenerator{}

// SDKTemplateGenerator creates Go templates for Terraform SDK resources
type SDKTemplateGenerator struct {
	extractor *SDKFieldExtractor
	renderer  *CommonTemplateRenderer
}

// NewSDKTemplateGenerator creates a new template generator for SDK resources
func NewSDKTemplateGenerator() *SDKTemplateGenerator {
	return &SDKTemplateGenerator{
		extractor: NewSDKFieldExtractor(),
		renderer:  &CommonTemplateRenderer{},
	}
}

// GenerateTemplate implements the TemplateGenerator interface
func (g *SDKTemplateGenerator) GenerateTemplate(schemaObj interface{}, resourceType string, kind ResourceKind) (string, error) {
	// For SDK resources, we expect a *schema.Resource
	resource, ok := schemaObj.(*schema.Resource)
	if !ok {
		return "", fmt.Errorf("SDK generator received non-SDK schema of type %T", schemaObj)
	}

	fields, err := g.extractor.ExtractFields(resource)
	if err != nil {
		return "", fmt.Errorf("error extracting fields: %w", err)
	}

	var timeoutsConfig TimeoutsConfig
	if resource.Timeouts != nil {
		timeoutsConfig = g.extractTimeoutsConfig(resource.Timeouts)
	}

	hasDependsOn := false
	if _, ok := resource.Schema["depends_on"]; ok {
		hasDependsOn = true
	}

	return g.renderer.GenerateTemplate(fields, resourceType, kind, timeoutsConfig, hasDependsOn)
}

// extractTimeoutsConfig extracts timeout configuration from a SDKv2 schema.ResourceTimeout
func (g *SDKTemplateGenerator) extractTimeoutsConfig(timeouts *schema.ResourceTimeout) TimeoutsConfig {
	if timeouts == nil {
		return TimeoutsConfig{}
	}

	return TimeoutsConfig{
		Create: timeouts.Create != nil,
		Read:   timeouts.Read != nil,
		Update: timeouts.Update != nil,
		Delete: timeouts.Delete != nil,
	}
}
