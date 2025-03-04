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
			var elements []string
			for _, elem := range val {
				elements = append(elements, fmt.Sprintf("%v", elem))
			}
			result = fmt.Sprintf("[%s]", strings.Join(elements, ", "))
		case []string:
			var elements []string
			for _, elem := range val {
				elements = append(elements, fmt.Sprintf("%q", elem))
			}
			result = fmt.Sprintf("[%s]", strings.Join(elements, ", "))
		case map[string]interface{}:
			var pairs []string
			keys := make([]string, 0, len(val))
			for k := range val {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				pairs = append(pairs, fmt.Sprintf("%q = %v", k, val[k]))
			}
			result = fmt.Sprintf("{ %s }", strings.Join(pairs, ", "))
		default:
			result = fmt.Sprintf("%q", val)
		}
		return template.HTML(result) // nosemgrep
	})
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
