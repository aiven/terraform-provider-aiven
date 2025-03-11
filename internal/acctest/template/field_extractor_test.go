package template

import (
	"testing"

	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestSDKExtractFieldsWithErrors(t *testing.T) {
	t.Parallel()

	sdkExtractor := NewSDKFieldExtractor()

	// Test with invalid schema type
	_, err := sdkExtractor.ExtractFields("not a schema")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a *schema.Resource")

	// Test with nil Schema
	nilSchemaResource := &sdkschema.Resource{
		Schema: nil,
	}
	_, err = sdkExtractor.ExtractFields(nilSchemaResource)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no fields could be extracted from SDK schema")

	// Test with empty Schema
	emptySchemaResource := &sdkschema.Resource{
		Schema: map[string]*sdkschema.Schema{},
	}
	_, err = sdkExtractor.ExtractFields(emptySchemaResource)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no fields could be extracted from SDK schema")
}

func TestFrameworkExtractFieldsWithErrors(t *testing.T) {
	t.Parallel()

	frameworkExtractor := NewFrameworkFieldExtractor()

	// Test with invalid schema type
	_, err := frameworkExtractor.ExtractFields("not a schema")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported schema type")

	// Test with empty resource schema
	emptyResourceSchema := resourceschema.Schema{
		Attributes: map[string]resourceschema.Attribute{},
	}
	_, err = frameworkExtractor.ExtractFields(emptyResourceSchema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no fields could be extracted from framework schema")

	// Test with empty datasource schema
	emptyDataSourceSchema := datasourceschema.Schema{
		Attributes: map[string]datasourceschema.Attribute{},
	}
	_, err = frameworkExtractor.ExtractFields(emptyDataSourceSchema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no fields could be extracted from framework schema")
}

func TestProcessComplexFieldTypes(t *testing.T) {
	t.Parallel()

	// Create a resource with map attributes and lists
	resource := &sdkschema.Resource{
		Schema: map[string]*sdkschema.Schema{
			"tags": {
				Type: sdkschema.TypeMap,
				Elem: &sdkschema.Schema{
					Type: sdkschema.TypeString,
				},
				Optional: true,
			},
			"nested_lists": {
				Type:     sdkschema.TypeList,
				Optional: true,
				Elem: &sdkschema.Resource{
					Schema: map[string]*sdkschema.Schema{
						"inner_list": {
							Type:     sdkschema.TypeList,
							Optional: true,
							Elem: &sdkschema.Schema{
								Type: sdkschema.TypeString,
							},
						},
					},
				},
			},
			"set_field": {
				Type:     sdkschema.TypeSet,
				Optional: true,
				Elem: &sdkschema.Schema{
					Type: sdkschema.TypeString,
				},
			},
		},
	}

	extractor := NewSDKFieldExtractor()
	fields, err := extractor.ExtractFields(resource)

	assert.NoError(t, err)
	assert.Len(t, fields, 3)

	// Verify map field
	var mapField *TemplateField
	for i := range fields {
		if fields[i].Name == "tags" {
			mapField = &fields[i]
			break
		}
	}
	assert.NotNil(t, mapField)
	assert.True(t, mapField.IsMap)

	// Verify nested list
	var nestedField *TemplateField
	for i := range fields {
		if fields[i].Name == "nested_lists" {
			nestedField = &fields[i]
			break
		}
	}
	assert.NotNil(t, nestedField)
	assert.True(t, nestedField.IsCollection)
	assert.True(t, nestedField.IsObject)
	assert.NotEmpty(t, nestedField.NestedFields)

	// Verify set field
	var setField *TemplateField
	for i := range fields {
		if fields[i].Name == "set_field" {
			setField = &fields[i]
			break
		}
	}
	assert.NotNil(t, setField)
	assert.True(t, setField.IsCollection)
	assert.False(t, setField.IsObject)
}
