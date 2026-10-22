package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsesStateForUnknown(t *testing.T) {
	t.Parallel()

	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			// Server assigned, never sent back: inferred.
			"application_id": {Computed: true},
			"created_at":     {Computed: true},
			"created_by":     {Computed: true},
			"create_time":    {Computed: true},
			"uuid":           {Computed: true},
			// Configuration named after creation, not creation metadata.
			"create_table":     {Computed: true},
			"created_manually": {Computed: true},
			// Configurable, so the new value is known from the plan.
			"version_id":  {Required: true},
			"instance_id": {Optional: true},
			// Optional with a server side default: the plan can't assume the
			// state value, the user may have just set it.
			"parent_id": {Optional: true, Computed: true},
			// Auto-promoted enum string (optional on Create, echoed on Read):
			// qualifies via AutoComputed even though Optional is set.
			// AutoComputed is cached by recalcDeep from the surrounding
			// Operations; preset here because this test exercises
			// UsesStateForUnknown in isolation.
			"content_type": {Optional: true, Computed: true, AutoComputed: true},
			// Changes over the resource's lifetime, so the state value is stale.
			"updated_at":     {Computed: true},
			"updated_by":     {Computed: true},
			"update_time":    {Computed: true},
			"state":          {Computed: true},
			"status":         {Computed: true},
			"last_used_time": {Computed: true},
			// Rotated by password resets and CA renewals.
			"access_cert": {Computed: true},
			"access_key":  {Computed: true},
			// The definition wins over the inference, in both directions.
			"opted_out_id": {Computed: true, OverrideUseStateForUnknown: new(false)},
			"host":         {Computed: true, OverrideUseStateForUnknown: new(true)},
			// A nested "*_id" belongs to a child entity that is replaced over
			// the lifetime of its parent.
			"current_deployment": {
				Computed:   true,
				Type:       SchemaTypeObject,
				Properties: map[string]*Item{"deployment_id": {Computed: true}},
			},
		},
	}

	for name, item := range root.Properties {
		item.Name = name
		item.Parent = root
		for nested, child := range item.Properties {
			child.Name = nested
			child.Parent = item
		}
	}

	expect := map[string]bool{
		"application_id":   true,
		"created_at":       true,
		"created_by":       true,
		"create_time":      true,
		"uuid":             true,
		"host":             true,
		"create_table":     false,
		"created_manually": false,
		"version_id":       false,
		"instance_id":      false,
		"parent_id":        false,
		"updated_at":       false,
		"updated_by":       false,
		"update_time":      false,
		"state":            false,
		"status":           false,
		"last_used_time":   false,
		"access_cert":      false,
		"access_key":       false,
		"opted_out_id":     false,
		"content_type":     true,
	}

	for name, want := range expect {
		assert.Equal(t, want, root.Properties[name].UsesStateForUnknown(), name)
	}

	nested := root.Properties["current_deployment"].Properties["deployment_id"]
	assert.False(t, nested.UsesStateForUnknown(), "nested deployment_id")
}

