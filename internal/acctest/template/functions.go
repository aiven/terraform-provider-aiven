package template

import (
	"fmt"
	"html/template"
	"sort"
	"strings"
)

// templateFunctions provides a centralized store of template functions
// that can be shared between different components of the template system.
type templateFunctions struct {
	funcMap template.FuncMap
}

// newTemplateFunctions creates a new function registry with default template functions
func newTemplateFunctions() *templateFunctions {
	tf := &templateFunctions{
		funcMap: make(template.FuncMap),
	}

	tf.registerDefaults()

	return tf
}

// registerDefaults registers the default set of template functions
func (tf *templateFunctions) registerDefaults() {
	tf.register("required", func(v any) any {
		if v == nil {
			panic("required field is missing")
		}
		return v
	})

	tf.register("renderValue", func(v interface{}) template.HTML {
		var result string
		switch val := v.(type) {
		case Value:
			if val.IsLiteral {
				result = fmt.Sprintf("%q", val.Value)
			} else {
				result = val.Value
			}
		case string:
			result = fmt.Sprintf("%q", val)
		case int, int64, float64:
			result = fmt.Sprintf("%v", val)
		case bool:
			result = fmt.Sprintf("%v", val)
		case []interface{}:
			// Check if we have a list of maps
			allMaps := true
			for _, item := range val {
				if _, ok := item.(map[string]interface{}); !ok {
					allMaps = false
					break
				}
			}

			if allMaps && len(val) > 0 {
				// Format list of maps with proper indentation
				var elements []string
				for _, item := range val {
					if m, ok := item.(map[string]interface{}); ok {
						var pairs []string
						keys := make([]string, 0, len(m))
						for k := range m {
							keys = append(keys, k)
						}
						sort.Strings(keys)

						for _, k := range keys {
							pairs = append(pairs, fmt.Sprintf("    %s = %s", k, renderValueAsString(m[k])))
						}
						elements = append(elements, fmt.Sprintf("  {\n%s\n  }", strings.Join(pairs, "\n")))
					}
				}
				result = fmt.Sprintf("[\n%s\n]", strings.Join(elements, ",\n"))
			} else {
				// Regular array formatting
				var elements []string
				for _, elem := range val {
					elements = append(elements, renderValueAsString(elem))
				}
				result = fmt.Sprintf("[\n    %s,\n  ]", strings.Join(elements, ",\n    "))
			}
		case []string:
			var elements []string
			for _, elem := range val {
				elements = append(elements, fmt.Sprintf("%q", elem))
			}
			result = fmt.Sprintf("[\n    %s,\n  ]", strings.Join(elements, ",\n    "))
		case map[string]interface{}:
			var pairs []string
			keys := make([]string, 0, len(val))
			for k := range val {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				pairs = append(pairs, fmt.Sprintf("%q = %s", k, renderValueAsString(val[k])))
			}
			result = fmt.Sprintf("{\n    %s\n  }", strings.Join(pairs, "\n    "))
		default:
			result = fmt.Sprintf("%q", val)
		}
		return template.HTML(result)
	})
}

// Helper function to render a value as a string
func renderValueAsString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case int, int64, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// register adds a new function to the template function map
func (tf *templateFunctions) register(name string, fn any) {
	tf.funcMap[name] = fn
}

// getFuncMap returns a copy of the template function map
func (tf *templateFunctions) getFuncMap() template.FuncMap {
	funcs := make(template.FuncMap, len(tf.funcMap))
	for k, v := range tf.funcMap {
		funcs[k] = v
	}
	return funcs
}
