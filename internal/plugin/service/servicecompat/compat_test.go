// Package servicecompat_test checks the first plan of the service resources on a state SDKv2 v4.65.1 stored.
// The fixtures are synthetic, in the shape SDKv2 stored, e.g., zero values for the user config options it didn't get.
package servicecompat_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/aiven/go-client-codegen/handler/staticip"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
)

const (
	testProject = "my-project"
	testCloud   = "google-europe-west1"
	testPlan    = "startup-4"
)

// fixture is a service as SDKv2 stored it, the API returns it, and the config that matches it.
type fixture struct {
	typeName string
	version  int64          // the SDKv2 schema version
	state    map[string]any // the state attributes, as in the state file
	config   map[string]any
	api      *service.ServiceGetOut
	specs    *project.ProjectServicePlanSpecsGetOut
	ips      []staticip.StaticIPOut
}

// compatCase changes the fixture and lists the changed paths, e.g., "pg_user_config[0].pg".
type compatCase struct {
	name   string
	state  map[string]any // replaces the state keys
	config map[string]any // replaces the config keys, nil removes a block
	api    func(*service.ServiceGetOut)
	ips    []staticip.StaticIPOut

	// Besides a zero value read as null, e.g., "" -> null, and a null read as an empty collection
	wantRefresh []string
	wantPlan    []string
}

// runCompat runs what the first "terraform plan" with the plugin does: upgrades the state, refreshes it,
// and plans the config.
func runCompat(t *testing.T, newFixture func() fixture, cases []compatCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := newFixture()
			maps.Copy(f.state, tc.state)
			maps.Copy(f.config, tc.config)
			if tc.api != nil {
				tc.api(f.api)
			}
			if tc.ips != nil {
				f.ips = tc.ips
			}

			upgraded, refreshed, planned := upgrade(t, f)
			refresh := changes(upgraded, refreshed)
			maps.DeleteFunc(refresh, func(_ string, c change) bool { return c.normalized() })
			assertChanges(t, "refresh", refresh, tc.wantRefresh)
			assertChanges(t, "plan", changes(refreshed, planned), tc.wantPlan)
		})
	}
}

func upgrade(t *testing.T, f fixture) (upgraded, refreshed, planned tftypes.Value) {
	t.Helper()
	ctx := t.Context()
	name := f.state["service_name"]

	client := avngen.NewMockClient(t)
	// The read requests the secrets
	client.EXPECT().ServiceGet(mock.Anything, testProject, name, mock.Anything).Return(f.api, nil).Maybe()
	client.EXPECT().ServiceGet(mock.Anything, testProject, name).Return(f.api, nil).Maybe()
	client.EXPECT().
		ProjectServicePlanSpecsGet(mock.Anything, testProject, mock.Anything, mock.Anything).
		Return(f.specs, nil).
		Maybe()
	client.EXPECT().
		ProjectServicePlanPriceGet(mock.Anything, testProject, mock.Anything, mock.Anything, mock.Anything).
		Return(&project.ProjectServicePlanPriceGetOut{ExtraDiskPricePerGbUsd: new("0.1")}, nil).
		Maybe()
	client.EXPECT().StaticIPList(mock.Anything, testProject).Return(f.ips, nil).Maybe()

	server, schemas := newServer(t, client)
	sch := schemas[f.typeName]
	require.NotNil(t, sch, f.typeName)
	typ := sch.ValueType()

	raw, err := json.Marshal(f.state)
	require.NoError(t, err)
	up, err := server.UpgradeResourceState(ctx, &tfprotov6.UpgradeResourceStateRequest{
		TypeName: f.typeName,
		Version:  f.version,
		RawState: &tfprotov6.RawState{JSON: raw},
	})
	require.NoError(t, err)
	requireNoErrors(t, up.Diagnostics)

	read, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{
		TypeName:     f.typeName,
		CurrentState: up.UpgradedState,
	})
	require.NoError(t, err)
	requireNoErrors(t, read.Diagnostics)

	refreshed = unmarshal(t, read.NewState, typ)
	config := valueForBlock(sch.Block, f.config)
	configValue := dynamicValue(t, typ, config)
	proposed := dynamicValue(t, typ, proposedNew(sch.Block, refreshed, config))
	plan, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName:         f.typeName,
		Config:           &configValue,
		PriorState:       read.NewState,
		ProposedNewState: &proposed,
		PriorPrivate:     read.Private,
		PriorIdentity:    read.NewIdentity,
	})
	require.NoError(t, err)
	requireNoErrors(t, plan.Diagnostics)
	return unmarshal(t, up.UpgradedState, typ), refreshed, unmarshal(t, plan.PlannedState, typ)
}

