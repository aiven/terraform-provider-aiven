package adapter

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/require"
)

func TestDataSourceAdapterOptionalIDFields(t *testing.T) {
	schema := &Schema{Type: SchemaTypeObject, Properties: map[string]*Schema{
		"id":     {Type: SchemaTypeString, Computed: true},
		"name":   {Type: SchemaTypeString},
		"region": {Type: SchemaTypeString, Computed: true},
		"timeouts": {Type: SchemaTypeObject, Properties: map[string]*Schema{
			"read": {Type: SchemaTypeString},
		}},
	}}
	raw, err := toTFValue(schema, map[string]any{"name": "connection"})
	require.NoError(t, err)
	config := tfsdk.Config{Raw: raw}
	options := DataSourceOptions{
		SchemaInternal:         schema,
		IdentityFields:         []string{"name", "region"},
		IdentityFieldsOptional: []string{"region"},
	}
	checkRead := func(t *testing.T, response map[string]any, wantID string) {
		t.Helper()
		opts := options
		opts.Read = func(_ context.Context, _ avngen.Client, d ResourceData) error {
			return d.Flatten(response)
		}
		a := NewDataSource(opts)
		var rsp datasource.ReadResponse
		a.Read(t.Context(), datasource.ReadRequest{Config: config}, &rsp)
		require.False(t, rsp.Diagnostics.HasError(), "%v", rsp.Diagnostics)
		state, err := fromTFValue(schema, rsp.State.Raw, false)
		require.NoError(t, err)
		require.Equal(t, wantID, state["id"])
	}

	t.Run("Read omits an absent optional suffix", func(t *testing.T) {
		checkRead(t, map[string]any{"name": "connection"}, "connection")
	})
	t.Run("Read preserves a populated optional suffix", func(t *testing.T) {
		checkRead(t, map[string]any{"name": "connection", "region": "eu-west-1"}, "connection/eu-west-1")
	})
	t.Run("ValidateConfig uses the same ID rules", func(t *testing.T) {
		opts := options
		opts.ValidateConfig = func(_ context.Context, _ avngen.Client, d ResourceData) error {
			return d.Flatten(map[string]any{"name": "connection"})
		}
		a := NewDataSource(opts).(datasource.DataSourceWithValidateConfig)
		var rsp datasource.ValidateConfigResponse
		a.ValidateConfig(t.Context(), datasource.ValidateConfigRequest{Config: config}, &rsp)
		require.False(t, rsp.Diagnostics.HasError(), "%v", rsp.Diagnostics)
	})
}
