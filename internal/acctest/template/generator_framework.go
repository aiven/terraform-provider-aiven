package template

import (
	"fmt"
	"strings"
)

// Ensure FrameworkTemplateGenerator implements TemplateGenerator
var _ TemplateGenerator = &FrameworkTemplateGenerator{}

// FrameworkTemplateGenerator implements TemplateGenerator for Terraform Plugin Framework resources
type FrameworkTemplateGenerator struct {
	extractor *FrameworkFieldExtractor
	renderer  *CommonTemplateRenderer
}

// NewFrameworkTemplateGenerator creates a template generator for Framework resources
func NewFrameworkTemplateGenerator() *FrameworkTemplateGenerator {
	return &FrameworkTemplateGenerator{
		extractor: NewFrameworkFieldExtractor(),
		renderer:  &CommonTemplateRenderer{},
	}
}

// GenerateTemplate generates a Go template string that resembles Terraform HCL configuration
// for a Framework resource or data source.
func (g *FrameworkTemplateGenerator) GenerateTemplate(schema interface{}, resourceType string, kind ResourceKind) (string, error) {
	fields, err := g.extractor.ExtractFields(schema)
	if err != nil {
		return "", fmt.Errorf("error extracting fields: %w", err)
	}

	var (
		hasTimeouts   bool
		hasDependsOn  bool
		timeoutsField TemplateField
		regularFields []TemplateField
	)

	for _, field := range fields {
		switch field.Name {
		case "timeouts":
			hasTimeouts = true
			timeoutsField = field
		case "depends_on":
			hasDependsOn = true
		default:
			regularFields = append(regularFields, field)
		}
	}

	var timeoutsConfig TimeoutsConfig
	if hasTimeouts {
		timeoutsConfig = g.extractTimeoutsConfig(timeoutsField)
	}

	return g.renderer.GenerateTemplate(regularFields, resourceType, kind, timeoutsConfig, hasDependsOn)
}

// extractTimeoutsConfig analyzes the timeouts field to determine which timeouts are configured
func (g *FrameworkTemplateGenerator) extractTimeoutsConfig(timeoutsField TemplateField) TimeoutsConfig {
	config := TimeoutsConfig{
		Create: false,
		Read:   false,
		Update: false,
		Delete: false,
	}

	for _, nestedField := range timeoutsField.NestedFields {
		fieldName := strings.ToLower(nestedField.Name)
		switch fieldName {
		case "create":
			config.Create = true
		case "read":
			config.Read = true
		case "update":
			config.Update = true
		case "delete":
			config.Delete = true
		}
	}

	return config
}
