package template

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ExportResourceAsHCL extracts configuration from ResourceData and writes it to a file.
// This is a quick-and-dirty implementation for exporting Terraform resources to HCL.
// The whole package should be extracted to avoid circular dependencies.
func ExportResourceAsHCL(d *schema.ResourceData, resourceType string) error {
	test := &testing.T{}
	store := InitializeTemplateStore(test)

	var schemaResource *schema.Resource = store.SchemaMap[resourceType]

	// extract values from ResourceData into a map
	configMap := extractResourceDataToMap(d, schemaResource)

	// apply transformations to the config map
	transformerRegistry := NewTransformerRegistry()
	transformerRegistry.RegisterDefaultTransformers()

	if transformer, ok := transformerRegistry.transformers[resourceType]; ok {
		configMap = transformer.TransformData(configMap)
	}

	// create a new composition builder
	builder := store.NewBuilder()

	// add the resource with the extracted configuration
	builder.AddResource(resourceType, configMap)

	// render the composition to HCL
	hcl, err := builder.Render(test)
	if err != nil {
		return fmt.Errorf("failed to render HCL: %w", err)
	}

	// format the HCL for consistent indentation
	hcl = normalizeHCL(hcl)

	// Add required providers header
	hcl = addProvidersHeader(hcl)

	exportFilePath := os.Getenv("AIVEN_PLUGIN_EXPORT_FILE_PATH")
	if exportFilePath == "" {
		return nil
	}

	if err = os.MkdirAll(filepath.Dir(exportFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// write HCL to file
	if err = os.WriteFile(exportFilePath, []byte(hcl), 0644); err != nil {
		return fmt.Errorf("failed to write HCL to file: %w", err)
	}

	return nil
}

func addProvidersHeader(hcl string) string {
	header := `terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
    }
  }
}

variable "aiven_api_token" {
  description = "Aiven authentication token"
  type        = string
  sensitive   = true
}

provider "aiven" {
  api_token = var.aiven_api_token
}

`
	return header + hcl
}

// extractResourceDataToMap converts ResourceData contents to a map that can be used with the composition builder
func extractResourceDataToMap(d *schema.ResourceData, resource *schema.Resource) map[string]any {
	result := make(map[string]any)

	// for identifying the resource in the template
	result["resource_name"] = "test"

	// process each field in the schema
	for fieldName, fieldSchema := range resource.Schema {
		if fieldSchema.Sensitive {
			continue
		}

		// skip computed-only fields
		if fieldSchema.Computed && !fieldSchema.Optional && !fieldSchema.Required {
			continue
		}

		// get value from ResourceData
		value, exists := d.GetOk(fieldName)
		if !exists {
			continue
		}

		// process the value based on its type
		processedValue := processResourceDataValue(value, fieldSchema)

		// only add non-nil values
		if processedValue != nil {
			result[fieldName] = processedValue
		}
	}

	return result
}

// processResourceDataValue converts ResourceData values to appropriate Go types for the template
func processResourceDataValue(value any, sc *schema.Schema) any {
	if value == nil {
		return nil
	}

	switch sc.Type {
	case schema.TypeList, schema.TypeSet:
		if sc.Elem == nil {
			return nil
		}

		// Convert to slice
		values := make([]any, 0)

		switch val := value.(type) {
		case []any:
			// Array is already in the right format
			if len(val) == 0 {
				return nil
			}

			// For list/set of objects
			if resource, ok := sc.Elem.(*schema.Resource); ok {
				for _, item := range val {
					if mapItem, ok := item.(map[string]any); ok {
						nestedResult := make(map[string]any)

						// Process each field in the nested resource
						for nestedName, nestedSchema := range resource.Schema {
							if nestedValue, exists := mapItem[nestedName]; exists && nestedValue != nil {
								processed := processResourceDataValue(nestedValue, nestedSchema)
								if processed != nil {
									nestedResult[nestedName] = processed
								}
							}
						}

						if len(nestedResult) > 0 {
							values = append(values, nestedResult)
						}
					}
				}
			} else {
				// for list/set of primitives
				for _, item := range val {
					if item != nil {
						values = append(values, item)
					}
				}
			}
		}

		if len(values) == 0 {
			return nil
		}
		return values

	case schema.TypeMap:
		if mapVal, ok := value.(map[string]any); ok {
			if len(mapVal) == 0 {
				return nil
			}

			processedMap := make(map[string]any)

			// Process each value in the map
			for k, v := range mapVal {
				if v != nil {
					processedMap[k] = v
				}
			}

			if len(processedMap) == 0 {
				return nil
			}
			return processedMap
		}
		return nil

	case schema.TypeBool, schema.TypeInt, schema.TypeFloat, schema.TypeString:
		// simple types can be returned as is
		return value

	default:
		// for other types, return as-is and let the renderer handle it
		return value
	}
}