// newServer is the provider with the client, configured with an empty provider block.
func newServer(t *testing.T, client avngen.Client) (tfprotov6.ProviderServer, map[string]*tfprotov6.Schema) {
	t.Helper()
	server, err := providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client})()
	require.NoError(t, err)

	schemas, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	requireNoErrors(t, schemas.Diagnostics)

	config := dynamicValue(t, schemas.Provider.ValueType(), valueForBlock(schemas.Provider.Block, nil))
	rsp, err := server.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{Config: &config})
	require.NoError(t, err)
	requireNoErrors(t, rsp.Diagnostics)
	return server, schemas.ResourceSchemas
}

type change struct{ old, new tftypes.Value }

func (c change) String() string {
	return formatValue(c.old) + " -> " + formatValue(c.new)
}

// normalized tells if the change only stores a zero value as null, or reads a null as an empty collection.
func (c change) normalized() bool {
	switch {
	case c.new.IsNull():
		return isZero(c.old)
	case c.old.IsNull():
		var items []tftypes.Value
		return c.new.IsKnown() && c.new.As(&items) == nil && len(items) == 0
	}
	return false
}

// changes returns the values that differ by path. Objects and lists of the same length are compared by item.
func changes(a, b tftypes.Value) map[string]change {
	result := make(map[string]change)
	walkChanges("", a, b, result)
	return result
}

func walkChanges(path string, a, b tftypes.Value, result map[string]change) {
	if a.Equal(b) {
		return
	}

	if a.IsKnown() && !a.IsNull() && b.IsKnown() && !b.IsNull() {
		var aObj, bObj map[string]tftypes.Value
		if a.Type().Is(tftypes.Object{}) && a.As(&aObj) == nil && b.As(&bObj) == nil {
			for k := range aObj {
				walkChanges(strings.TrimPrefix(path+"."+k, "."), aObj[k], bObj[k], result)
			}
			return
		}

		var aList, bList []tftypes.Value
		if a.Type().Is(tftypes.List{}) && a.As(&aList) == nil && b.As(&bList) == nil && len(aList) == len(bList) {
			for i := range aList {
				walkChanges(fmt.Sprintf("%s[%d]", path, i), aList[i], bList[i], result)
			}
			return
		}
	}
	result[path] = change{old: a, new: b}
}

func assertChanges(t *testing.T, kind string, got map[string]change, want []string) {
	t.Helper()
	paths := slices.Sorted(maps.Keys(got))
	lines := make([]string, len(paths))
	for i, p := range paths {
		lines[i] = p + ": " + got[p].String()
	}
	assert.ElementsMatch(t, want, paths, "%s changes:\n%s", kind, strings.Join(lines, "\n"))
}

func isZero(v tftypes.Value) bool {
	if !v.IsKnown() || v.IsNull() {
		return false
	}

	var items []tftypes.Value
	var s string
	var n big.Float
	var b bool
	switch {
	case v.As(&items) == nil:
		return len(items) == 0
	case v.As(&s) == nil:
		return s == ""
	case v.As(&n) == nil:
		return n.Sign() == 0
	case v.As(&b) == nil:
		return !b
	}
	return false
}

// formatValue is a short value, e.g., "foo", 1, null, [2 items].
func formatValue(v tftypes.Value) string {
	if !v.IsKnown() {
		return "(known after apply)"
	}
	if v.IsNull() {
		return "null"
	}

	var items []tftypes.Value
	var attrs map[string]tftypes.Value
	var s string
	var n big.Float
	var b bool
	switch {
	case v.As(&items) == nil:
		return fmt.Sprintf("[%d items]", len(items))
	case v.As(&attrs) == nil:
		return fmt.Sprintf("{%d items}", len(attrs))
	case v.As(&s) == nil:
		return strconv.Quote(s)
	case v.As(&n) == nil:
		return n.String()
	case v.As(&b) == nil:
		return strconv.FormatBool(b)
	}
	return v.String()
}

// proposedNew is the plan Terraform proposes (objchange.ProposedNew) for these schemas: a computed attribute
// the config doesn't set keeps the prior value. The blocks of a list match the prior ones by index,
// of a set by the values that aren't computed. A block without a match is the config.
func proposedNew(block *tfprotov6.SchemaBlock, prior, config tftypes.Value) tftypes.Value {
	if prior.IsNull() || config.IsNull() || !config.IsKnown() {
		return config
	}

	var priorAttrs, configAttrs map[string]tftypes.Value
	_ = prior.As(&priorAttrs)
	_ = config.As(&configAttrs)

	values := make(map[string]tftypes.Value, len(configAttrs))
	for _, a := range block.Attributes {
		v := configAttrs[a.Name]
		if a.Computed && v.IsNull() {
			v = priorAttrs[a.Name]
		}
		values[a.Name] = v
	}

	for _, b := range block.BlockTypes {
		p, c := priorAttrs[b.TypeName], configAttrs[b.TypeName]
		switch b.Nesting {
		case tfprotov6.SchemaNestedBlockNestingModeList, tfprotov6.SchemaNestedBlockNestingModeSet:
			values[b.TypeName] = proposedNewBlocks(b, p, c)
		default:
			values[b.TypeName] = proposedNew(b.Block, p, c)
		}
	}
	return tftypes.NewValue(config.Type(), values)
}

