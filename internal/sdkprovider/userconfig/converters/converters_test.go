package converters

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestFlattenSafe(t *testing.T) {
	sch := map[string]*schema.Schema{
		"ip_filter": {
			Elem: &schema.Schema{Type: schema.TypeString},
			Type: schema.TypeList,
		},
		"ip_filter_object": {
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"network": {Type: schema.TypeString},
			}},
			Type: schema.TypeList,
		},
	}

	cases := []struct {
		description string
		src         map[string]any
		expected    map[string]any
	}{
		{
			description: "converts objects list into a list of strings",
			src: map[string]any{
				"ip_filter_object": []string{"0.0.0.0/0"},
			},
			expected: map[string]any{
				"ip_filter_object": []map[string]string{{"network": "0.0.0.0/0"}},
			},
		},
		{
			description: "converts objects list into a list of strings",
			src: map[string]any{
				"ip_filter": []map[string]string{{"network": "0.0.0.0/0"}},
			},
			expected: map[string]any{
				"ip_filter": []string{"0.0.0.0/0"},
			},
		},
	}

	for _, opt := range cases {
		t.Run(opt.description, func(t *testing.T) {
			// Converts ip filters first
			err := convertIPFilter(opt.src)
			assert.NoError(t, err)

			// Then flattens
			newDto, err := flattenSafe(sch, opt.src)
			assert.NoError(t, err)

			result, err := json.Marshal(&newDto)
			assert.NoError(t, err)

			expected, err := json.Marshal(&opt.expected)
			assert.NoError(t, err)

			assert.Equal(t, expected, result)
		})
	}
}
