package util

import (
	"fmt"

	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// ResourceSchemaToDataSourceSchema converts a resource schema to a data source schema.
// All attributes are made computed by default, except for the specified required fields.
// See https://github.com/hashicorp/terraform-plugin-framework/issues/568.
// Adapted from https://github.com/Doridian/terraform-provider-hexonet/blob/42a3ee1ab12544cde4d31b4897ceacf75d46cb3a/hexonet/utils/schema_helpers.go#L10.
func ResourceSchemaToDataSourceSchema(resourceSchema map[string]resource_schema.Attribute, requiredFields ...string) map[string]datasource_schema.Attribute {
	// Create a map to track required fields
	requiredFieldsMap := make(map[string]bool)
	for _, field := range requiredFields {
		requiredFieldsMap[field] = true
	}

	datasourceSchema := make(map[string]datasource_schema.Attribute)
	for name, srcAttr := range resourceSchema {
		// Default to computed
		optional := false
		required := false
		computed := true

		// If this is a required field, make it required and not computed
		if requiredFieldsMap[name] {
			required = true
			computed = false
		}

		// Handle different attribute types
		switch srcAttrTyped := srcAttr.(type) {
		case resource_schema.StringAttribute:
			datasourceSchema[name] = datasource_schema.StringAttribute{
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				Sensitive:           srcAttrTyped.Sensitive,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.BoolAttribute:
			datasourceSchema[name] = datasource_schema.BoolAttribute{
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				Sensitive:           srcAttrTyped.Sensitive,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.Int64Attribute:
			datasourceSchema[name] = datasource_schema.Int64Attribute{
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				Sensitive:           srcAttrTyped.Sensitive,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.ListAttribute:
			datasourceSchema[name] = datasource_schema.ListAttribute{
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				Sensitive:           srcAttrTyped.Sensitive,
				ElementType:         srcAttrTyped.ElementType,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.MapAttribute:
			datasourceSchema[name] = datasource_schema.MapAttribute{
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				Sensitive:           srcAttrTyped.Sensitive,
				ElementType:         srcAttrTyped.ElementType,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.SetAttribute:
			datasourceSchema[name] = datasource_schema.SetAttribute{
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				Sensitive:           srcAttrTyped.Sensitive,
				ElementType:         srcAttrTyped.ElementType,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.ObjectAttribute:
			datasourceSchema[name] = datasource_schema.ObjectAttribute{
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				Sensitive:           srcAttrTyped.Sensitive,
				AttributeTypes:      srcAttrTyped.AttributeTypes,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.SingleNestedAttribute:
			datasourceSchema[name] = datasource_schema.SingleNestedAttribute{
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				Sensitive:           srcAttrTyped.Sensitive,
				Attributes:          ResourceSchemaToDataSourceSchema(srcAttrTyped.Attributes),
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		default:
			// This would be a compile-time error, but we'll handle it at runtime too
			panic(fmt.Sprintf("unsupported attribute type: %T", srcAttr))
		}
	}

	return datasourceSchema
}
