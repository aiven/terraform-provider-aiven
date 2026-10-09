package mirrormakerreplicationflow

import (
	"context"
	"fmt"
	"strings"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// configPropertiesExclude is a set of strings in the schema, while the API
// takes and returns it as a comma-separated string.
const configPropertiesExclude = "config_properties_exclude"

func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return expandConfigPropertiesExclude
}

func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return flattenConfigPropertiesExclude
}

// expandConfigPropertiesExclude joins the set of strings into the comma-separated
// string the API expects. The value is read from the resource data rather than the
// dto, so an empty one can be sent, which is how the backend is told to clear it.
//
// An empty value is sent only when the field is being cleared. Omitting the field
// leaves the API's own list of exclusions in place, while an empty string replaces
// it with nothing, so a field that was never configured must stay out of the
// request. The SDK resource drew the same line: it sent the field only when it had
// changed.
func expandConfigPropertiesExclude(d adapter.ResourceData, dto map[string]any) error {
	var parts []string
	switch v := d.Get(configPropertiesExclude).(type) {
	case nil:
		// Not configured: nothing to join.
	case []any:
		parts = make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return fmt.Errorf("unexpected type %T for %s element", item, configPropertiesExclude)
			}
			parts = append(parts, s)
		}
	default:
		return fmt.Errorf("unexpected type %T for %s", v, configPropertiesExclude)
	}

	if len(parts) == 0 && !d.HasChange(configPropertiesExclude) {
		delete(dto, configPropertiesExclude)
		return nil
	}

	dto[configPropertiesExclude] = strings.Join(parts, ",")
	return nil
}

// flattenConfigPropertiesExclude splits the comma-separated string returned by the
// API back into a set of strings. An empty string becomes an empty set instead of
// a set holding one empty string.
func flattenConfigPropertiesExclude(_ adapter.ResourceData, dto map[string]any) error {
	v, ok := dto[configPropertiesExclude]
	if !ok || v == nil {
		return nil
	}

	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("unexpected type %T for %s", v, configPropertiesExclude)
	}

	out := make([]any, 0)
	if s != "" {
		for p := range strings.SplitSeq(s, ",") {
			out = append(out, p)
		}
	}

	dto[configPropertiesExclude] = out
	return nil
}
