package diff

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDefaultIPFilterList(t *testing.T) {
	type testCase struct {
		name     string
		networks []any
		expected bool
	}

	cases := []testCase{
		// Positive cases
		{
			name:     `old efault value`,
			networks: []any{"0.0.0.0/0"},
			expected: true,
		},
		{
			name:     `new default ip_filter list`,
			networks: []any{"0.0.0.0/0", "::/0"},
			expected: true,
		},
		{
			name:     "new ip_filter list reordered",
			networks: []any{"::/0", "0.0.0.0/0"},
			expected: true,
		},
		{
			name:     `["::/0"] is not default value`,
			networks: []any{"::/0"},
			expected: false,
		},
		{
			name:     `default value with extra network`,
			networks: []any{"0.0.0.0/0", "127.0.0.1/32", "::/0"},
			expected: false,
		},
		{
			name:     `a random network`,
			networks: []any{"127.0.0.1/32"},
			expected: false,
		},
	}

	// Copies cases for ip_filter_object
	j := len(cases)
	for i := 0; i < j; i++ {
		cases = append(cases, testCase{
			name:     "ip_filter_object " + cases[i].name,
			expected: cases[i].expected,
			networks: lo.Map(cases[i].networks, func(item any, _ int) any {
				return map[string]any{"network": item}
			}),
		})
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := IsDefaultIPFilterList(c.networks)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestSuppressIPFilterSet(t *testing.T) {
	t.Parallel()

	resourceSchema := map[string]*schema.Schema{
		"foo_user_config": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"ip_filter": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name           string
		key            string
		oldValue       interface{}
		newValue       interface{}
		shouldSuppress bool
	}{
		{
			name:           "identical default IPv4 values",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"0.0.0.0/0"},
			newValue:       []interface{}{"0.0.0.0/0"},
			shouldSuppress: false,
		},
		{
			name:           "identical default IPv4 and IPv6",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"0.0.0.0/0", "::/0"},
			newValue:       []interface{}{"0.0.0.0/0", "::/0"},
			shouldSuppress: false,
		},
		{
			name:           "default values in different order",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"::/0", "0.0.0.0/0"},
			newValue:       []interface{}{"0.0.0.0/0", "::/0"},
			shouldSuppress: false,
		},
		{
			name:           "custom IP filter change",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"192.168.1.0/24"},
			newValue:       []interface{}{"10.0.0.0/24"},
			shouldSuppress: false,
		},
		{
			name:           "change from default to custom",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"0.0.0.0/0"},
			newValue:       []interface{}{"192.168.1.0/24"},
			shouldSuppress: false,
		},
		{
			name:           "default IPv4",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"0.0.0.0/0"},
			newValue:       []interface{}{"0.0.0.0/0"},
			shouldSuppress: false,
		},
		{
			name:           "default IPv4 and IPv6",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"0.0.0.0/0", "::/0"},
			newValue:       []interface{}{"0.0.0.0/0", "::/0"},
			shouldSuppress: false,
		},
		{
			name:           "default values in different order",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"::/0", "0.0.0.0/0"},
			newValue:       []interface{}{"0.0.0.0/0", "::/0"},
			shouldSuppress: false,
		},
		{
			name:           "private network change",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"192.168.1.0/24"},
			newValue:       []interface{}{"192.168.2.0/24"},
			shouldSuppress: false,
		},
		{
			name:           "private to public network",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"10.0.0.0/8"},
			newValue:       []interface{}{"203.0.113.0/24"},
			shouldSuppress: false,
		},
		{
			name:           "multiple networks",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"},
			newValue:       []interface{}{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"},
			shouldSuppress: false,
		},
		{
			name:           "IPv6 networks",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"2001:db8::/32", "2001:db8:1234::/48"},
			newValue:       []interface{}{"2001:db8::/32", "2001:db8:5678::/48"},
			shouldSuppress: false,
		},
		{
			name:           "mixed IPv4 and IPv6",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"192.168.1.0/24", "2001:db8::/32"},
			newValue:       []interface{}{"192.168.1.0/24", "2001:db8::/32"},
			shouldSuppress: false,
		},
		{
			name:           "mixed default changes to mix with custom",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"::/0", "0.0.0.0/0"},
			newValue:       []interface{}{"::/0", "0.0.0.0/0", "192.168.1.0/24"},
			shouldSuppress: false,
		},
		{
			name:           "mixed custom changes to default",
			key:            "foo_user_config.0.ip_filter",
			oldValue:       []interface{}{"::/0", "0.0.0.0/0", "192.168.1.0/24"},
			newValue:       []interface{}{"::/0", "0.0.0.0/0"},
			shouldSuppress: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a new resource data with old value
			d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
				"foo_user_config": []interface{}{
					map[string]interface{}{
						"ip_filter": tc.oldValue,
					},
				},
			})

			// Set the new value
			newConfig := map[string]interface{}{
				"foo_user_config": []interface{}{
					map[string]interface{}{
						"ip_filter": tc.newValue,
					},
				},
			}
			require.NoError(t, d.Set("foo_user_config", newConfig["foo_user_config"]))

			result := suppressIPFilterSet(tc.key, "", "", d)
			if result != tc.shouldSuppress {
				t.Errorf("Test %q: expected result to be %v but got %v", tc.name, tc.shouldSuppress, result)
			}
		})
	}
}
