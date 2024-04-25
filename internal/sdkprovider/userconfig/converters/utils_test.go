package converters

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenameAliasesToDto(t *testing.T) {
	cases := []struct {
		serviceType string
		name        string
		src         string
		expected    string
	}{
		{
			serviceType: "kafka",
			name:        "keeps original key",
			src:         `{"ip_filter": ["0.0.0.0/0"]}`,
			expected:    `{"ip_filter": ["0.0.0.0/0"]}`,
		},
		{
			serviceType: "kafka",
			name:        "chooses original out if 3",
			src:         `{"ip_filter": ["0.0.0.0/0"], "ip_filter_string": [], "ip_filter_object": []}`,
			expected:    `{"ip_filter": ["0.0.0.0/0"]}`,
		},
		{
			serviceType: "kafka",
			name:        "chooses string out if 3",
			src:         `{"ip_filter": [], "ip_filter_string": ["0.0.0.0/0"], "ip_filter_object": []}`,
			expected:    `{"ip_filter": ["0.0.0.0/0"]}`,
		},
		{
			serviceType: "kafka",
			name:        "ignores unknown key",
			src:         `{"whatever": ["0.0.0.0/0"]}`,
			expected:    `{"whatever": ["0.0.0.0/0"]}`,
		},
		{
			serviceType: `kafka`,
			name:        `renames "_string" prefix`,
			src:         `{"ip_filter_string": ["0.0.0.0/0"]}`,
			expected:    `{"ip_filter": ["0.0.0.0/0"]}`,
		},
		{
			serviceType: `kafka`,
			name:        `renames "_string" prefix empty`,
			src:         `{"ip_filter_string": []}`,
			expected:    `{"ip_filter": []}`,
		},
		{
			serviceType: "kafka",
			name:        "renames _object prefix",
			src:         `{"ip_filter_object": [{"name": "foo"}]}`,
			expected:    `{"ip_filter": [{"name": "foo"}]}`,
		},
		{
			serviceType: "m3db",
			name:        "ignores namespaces_string on the root level",
			src:         `{"namespaces_string": {"name": "foo"}}`,
			expected:    `{"namespaces_string": {"name": "foo"}}`,
		},
		{
			serviceType: "m3db",
			name:        "renames namespaces_string where expected",
			src:         `{"rules": {"mapping": [{"namespaces_string": ["string"]}]}}`,
			expected:    `{"rules": {"mapping": [{"namespaces": ["string"]}]}}`,
		},
		{
			serviceType: "m3db",
			name:        "renames namespaces_object where expected",
			src:         `{"rules": {"mapping": [{"namespaces_object": [{"name": "foo"}]}]}}`,
			expected:    `{"rules": {"mapping": [{"namespaces": [{"name": "foo"}]}]}}`,
		},
		{
			serviceType: "thanos",
			name:        "with a prent field to rename",
			src:         `{"query_frontend": {"query_range_align_range_with_step": 0}}`,
			expected:    `{"query-frontend": {"query-range.align-range-with-step": 0}}`,
		},
		{
			serviceType: "pg",
			name:        "pg __dot__ legacy field",
			src:         `{"pg": {"pg_stat_statements__dot__track": 0}}`,
			expected:    `{"pg": {"pg_stat_statements.track": 0}}`,
		},
	}

	reSpaces := regexp.MustCompile(`\s+`)
	for _, opt := range cases {
		t.Run(opt.name, func(t *testing.T) {
			var m map[string]any
			err := json.Unmarshal([]byte(opt.src), &m)
			require.NoError(t, err)

			renameAliasesToDto(ServiceUserConfig, opt.serviceType, m)
			b, err := json.Marshal(&m)
			require.NoError(t, err)
			assert.Equal(t, reSpaces.ReplaceAllString(opt.expected, ""), string(b))
		})
	}
}