func proposedNewBlocks(b *tfprotov6.SchemaNestedBlock, prior, config tftypes.Value) tftypes.Value {
	var priorItems, configItems []tftypes.Value
	if !prior.IsKnown() || prior.As(&priorItems) != nil || !config.IsKnown() || config.As(&configItems) != nil {
		return config
	}

	result := make([]tftypes.Value, len(configItems))
	used := make([]bool, len(priorItems))
	for i, c := range configItems {
		result[i] = c
		if b.Nesting == tfprotov6.SchemaNestedBlockNestingModeList {
			if i < len(priorItems) {
				result[i] = proposedNew(b.Block, priorItems[i], c)
			}
			continue
		}

		for j, p := range priorItems {
			if !used[j] && withoutComputed(b.Block, p).Equal(c) {
				result[i] = proposedNew(b.Block, p, c)
				used[j] = true
				break
			}
		}
	}
	return tftypes.NewValue(config.Type(), result)
}

// withoutComputed nulls the computed attributes. The set blocks of the services have no nested blocks.
func withoutComputed(block *tfprotov6.SchemaBlock, v tftypes.Value) tftypes.Value {
	var attrs map[string]tftypes.Value
	if v.IsNull() || v.As(&attrs) != nil {
		return v
	}

	for _, a := range block.Attributes {
		if a.Computed {
			attrs[a.Name] = tftypes.NewValue(a.ValueType(), nil)
		}
	}
	return tftypes.NewValue(v.Type(), attrs)
}

// valueForBlock is the config value of the data, e.g., {"tag": []any{map[string]any{"key": "env"}}}.
// A missing or nil block is empty, as Terraform sends it.
func valueForBlock(block *tfprotov6.SchemaBlock, data map[string]any) tftypes.Value {
	values := make(map[string]tftypes.Value, len(block.Attributes)+len(block.BlockTypes))
	for _, a := range block.Attributes {
		values[a.Name] = valueForType(a.ValueType(), data[a.Name])
	}
	for _, b := range block.BlockTypes {
		v := data[b.TypeName]
		if v == nil && b.Nesting != tfprotov6.SchemaNestedBlockNestingModeSingle {
			v = []any{}
		}
		values[b.TypeName] = valueForType(b.ValueType(), v)
	}
	return tftypes.NewValue(block.ValueType(), values)
}

func valueForType(typ tftypes.Type, value any) tftypes.Value {
	if value == nil {
		return tftypes.NewValue(typ, nil)
	}

	switch typed := typ.(type) {
	case tftypes.Object:
		input := value.(map[string]any)
		values := make(map[string]tftypes.Value, len(typed.AttributeTypes))
		for name, attributeType := range typed.AttributeTypes {
			values[name] = valueForType(attributeType, input[name])
		}
		return tftypes.NewValue(typ, values)
	case tftypes.List:
		return tftypes.NewValue(typ, collectionValues(typed.ElementType, value))
	case tftypes.Set:
		return tftypes.NewValue(typ, collectionValues(typed.ElementType, value))
	case tftypes.Map:
		input := value.(map[string]any)
		values := make(map[string]tftypes.Value, len(input))
		for key, item := range input {
			values[key] = valueForType(typed.ElementType, item)
		}
		return tftypes.NewValue(typ, values)
	default:
		return tftypes.NewValue(typ, value)
	}
}

func collectionValues(elementType tftypes.Type, value any) []tftypes.Value {
	input := value.([]any)
	values := make([]tftypes.Value, len(input))
	for i, item := range input {
		values[i] = valueForType(elementType, item)
	}
	return values
}

func dynamicValue(t *testing.T, typ tftypes.Type, value tftypes.Value) tfprotov6.DynamicValue {
	t.Helper()
	result, err := tfprotov6.NewDynamicValue(typ, value)
	require.NoError(t, err)
	return result
}

func unmarshal(t *testing.T, v *tfprotov6.DynamicValue, typ tftypes.Type) tftypes.Value {
	t.Helper()
	require.NotNil(t, v)
	result, err := v.Unmarshal(typ)
	require.NoError(t, err)
	return result
}

func requireNoErrors(t *testing.T, diags []*tfprotov6.Diagnostic) {
	t.Helper()
	for _, d := range diags {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
	}
}