func TestAutoComputed(t *testing.T) {
	t.Parallel()

	// A canonical definition with one create and two reads. The second read is
	// an auxiliary GET (e.g. ProjectKmsGetCA) that only returns an unrelated
	// field; a round-trip body is expected to appear in the primary read, not
	// in every read.
	def := &Definition{
		Operations: Operations{
			{ID: "FooCreate", Type: OperationCreate},
			{ID: "FooRead", Type: OperationRead},
			{ID: "FooAuxRead", Type: OperationRead},
			{ID: "FooUpdate", Type: OperationUpdate},
		},
	}
	ops := def.Operations
	createReqBit := ops.AppearsInID("FooCreate", OperationCreate, RequestBody)
	createResBit := ops.AppearsInID("FooCreate", OperationCreate, ResponseBody)
	readResBit := ops.AppearsInID("FooRead", OperationRead, ResponseBody)
	readPathBit := ops.AppearsInID("FooRead", OperationRead, PathParameter)
	auxReadResBit := ops.AppearsInID("FooAuxRead", OperationRead, ResponseBody)
	updateReqBit := ops.AppearsInID("FooUpdate", OperationUpdate, RequestBody)

	// Round-trip: present in Create request + Update request + primary Read
	// response. The Update bit is required so recalcDeep does not infer
	// ForceNew (which would veto AutoComputed). Not in the aux Read response —
	// matches real-world fields like `aiven_organization_project.base_port`
	// which the OR-merged mask would have missed.
	roundTrip := createReqBit | updateReqBit | readResBit

	root := &Item{Name: "root", Type: SchemaTypeObject}
	mk := func(name string, item *Item) *Item {
		item.Name = name
		item.Parent = root
		return item
	}

	// Without Update, configurable fields require replacement. Automatic Computed
	// must not change config removal into prior state preservation.
	// Models aiven_organization_application_user_token.max_age_seconds.
	noUpdateDef := &Definition{
		Operations: Operations{
			{ID: "TokenCreate", Type: OperationCreate},
			{ID: "TokenRead", Type: OperationRead},
			{ID: "TokenDelete", Type: OperationDelete},
		},
	}
	noUpdateCreateReq := noUpdateDef.Operations.AppearsInID("TokenCreate", OperationCreate, RequestBody)
	noUpdateReadRes := noUpdateDef.Operations.AppearsInID("TokenRead", OperationRead, ResponseBody)

	cases := []struct {
		name string
		def  *Definition
		item *Item
		want bool
	}{
		// Scalars eligible for auto-promotion.
		{"integer round-trip", def, mk("replication_factor", &Item{Type: SchemaTypeInteger, AppearsIn: roundTrip}), true},
		{"number round-trip", def, mk("offset_lag_max", &Item{Type: SchemaTypeNumber, AppearsIn: roundTrip}), true},
		{"boolean round-trip", def, mk("restart_enabled", &Item{Type: SchemaTypeBoolean, AppearsIn: roundTrip}), true},
		{"enum string round-trip", def, mk("content_type", &Item{Type: SchemaTypeString, AppearsIn: roundTrip, Enum: []any{"application/java-archive", "application/zip"}}), true},
		// Free-form strings: the resource adapter clears them by sending "".
		{"free-form string", def, mk("host", &Item{Type: SchemaTypeString, AppearsIn: roundTrip}), false},
		// Collections: the adapter clears them by sending [].
		{"array", def, mk("topics", &Item{Type: SchemaTypeArrayOrdered, AppearsIn: roundTrip, Items: &Item{Type: SchemaTypeString}}), false},
		// Required+Computed is rejected by the framework.
		{"required scalar", def, mk("service_type", &Item{Type: SchemaTypeString, AppearsIn: roundTrip, Required: true, Enum: []any{"kafka_connect"}}), false},
		// Response-only is handled by the existing "response-only ⇒ Computed" branch.
		{"response only", def, mk("created_at", &Item{Type: SchemaTypeString, AppearsIn: readResBit}), false},
		// Request-only has nothing to flatten.
		{"request only", def, mk("plugin_description", &Item{Type: SchemaTypeString, AppearsIn: createReqBit}), false},
		// Mixed-origin: in Create response + Update request + Read path parameter.
		// The old implementation OR'd generic bits (CreateHandler | RequestBody
		// | ReadHandler | ResponseBody) and accepted this. The per-operation
		// check rejects it. (CodeRabbit review on PR #3121.)
		{"mixed origin bits", def, mk("etag_mixed_origin", &Item{Type: SchemaTypeString, AppearsIn: createResBit | updateReqBit | readPathBit, Enum: []any{"a", "b"}}), false},
		// Multiple reads: the primary read echoes the field, the aux read does
		// not. OR-merged handlerMask would require both; AppearsInHandler allows one.
		// Includes updateReqBit so recalcDeep does not infer ForceNew.
		{"primary read only, aux read ignored", def, mk("base_port", &Item{Type: SchemaTypeInteger, AppearsIn: createReqBit | updateReqBit | readResBit}), true},
		{"only aux read echoes the field", def, mk("ca_cert", &Item{Type: SchemaTypeString, AppearsIn: createReqBit | updateReqBit | auxReadResBit, Enum: []any{"a", "b"}}), true},
		// No Update op → recalcDeep infers ForceNew=true, so AutoComputed must
		// stay false to preserve replacement when the field leaves the config.
		// Models aiven_organization_application_user_token.max_age_seconds.
		{"create and read only, no update op", noUpdateDef, mk("max_age_seconds", &Item{Type: SchemaTypeInteger, AppearsIn: noUpdateCreateReq | noUpdateReadRes}), false},
		// The inverse: no Update op in the spec, but OverrideForceNew=false
		// restores USFU for fields updatable out-of-band (e.g.
		// aiven_mysql_user.authentication via ServiceUserCredentialsModify).
		{"create and read only, forceNew override cleared", noUpdateDef, mk("authentication", &Item{Type: SchemaTypeString, AppearsIn: noUpdateCreateReq | noUpdateReadRes, Enum: []any{"caching_sha2_password", "mysql_native_password"}, OverrideForceNew: lo.ToPtr(false)}), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.NoError(t, recalcDeep(tc.def, tc.item))
			assert.Equal(t, tc.want, tc.item.AutoComputed)
		})
	}

	// Assert the final schema contract as well as the inference flag above.
	t.Run("mutable scalar accepts server values", func(t *testing.T) {
		item := mk("replication_factor", &Item{Type: SchemaTypeInteger, AppearsIn: roundTrip})
		require.NoError(t, recalcDeep(def, item))
		assert.True(t, item.Optional)
		assert.True(t, item.Computed)
		assert.False(t, item.ForceNew)
		assert.True(t, item.UsesStateForUnknown())
	})

	t.Run("immutable scalar preserves replacement on removal", func(t *testing.T) {
		item := mk("max_age_seconds", &Item{Type: SchemaTypeInteger, AppearsIn: noUpdateCreateReq | noUpdateReadRes})
		require.NoError(t, recalcDeep(noUpdateDef, item))
		assert.True(t, item.Optional)
		assert.False(t, item.Computed)
		assert.True(t, item.ForceNew)
		assert.False(t, item.UsesStateForUnknown())
	})

	t.Run("forceNew false permits automatic computed", func(t *testing.T) {
		item := mk("authentication", &Item{
			Type: SchemaTypeString, Enum: []any{"caching_sha2_password", "mysql_native_password"},
			AppearsIn: noUpdateCreateReq | noUpdateReadRes, OverrideForceNew: lo.ToPtr(false),
		})
		require.NoError(t, recalcDeep(noUpdateDef, item))
		assert.True(t, item.Optional)
		assert.True(t, item.Computed)
		assert.False(t, item.ForceNew)
		assert.True(t, item.UsesStateForUnknown())
	})

	t.Run("forceNew true prevents automatic computed despite update support", func(t *testing.T) {
		item := mk("replication_factor", &Item{
			Type: SchemaTypeInteger, AppearsIn: roundTrip, OverrideForceNew: lo.ToPtr(true),
		})
		require.NoError(t, recalcDeep(def, item))
		assert.False(t, item.AutoComputed)
		assert.True(t, item.Optional)
		assert.False(t, item.Computed)
		assert.True(t, item.ForceNew)
		assert.False(t, item.UsesStateForUnknown())
	})

	t.Run("immutable scalar keeps default without automatic state preservation", func(t *testing.T) {
		item := mk("extend_when_used", &Item{
			Type: SchemaTypeBoolean, AppearsIn: noUpdateCreateReq | noUpdateReadRes, Default: false,
		})
		require.NoError(t, recalcDeep(noUpdateDef, item))
		assert.False(t, item.AutoComputed)
		assert.True(t, item.Optional)
		assert.True(t, item.Computed)
		assert.True(t, item.ForceNew)
		assert.Equal(t, false, item.Default)
		assert.False(t, item.UsesStateForUnknown())
	})

	t.Run("immutable scalar explicitly retains server values", func(t *testing.T) {
		item := mk("parallelism", &Item{
			Type: SchemaTypeInteger, AppearsIn: noUpdateCreateReq | noUpdateReadRes,
			OverrideComputed: lo.ToPtr(true), OverrideUseStateForUnknown: lo.ToPtr(true),
		})
		require.NoError(t, recalcDeep(noUpdateDef, item))
		assert.False(t, item.AutoComputed)
		assert.True(t, item.Optional)
		assert.True(t, item.Computed)
		assert.True(t, item.ForceNew)
		assert.True(t, item.UsesStateForUnknown())
	})

	t.Run("computed false overrides automatic promotion", func(t *testing.T) {
		item := mk("replication_factor", &Item{
			Type: SchemaTypeInteger, AppearsIn: roundTrip, OverrideComputed: lo.ToPtr(false),
		})
		require.NoError(t, recalcDeep(def, item))
		assert.True(t, item.Optional)
		assert.False(t, item.Computed)
		assert.False(t, item.ForceNew)
		assert.False(t, item.UsesStateForUnknown())
	})

	t.Run("nested never auto-promotes", func(t *testing.T) {
		t.Parallel()
		parent := &Item{
			Name: "config", Type: SchemaTypeObject, Parent: root, AppearsIn: roundTrip,
			Properties: map[string]*Item{
				"retention_ms": {Name: "retention_ms", Type: SchemaTypeInteger, AppearsIn: roundTrip},
			},
		}
		parent.Properties["retention_ms"].Parent = parent
		require.NoError(t, recalcDeep(def, parent))
		assert.False(t, parent.Properties["retention_ms"].AutoComputed)
	})
}

