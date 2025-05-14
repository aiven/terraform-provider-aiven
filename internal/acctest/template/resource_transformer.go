package template

import (
	"strings"
)

// FieldTransformer defines a function that can transform field values
type FieldTransformer func(value any) (any, bool)

// FieldPath represents a path to a field in a nested structure
type FieldPath struct {
	Parts []string
}

// NewFieldPath creates a new field path from a dot-separated string path
func NewFieldPath(path string) FieldPath {
	return FieldPath{
		Parts: strings.Split(path, "."),
	}
}

// String returns the path as a dot-separated string
func (p FieldPath) String() string {
	return strings.Join(p.Parts, ".")
}

// ResourceTransformer manages transformations for a specific resource type
type ResourceTransformer struct {
	resourceType string
	transformers map[string]FieldTransformer // path string -> transformer
}

// NewResourceTransformer creates a new transformer for a resource type
func NewResourceTransformer(resourceType string) *ResourceTransformer {
	return &ResourceTransformer{
		resourceType: resourceType,
		transformers: make(map[string]FieldTransformer),
	}
}

// RegisterTransformer registers a transformer for a specific field path
func (rt *ResourceTransformer) RegisterTransformer(path FieldPath, transformer FieldTransformer) *ResourceTransformer {
	rt.transformers[path.String()] = transformer

	return rt
}

// TransformData applies all registered transformers to the provided data
func (rt *ResourceTransformer) TransformData(data map[string]any) map[string]any {
	return rt.processMap(data, FieldPath{})
}

// processMap processes a map and applies transformers to its fields
func (rt *ResourceTransformer) processMap(data map[string]any, currentPath FieldPath) map[string]any {
	result := make(map[string]any)

	for key, value := range data {
		// Create field path for this key
		fieldPath := currentPath
		if len(fieldPath.Parts) == 0 {
			fieldPath = NewFieldPath(key)
		} else {
			newParts := append([]string{}, fieldPath.Parts...)
			newParts = append(newParts, key)
			fieldPath = FieldPath{Parts: newParts}
		}

		// Check if we have a transformer for this path
		if transformer, ok := rt.transformers[fieldPath.String()]; ok {
			if transformedValue, keep := transformer(value); keep {
				result[key] = rt.processValue(transformedValue, fieldPath)
			}
			continue // Skip further processing for this field
		}

		// Process based on value type
		processedValue := rt.processValue(value, fieldPath)
		if processedValue != nil {
			result[key] = processedValue
		}
	}

	return result
}

// processValue processes a value based on its type
func (rt *ResourceTransformer) processValue(value any, currentPath FieldPath) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		return rt.processMap(v, currentPath)

	case []any:
		return rt.processArray(v, currentPath)

	default:
		// For primitive values, just return them
		return v
	}
}

// processArray processes an array and applies transformers to its elements
func (rt *ResourceTransformer) processArray(array []any, currentPath FieldPath) []any {
	result := make([]any, 0, len(array))

	for _, item := range array {
		// For arrays, we don't add the index to the path, but we process the item
		processed := rt.processValue(item, currentPath)
		if processed != nil {
			result = append(result, processed)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// TransformerRegistry holds transformers for different resource types
type TransformerRegistry struct {
	transformers map[string]*ResourceTransformer // resourceType -> transformer
}

// NewTransformerRegistry creates a new transformer registry
func NewTransformerRegistry() *TransformerRegistry {
	return &TransformerRegistry{
		transformers: make(map[string]*ResourceTransformer),
	}
}

// GetTransformer gets or creates a transformer for a resource type
func (tr *TransformerRegistry) GetTransformer(resourceType string) *ResourceTransformer {
	if t, ok := tr.transformers[resourceType]; ok {
		return t
	}

	t := NewResourceTransformer(resourceType)
	tr.transformers[resourceType] = t
	return t
}

// RegisterDefaultTransformers sets up common transformations
func (tr *TransformerRegistry) RegisterDefaultTransformers() {
	pgTransformer := tr.GetTransformer("aiven_pg")       // register transformers for aiven_pg
	kafkaTransformer := tr.GetTransformer("aiven_kafka") // register transformers for aiven_kafka

	// skip 0B disk space
	pgTransformer.RegisterTransformer(
		NewFieldPath("additional_disk_space"),
		func(value any) (any, bool) {
			if str, ok := value.(string); ok && str == "0B" {
				return nil, false
			}
			return value, true
		},
	)

	// skip 0B disk space
	kafkaTransformer.RegisterTransformer(
		NewFieldPath("additional_disk_space"),
		func(value any) (any, bool) {
			if str, ok := value.(string); ok && str == "0B" {
				return nil, false
			}
			return value, true
		},
	)
}
