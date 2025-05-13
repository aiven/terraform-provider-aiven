package kafkatopic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApiConfigTypesFieldsExistInSchema
// Proves that all fields in the apiConfigTypes are also present in the schema.
func TestApiConfigTypesFieldsExistInSchema(t *testing.T) {
	fieldTypes := apiConfigTypes()
	for k := range aivenKafkaTopicConfigSchema() {
		_, ok := fieldTypes[k]
		assert.True(t, ok, "missing field %q in apiConfigTypes", k)
	}
}

// TestSchemaFieldsExistInApiConfigTypes
// Proves that all fields in the schema are also present in the apiConfigTypes.
func TestSchemaFieldsExistInApiConfigTypes(t *testing.T) {
	fieldSchemas := aivenKafkaTopicConfigSchema()
	for k := range apiConfigTypes() {
		_, ok := fieldSchemas[k]
		assert.True(t, ok, "missing field %q in aivenKafkaTopicConfigSchema", k)
	}
}

func TestTypedConfigValue(t *testing.T) {
	cases := []struct {
		field, description string
		in, expect         any
	}{
		{
			field:       "cleanup_policy",
			description: "should return a string",
			in:          "foo",
			expect:      "foo",
		},
		{
			field:       "local_retention_bytes",
			description: "should return a integer, it is a string in TF",
			in:          "1234",
			expect:      int64(1234),
		},
		{
			field:       "min_cleanable_dirty_ratio",
			description: "should return a number",
			in:          "0.51",
			expect:      0.51,
		},
		{
			field:       "remote_storage_enable",
			description: "true boolean",
			in:          true,
			expect:      true,
		},
		{
			field:       "message_downconversion_enable",
			description: "false boolean",
			in:          false,
			expect:      false,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			out, err := typedConfigValue(c.field, c.in)
			require.NoError(t, err)
			assert.Equal(t, c.expect, out)
		})
	}
}