func TestItemPath(t *testing.T) {
	items := &Item{
		Name: "array",
		Type: SchemaTypeString,
	}
	array := &Item{
		Name:  "array",
		Type:  SchemaTypeArray,
		Items: items,
	}
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"array": array,
		},
	}

	array.Parent = root
	items.Parent = array
	assert.Equal(t, "array", array.Path())
	assert.Equal(t, "array", items.Path())
}

// TestApplyDescModifiers pins the description post-processing that wraps a lone
// symbol in backticks: whitespace-delimited (e.g. the "*" wildcard) or
// single-quoted (e.g. '.', '-'), while leaving Markdown emphasis, quoted
// words/values and apostrophes inside words untouched.
func TestApplyDescModifiers(t *testing.T) {
	cases := map[string]string{
		// whitespace-delimited symbols
		"the host or * for all hosts": "the host or `*` for all hosts",
		"foo * bar * baz":             "foo `*` bar `*` baz",
		// single-quoted symbols (quotes dropped), "_" included
		"symbols: '.', '_', and '-'.": "symbols: `.`, `_`, and `-`.",
		"an '_id' suffix stays":       "an '_id' suffix stays",
		// double-quoted symbols, single or a run, literal or &quot; entity
		`use "/" as a separator`:       "use `/` as a separator",
		`defaults to [".*"]`:           "defaults to [`.*`]",
		`defaults to [&quot;.*&quot;]`: "defaults to [`.*`]",
		// a lazy run stops at the closing quote instead of spanning two of them
		`either ".*" or "-" works`: "either `.*` or `-` works",
		// a quoted token never contains whitespace, so the gap between two
		// quoted values must not be mistaken for one (JSON snippets in
		// descriptions rely on this)
		`{"key": "value"}`:                     `{"key": "value"}`,
		`{&quot;key&quot;: &quot;value&quot;}`: `{&quot;key&quot;: &quot;value&quot;}`,
		`values include 'delete', 'compact'`:   `values include 'delete', 'compact'`,
		// left untouched
		`matched by "topics.exclude"`: `matched by "topics.exclude"`,
		"the value 'all' is allowed":  "the value 'all' is allowed",
		"a digit like '5' stays":      "a digit like '5' stays",
		"don't touch apostrophes":     "don't touch apostrophes",
		"**bold** and *italic*":       "**bold** and *italic*",
		"(*)":                         "(*)",                 // no surrounding whitespace
		"ends with *":                 "ends with *",         // no trailing whitespace
		"* leads":                     "* leads",             // no leading whitespace
		"**bold** text":               "**bold** text",       // emphasis left intact
		"a*b":                         "a*b",                 // not whitespace-delimited
		"already `*` wrapped":         "already `*` wrapped", // no double-wrapping
	}

	for in, want := range cases {
		assert.Equal(t, want, applyDescModifiers(in), "input: %q", in)
	}
}

