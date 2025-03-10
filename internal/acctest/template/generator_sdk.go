package template

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
func (g *SDKTemplateGenerator) GenerateTemplate(schemaObj interface{}, resourceType string, kind ResourceKind) string {
	// For SDK resources, we expect a *schema.Resource
	resource, ok := schemaObj.(*schema.Resource)
	if !ok {
		// If not a *schema.Resource, return a basic template with an error comment
		var b strings.Builder
		_, _ = fmt.Fprintf(&b, "%s %q %q {\n", kind, resourceType, "{{ required .resource_name }}")
		b.WriteString("  # Error: SDK generator received non-SDK schema\n")
		b.WriteString("}")
		return b.String()
	}

	fields, err := g.extractor.ExtractFields(resource)
	if err != nil {
		// Handle extraction error
		var b strings.Builder
		_, _ = fmt.Fprintf(&b, "%s %q %q {\n", kind, resourceType, "{{ required .resource_name }}")
		b.WriteString(fmt.Sprintf("  # Error extracting fields: %v\n", err))
		b.WriteString("}")
		return b.String()
	}

	return g.generateTemplateFromFields(fields, resource, resourceType, kind)
}

func (g *SDKTemplateGenerator) generateTemplateFromFields(fields []TemplateField, r *schema.Resource, resourceType string, kind ResourceKind) string {
	var b strings.Builder

	// Header differs based on kind
	_, _ = fmt.Fprintf(&b, "%s %q %q {\n", kind, resourceType, "{{ required .resource_name }}")

	// Handle regular fields
	for _, field := range fields {
		g.renderer.RenderField(&b, field, 1, TemplatePath{})
	}

	// Handle timeouts
	if r.Timeouts != nil {
		timeoutsConfig := g.extractTimeoutsConfig(r.Timeouts)

		// Only generate the timeouts block if at least one timeout is configured
		if timeoutsConfig.Create || timeoutsConfig.Read || timeoutsConfig.Update || timeoutsConfig.Delete {
			g.renderer.RenderTimeouts(&b, 1, timeoutsConfig)
		}
	}

	// Handle depends_on
	if _, hasDependsOn := r.Schema["depends_on"]; hasDependsOn {
		g.renderer.RenderDependsOn(&b, 1)
	}

	b.WriteString("}")

	return b.String()
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