func TestRenameAliasesToTfo(t *testing.T) {
	cases := []struct {
		serviceType string
		name        string
		dto         string
		expected    string
		tfo         resourceData
	}{
		{
			serviceType: "kafka",
			name:        "keeps original key",
			expected:    `{"ip_filter": ["0.0.0.0/0"]}`,
			dto:         `{"ip_filter": ["0.0.0.0/0"]}`,
			tfo: newResourceDataMock(
				newResourceDataKV("kafka_user_config.0.ip_filter", []string{"0.0.0.0/0"}),
			),
		},
		{
			serviceType: "kafka",
			name:        "uses ip_filter_string",
			expected:    `{"ip_filter_string": ["0.0.0.0/0"]}`,
			dto:         `{"ip_filter": ["0.0.0.0/0"]}`,
			tfo: newResourceDataMock(
				newResourceDataKV("kafka_user_config.0.ip_filter_string", []string{"0.0.0.0/0"}),
			),
		},
		{
			serviceType: "kafka",
			name:        "uses ip_filter_object",
			expected:    `{"ip_filter_object": [{"network": "0.0.0.0/0"}]}`,
			dto:         `{"ip_filter": [{"network": "0.0.0.0/0"}]}`,
			tfo: newResourceDataMock(
				newResourceDataKV("kafka_user_config.0.ip_filter_object", []map[string]string{{"network": "0.0.0.0/0"}}),
			),
		},
		{
			serviceType: "m3db",
			name:        "renames namespaces_string",
			expected:    `{"rules": {"mapping": [{"namespaces_string": ["string"]}]}}`,
			dto:         `{"rules": {"mapping": [{"namespaces": ["string"]}]}}`,
			tfo: newResourceDataMock(
				newResourceDataKV("m3db_user_config.0.rules.0.mapping.0.namespaces_string", "string"),
			),
		},
		{
			serviceType: "m3db",
			name:        "renames namespaces_object",
			expected:    `{"rules": {"mapping": [{"namespaces_object": [{"name": "foo"}]}]}}`,
			dto:         `{"rules": {"mapping": [{"namespaces": [{"name": "foo"}]}]}}`,
			tfo: newResourceDataMock(
				newResourceDataKV("m3db_user_config.0.rules.0.mapping.0.namespaces_object", []map[string]string{{"name": "foo"}}),
			),
		},
		{
			serviceType: "thanos",
			name:        "renames dots and dashes",
			expected:    `{"query_frontend": {"query_range_align_range_with_step": 0}}`,
			dto:         `{"query-frontend": {"query-range.align-range-with-step": 0}}`,
			tfo:         newResourceDataMock(),
		},
		{
			serviceType: "pg",
			name:        "pg __dot__ legacy field",
			expected:    `{"pg": {"pg_stat_statements__dot__track": 0}}`,
			dto:         `{"pg": {"pg_stat_statements.track": 0}}`,
			tfo:         newResourceDataMock(),
		},
	}

	reSpaces := regexp.MustCompile(`\s+`)
	for _, opt := range cases {
		t.Run(opt.name, func(t *testing.T) {
			var m map[string]any
			err := json.Unmarshal([]byte(opt.dto), &m)
			require.NoError(t, err)

			renameAliasesToTfo(ServiceUserConfig, opt.serviceType, m, opt.tfo)
			b, err := json.Marshal(&m)
			require.NoError(t, err)
			assert.Equal(t, reSpaces.ReplaceAllString(opt.expected, ""), string(b))
		})
	}
}

type resourceDataMock struct {
	m map[string]any
}

func (r *resourceDataMock) GetOk(k string) (any, bool) {
	v, ok := r.m[k]
	return v, ok
}

func newResourceDataKV(k string, v any) func(map[string]any) {
	return func(m map[string]any) { m[k] = v }
}

func newResourceDataMock(kv ...func(map[string]any)) resourceData {
	m := make(map[string]any)
	for _, v := range kv {
		v(m)
	}
	return &resourceDataMock{m: m}
}

func TestIsZero(t *testing.T) {
	cases := map[string]struct {
		value  any
		expect bool
	}{
		"empty string": {
			value:  "",
			expect: true,
		},
		"non-empty string": {
			value:  "foo",
			expect: false,
		},
		"empty array": {
			value:  []string{},
			expect: true,
		},
		"non-empty array": {
			value:  []string{"foo"},
			expect: false,
		},
		"empty map": {
			value:  map[string]string{},
			expect: true,
		},
		"non-empty map": {
			value:  map[string]string{"foo": "bar"},
			expect: false,
		},
	}

	for name, opt := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, opt.expect, isZero(opt.value))
		})
	}
}