// TestNormalizeDescription pins the newline handling: root descriptions keep
// author-intended blank lines as paragraph breaks while hard-wrapped lines are
// unwrapped, and non-root descriptions collapse to a single line.
func TestNormalizeDescription(t *testing.T) {
	const src = "First paragraph line one\nline two.\n\nSecond paragraph with a * wildcard."

	assert.Equal(t,
		"First paragraph line one line two.\n\nSecond paragraph with a `*` wildcard.",
		normalizeDescription(src, true),
		"root: blank line preserved, soft wrap collapsed, wildcard backticked",
	)

	assert.Equal(t,
		"First paragraph line one line two. Second paragraph with a `*` wildcard.",
		normalizeDescription(src, false),
		"non-root: everything collapses to one line",
	)
}

func TestLimitedAvailabilitySchemaDescription(t *testing.T) {
	code, err := genSchema(
		&Definition{
			LimitedAvailability: lo.ToPtr(true),
			Resource:            &SchemaMeta{},
		},
		resourceType,
		&Item{
			Name:        "foo",
			Description: "Does stuff",
			Type:        SchemaTypeObject,
			Properties:  map[string]*Item{},
		},
	)
	require.NoError(t, err)

	file := jen.NewFile("test")
	file.Add(code)

	var b bytes.Buffer
	require.NoError(t, file.Render(&b))
	got := b.String()

	require.Contains(t, got, "limited availability")
	require.Contains(t, got, "contact the [sales team](http://aiven.io/contact)")
}

func TestLimitedAvailabilityViewOptions(t *testing.T) {
	code, err := genNewResource(resourceType, &Definition{LimitedAvailability: lo.ToPtr(true), Resource: &SchemaMeta{}}, &Item{}, false)
	require.NoError(t, err)

	file := jen.NewFile("test")
	file.Add(code)

	var b bytes.Buffer
	require.NoError(t, file.Render(&b))
	got := b.String()

	require.NotContains(t, got, "Limited:")
}

// TestItemRemoveElements verifies that elements are correctly removed from nested structures.
func TestItemRemoveElements(t *testing.T) {
	tests := []struct {
		name   string
		schema *OASchema
		remove []string
		want   map[string]*Item
	}{
		{
			name: "array of objects",
			schema: &OASchema{
				Type: SchemaTypeObject,
				Properties: map[string]*OASchema{
					"plans": {
						Type: SchemaTypeArray,
						Items: &OASchema{
							Type: SchemaTypeObject,
							Properties: map[string]*OASchema{
								"count": {Type: SchemaTypeInteger},
								"name":  {Type: SchemaTypeString},
							},
						},
					},
				},
			},
			remove: []string{"plans/count"},
			want: map[string]*Item{
				"plans": {
					Properties: map[string]*Item{},
					Items: &Item{
						Properties: map[string]*Item{
							"name": {Properties: map[string]*Item{}}, // "count" removed
						},
					},
				},
			},
		},
		{
			name: "nested arrays",
			schema: &OASchema{
				Type: SchemaTypeObject,
				Properties: map[string]*OASchema{
					"backups": {
						Type: SchemaTypeArray,
						Items: &OASchema{
							Type: SchemaTypeObject,
							Properties: map[string]*OASchema{
								"files": {
									Type: SchemaTypeArray,
									Items: &OASchema{
										Type: SchemaTypeObject,
										Properties: map[string]*OASchema{
											"size": {Type: SchemaTypeInteger},
											"name": {Type: SchemaTypeString},
										},
									},
								},
							},
						},
					},
				},
			},
			remove: []string{"backups/files/size"},
			want: map[string]*Item{
				"backups": {
					Properties: map[string]*Item{},
					Items: &Item{
						Properties: map[string]*Item{
							"files": {
								Properties: map[string]*Item{},
								Items: &Item{
									Properties: map[string]*Item{
										"name": {Properties: map[string]*Item{}}, // "size" removed
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "map of objects",
			schema: &OASchema{
				Type: SchemaTypeObject,
				Properties: map[string]*OASchema{
					"regions": {
						Type: SchemaTypeObject,
						AdditionalProperties: &OASchema{
							Type: SchemaTypeObject,
							Properties: map[string]*OASchema{
								"zone":   {Type: SchemaTypeString},
								"active": {Type: SchemaTypeBoolean},
							},
						},
					},
				},
			},
			remove: []string{"regions/zone"},
			want: map[string]*Item{
				"regions": {
					Properties: map[string]*Item{},
					Items: &Item{
						Properties: map[string]*Item{
							"active": {Properties: map[string]*Item{}}, // "zone" removed
						},
					},
				},
			},
		},
		{
			name: "glob matches within a segment only",
			schema: &OASchema{
				Type: SchemaTypeObject,
				Properties: map[string]*OASchema{
					"azure_tenant_id":       {Type: SchemaTypeString},
					"aiven_azure_principal": {Type: SchemaTypeString},
					"aws_iam_role_arn":      {Type: SchemaTypeString},
					"plans": {
						Type: SchemaTypeArray,
						Items: &OASchema{
							Type: SchemaTypeObject,
							Properties: map[string]*OASchema{
								"azure_sku": {Type: SchemaTypeString},
								"name":      {Type: SchemaTypeString},
							},
						},
					},
				},
			},
			remove: []string{"*azure*"},
			want: map[string]*Item{
				"aws_iam_role_arn": {Properties: map[string]*Item{}},
				"plans": {
					Properties: map[string]*Item{},
					Items: &Item{
						Properties: map[string]*Item{
							// "*" does not cross "/", so the nested field survives
							"azure_sku": {Properties: map[string]*Item{}},
							"name":      {Properties: map[string]*Item{}},
						},
					},
				},
			},
		},
		{
			name: "doublestar glob crosses segments",
			schema: &OASchema{
				Type: SchemaTypeObject,
				Properties: map[string]*OASchema{
					"errors": {Type: SchemaTypeString},
					"plans": {
						Type: SchemaTypeArray,
						Items: &OASchema{
							Type: SchemaTypeObject,
							Properties: map[string]*OASchema{
								"errors": {Type: SchemaTypeString},
								"name":   {Type: SchemaTypeString},
							},
						},
					},
				},
			},
			remove: []string{"**/errors"},
			want: map[string]*Item{
				"plans": {
					Properties: map[string]*Item{},
					Items: &Item{
						Properties: map[string]*Item{
							"name": {Properties: map[string]*Item{}},
						},
					},
				},
			},
		},
	}

	transformer := cmp.Transformer("Item", func(i *Item) any {
		if i == nil {
			return nil
		}

		return struct {
			Properties map[string]*Item
			Items      *Item
		}{
			Properties: i.Properties,
			Items:      i.Items,
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Item{Name: "root", Type: SchemaTypeObject, Properties: make(map[string]*Item)}
			scope := &Scope{Definition: &Definition{Remove: tt.remove}}

			err := fromSchema(scope, &Operation{}, root, tt.schema, ReadHandler)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.want, root.Properties, transformer); diff != "" {
				t.Errorf("properties mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOperationsIDAppearsIn(t *testing.T) {
	operations := Operations{
		{ID: "FooCreate", Type: OperationCreate},
		{ID: "FooRead", Type: OperationRead},
		{ID: "FooUpdate", Type: OperationUpdate},
		{ID: "FooDelete", Type: OperationDelete},
	}

	sources := listSources()
	handlers := operationToHandler()
	seen := make(map[AppearsIn]bool)

	for _, op := range operations {
		for i, source := range sources {
			t.Run(fmt.Sprintf("%s_%d_source", op.ID, i), func(t *testing.T) {
				this := operations.AppearsInID(op.ID, op.Type, source)
				require.NotEqual(t, 0, this)
				require.Positive(t, this&handlers[op.Type])
				require.Positive(t, this&source)

				// Same as by handler type
				require.Equal(t,
					fmt.Sprintf("%b", this),
					fmt.Sprintf("%b", operations.handlerMask(op.Type, source)),
				)

				// Proves the mask doesn't contain other operations: create, update, etc
				for opType, a := range handlers {
					if opType != op.Type {
						require.EqualValues(t, 0, this&a)
					}
				}

				// Proves the mask doesn't contain other sources: RequestBody, ResponseBody, etc
				for _, s := range sources {
					if s != source {
						require.EqualValues(t, 0, this&s)
					}
				}

				// Stores to make sure all values are unique
				require.False(t, seen[this])
				seen[this] = true
			})
		}
	}
}

// TestOperationsIDAppearsInUpsert tests when the same operation ID is used for both create and update (upsert pattern)
func TestOperationsIDAppearsInUpsert(t *testing.T) {
	operations := Operations{
		{ID: "FooUpsert", Type: OperationCreate},
		{ID: "FooRead", Type: OperationRead},
		{ID: "FooUpsert", Type: OperationUpdate}, // Same ID as create
		{ID: "FooDelete", Type: OperationDelete},
	}

	sources := listSources()
	handlers := operationToHandler()
	seen := make(map[AppearsIn]bool)

	for _, op := range operations {
		for i, source := range sources {
			t.Run(fmt.Sprintf("%s_%s_%d_source", op.ID, op.Type, i), func(t *testing.T) {
				this := operations.AppearsInID(op.ID, op.Type, source)
				require.NotEqual(t, 0, this)
				require.Positive(t, this&handlers[op.Type])
				require.Positive(t, this&source)

				// Proves the mask doesn't contain other operations: create, update, etc
				for opType, a := range handlers {
					if opType != op.Type {
						require.EqualValues(t, 0, this&a)
					}
				}

				// Proves the mask doesn't contain other sources: RequestBody, ResponseBody, etc
				for _, s := range sources {
					if s != source {
						require.EqualValues(t, 0, this&s)
					}
				}

				// Stores to make sure all values are unique
				require.False(t, seen[this])
				seen[this] = true
			})
		}
	}

	// Test that handlerMask merges both create and update when same ID is used
	for _, source := range sources {
		createMask := operations.AppearsInID("FooUpsert", OperationCreate, source)
		updateMask := operations.AppearsInID("FooUpsert", OperationUpdate, source)

		// Create and update masks should be different
		require.NotEqual(t, createMask, updateMask)

		// handlerMask should return only the specific handler's mask
		require.Equal(t,
			fmt.Sprintf("%b", createMask),
			fmt.Sprintf("%b", operations.handlerMask(OperationCreate, source)),
		)
		require.Equal(t,
			fmt.Sprintf("%b", updateMask),
			fmt.Sprintf("%b", operations.handlerMask(OperationUpdate, source)),
		)
	}
}

// TestApplySchemaOverride mirrors the aiven_project_vpc datasource overlay: it
// relaxes a Required API field (`project`) to Optional+Computed and adds a
// brand-new alias field (`vpc_id`) that does not exist in the base API schema.
// The test pins down the contract that the FromSchemaOverride flag is set on
// every touched item, that user-set overrides win over the entity-aware
// defaults, that the original root is never mutated, and that items outside
// the overlay are left alone.
func TestApplySchemaOverride(t *testing.T) {
	def := &Definition{
		typeName: "aiven_project_vpc",
		Operations: Operations{
			{ID: "VpcCreate", Type: OperationCreate},
			{ID: "VpcGet", Type: OperationRead},
			{ID: "VpcDelete", Type: OperationDelete},
		},
		IDAttributeComposed: []string{"project", "project_vpc_id"},
		Datasource: &SchemaMeta{
			SchemaOverride: map[string]*Item{
				"project": {
					OverrideOptional: lo.ToPtr(true),
					OverrideComputed: lo.ToPtr(true),
				},
				"vpc_id": {
					Type:               SchemaTypeString,
					OverrideOptional:   lo.ToPtr(true),
					Description:        "The ID of the VPC in `project/project_vpc_id` format.",
					DeprecationMessage: "Use `project_vpc_id` instead.",
					ConflictsWith:      []string{"project", "project_vpc_id"},
				},
			},
		},
	}

	createPath := def.Operations.AppearsInID("VpcCreate", OperationCreate, PathParameter)
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"project": {
				Name:        "project",
				Type:        SchemaTypeString,
				AppearsIn:   createPath,
				Required:    true,
				IDAttribute: true,
				Description: "The name of the project this resource belongs to.",
			},
			"cloud_name": {
				Name:        "cloud_name",
				Type:        SchemaTypeString,
				Description: "Cloud region.",
			},
		},
	}
	setParents(root, nil)

	out, err := applySchemaOverride(def, def.Datasource, root)
	require.NoError(t, err)
	require.NotSame(t, root, out, "applySchemaOverride must not mutate the original root")

	require.False(t, root.Properties["project"].FromSchemaOverride,
		"original root.project must not be tagged from the overlay")
	require.NotContains(t, root.Properties, "vpc_id",
		"original root must not gain overlay-only attributes")

	project := out.Properties["project"]
	require.NotNil(t, project)
	assert.True(t, project.FromSchemaOverride, "project: FromSchemaOverride must be set after overlay merge")
	assert.True(t, project.Optional, "project: user-set optional must win over the entity-aware default")
	assert.True(t, project.Computed, "project: user-set computed must win over the entity-aware default")
	assert.False(t, project.Required, "project: optional+computed override must clear Required")

	vpcID := out.Properties["vpc_id"]
	require.NotNil(t, vpcID, "vpc_id: overlay-introduced attribute must be present in the copy")
	assert.True(t, vpcID.FromSchemaOverride)
	assert.True(t, vpcID.Optional)
	assert.False(t, vpcID.Required)
	assert.Equal(t, SchemaTypeString, vpcID.Type)
	assert.Equal(t, "Use `project_vpc_id` instead.", vpcID.DeprecationMessage)
	assert.Equal(t, []string{"project", "project_vpc_id"}, vpcID.ConflictsWith)

	cloudName := out.Properties["cloud_name"]
	require.NotNil(t, cloudName)
	assert.False(t, cloudName.FromSchemaOverride, "cloud_name: untouched by the overlay")
}

// TestApplySchemaOverrideNoOp asserts that an empty (or nil) overlay returns
// the input root verbatim without a deep copy.
func TestApplySchemaOverrideNoOp(t *testing.T) {
	def := &Definition{Operations: Operations{{ID: "Get", Type: OperationRead}}}
	root := &Item{Name: "root", Type: SchemaTypeObject, Properties: map[string]*Item{}}

	out, err := applySchemaOverride(def, nil, root)
	require.NoError(t, err)
	require.Same(t, root, out, "nil SchemaMeta must short-circuit without copying")

	out, err = applySchemaOverride(def, &SchemaMeta{}, root)
	require.NoError(t, err)
	require.Same(t, root, out, "empty SchemaOverride must short-circuit without copying")
}

// TestDeepCopyItem guards the yaml round-trip used by applySchemaOverride.
// Item carries runtime state on underscore-prefixed yaml tags so the copy must
// preserve every public-facing and internal field set on the original, and
// the copy's Parent pointers must point at the new tree (not the old one).
// If a new Item field is added in the future without a yaml tag, this test
// should fail and alert the author.
func TestDeepCopyItem(t *testing.T) {
	leaf := &Item{
		Name:        "leaf",
		JSONName:    "leaf",
		Type:        SchemaTypeString,
		Description: "leaf description",
		Required:    true,
		Computed:    true,
		AppearsIn:   ReadHandler | ResponseBody,
	}
	arr := &Item{
		Name:     "items",
		JSONName: "items",
		Type:     SchemaTypeArray,
		Items:    leaf,
	}
	root := &Item{
		Name:                       "root",
		JSONName:                   "root",
		Type:                       SchemaTypeObject,
		Description:                "root description",
		DeprecationMessage:         "deprecated",
		OverrideOptional:           lo.ToPtr(true),
		OverrideRequired:           lo.ToPtr(false),
		OverrideUseStateForUnknown: new(true),
		ConflictsWith:              []string{"a", "b"},
		FromSchemaOverride:         true,
		MinLength:                  3,
		MaxLength:                  64,
		Enum:                       []any{"x", "y"},
		Properties: map[string]*Item{
			"items":     arr,
			"id":        {Name: "id", JSONName: "id", Type: SchemaTypeString, IDAttribute: true, IDAttributePosition: 0},
			"is_active": {Name: "is_active", JSONName: "is_active", Type: SchemaTypeBoolean, Optional: true},
		},
	}
	setParents(root, nil)

	cp, err := deepCopyItem(root)
	require.NoError(t, err)
	require.NotSame(t, root, cp)

	// Parent is yaml:"-" so it is intentionally not serialized; cmp.Diff with
	// Parent ignored is the right contract.
	ignoreParent := cmpopts.IgnoreFields(Item{}, "Parent")
	if diff := cmp.Diff(root, cp, ignoreParent); diff != "" {
		t.Fatalf("deepCopyItem mismatch (-want +got):\n%s", diff)
	}

	assert.Nil(t, cp.Parent, "root: Parent must be nil in the copy")
	assert.Same(t, cp, cp.Properties["items"].Parent, "Parent must be re-wired into the new tree")
	assert.Same(t, cp.Properties["items"], cp.Properties["items"].Items.Parent)

	// Mutating the copy must not leak back into the original.
	cp.Properties["items"].Items.Description = "mutated"
	assert.Equal(t, "leaf description", leaf.Description)
}

func TestExampleArrayElems(t *testing.T) {
	def := &Definition{}
	stringItems := &Item{Type: SchemaTypeString}

	t.Run("list-valued example renders every element", func(t *testing.T) {
		item := &Item{Type: SchemaTypeArrayOrdered, Items: stringItems, Example: []any{"+@write", "+@keyspace"}}
		elems, err := exampleArrayElems(def, item)
		require.NoError(t, err)
		got := make([]string, len(elems))
		for i, e := range elems {
			got[i] = e.AsString()
		}
		assert.Equal(t, []string{"+@write", "+@keyspace"}, got)
	})

	t.Run("falls back to a single items example", func(t *testing.T) {
		item := &Item{Type: SchemaTypeArrayOrdered, Items: &Item{Type: SchemaTypeString, Example: "some*chan"}}
		elems, err := exampleArrayElems(def, item)
		require.NoError(t, err)
		require.Len(t, elems, 1)
		assert.Equal(t, "some*chan", elems[0].AsString())
	})

	t.Run("ignores flattened-object items and uses items example", func(t *testing.T) {
		// e.g. billing emails flatten an object list to a scalar list: the items schema
		// is a scalar string that still carries the object's "email" property, and the
		// spec example is a list of objects that doesn't fit a scalar element.
		item := &Item{
			Type:    SchemaTypeArrayOrdered,
			Example: []any{map[string]any{"email": "jane@example.com"}},
			Items: &Item{
				Type:       SchemaTypeString,
				Properties: map[string]*Item{"email": {Type: SchemaTypeString}},
				Example:    "test@example.com",
			},
		}
		elems, err := exampleArrayElems(def, item)
		require.NoError(t, err)
		require.Len(t, elems, 1)
		assert.Equal(t, "test@example.com", elems[0].AsString())
	})
}

func TestExampleScalarItemUUIDFormat(t *testing.T) {
	def := &Definition{}
	for _, example := range []any{nil, "adcf7194-d877-4505-a47a-91fefd96e3b8", "123456789012"} {
		item := &Item{Type: SchemaTypeString, Format: "uuid", Example: example}
		got, err := exampleScalarItem(def, item)
		require.NoError(t, err)
		assert.Equal(t, uuidExample, got.AsString(), "example: %v", example)
	}
}

// todo: drop in v5.0.0 with Item.DropDefault.
func TestMergeItemDropDefault(t *testing.T) {
	t.Parallel()

	parent := &Item{Name: "root", Type: SchemaTypeObject, Properties: map[string]*Item{}}
	got, err := mergeItem(parent, &Item{
		Name:    "restart_enabled",
		Type:    SchemaTypeBoolean,
		Default: true,
	}, &Item{
		Name:        "restart_enabled",
		Type:        SchemaTypeBoolean,
		DropDefault: true,
	})
	require.NoError(t, err)
	require.True(t, got.DropDefault)
	require.Nil(t, got.Default)
}

func TestMergeItemEnum(t *testing.T) {
	t.Parallel()

	openAPI := func() *Item {
		return &Item{Name: "model", Type: SchemaTypeString, AppearsIn: RequestBody, Enum: []any{"a", "b", "c"}}
	}
	parent := &Item{Name: "root", Type: SchemaTypeObject, Properties: map[string]*Item{}}

	got, err := mergeItem(parent, openAPI(), &Item{Name: "model", Type: SchemaTypeString, Enum: []any{"a", "b"}})
	require.NoError(t, err)
	assert.Equal(t, []any{"a", "b"}, got.Enum, "definition enum overrides")

	got, err = mergeItem(parent, openAPI(), &Item{Name: "model", Type: SchemaTypeString, AppearsIn: ResponseBody, Enum: []any{"d"}})
	require.NoError(t, err)
	assert.Equal(t, []any{"a", "b", "c", "d"}, got.Enum, "OpenAPI enums merge")
}

func TestExampleNestedAttributeUsesAssignment(t *testing.T) {
	t.Parallel()

	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
	}
	fileInfo := &Item{
		Name:     "file_info",
		JSONName: "file_info",
		Type:     SchemaTypeObject,
		Computed: true,
		Parent:   root,
		Properties: map[string]*Item{
			"file_sha256": {Name: "file_sha256", JSONName: "file_sha256", Type: SchemaTypeString, Computed: true, Example: "abc"},
			"file_status": {Name: "file_status", JSONName: "file_status", Type: SchemaTypeString, Computed: true, Example: "READY"},
		},
	}
	for _, p := range fileInfo.Properties {
		p.Parent = fileInfo
	}
	source := &Item{
		Name:     "source",
		JSONName: "source",
		Type:     SchemaTypeString,
		Required: true,
		Example:  "./app.jar",
		Parent:   root,
	}
	root.Properties = map[string]*Item{"file_info": fileInfo, "source": source}

	got, err := exampleRoot(&Definition{typeName: "aiven_test"}, resourceType, root)
	require.NoError(t, err)
	s := string(got)
	require.Contains(t, s, "file_info = [{")
	require.NotRegexp(t, `(?m)^\s*file_info \{`, s)
}
