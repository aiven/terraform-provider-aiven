package kafkaschema

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// TestGetSchemaVersion_NotFoundIsRetried covers the registry being eventually consistent:
// right after the POST, the subject (or one of its versions) can still return 404.
func TestGetSchemaVersion_NotFoundIsRetried(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project"
		serviceName = "test-service"
		subjectName = "test-subject"
		schemaID    = 42
	)

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	notFound := avngen.Error{Status: http.StatusNotFound, Message: "Subject not found."}

	// The first poll doesn't see the subject yet
	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subjectName).
		Return(nil, notFound).
		Once()
	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subjectName).
		Return([]int{1, 3, 2}, nil).
		Once()
	// Versions are checked newest first, and one that isn't readable yet is skipped.
	// Version 1 is never read: the mock fails on unexpected calls.
	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subjectName, 3).
		Return(nil, notFound).
		Once()
	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subjectName, 2).
		Return(&kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut{Id: schemaID, Version: 2}, nil).
		Once()

	version, err := getSchemaVersion(ctx, mockClient, project, serviceName, subjectName, schemaID)
	require.NoError(t, err)
	require.Equal(t, 2, version)
}

// TestGetSchemaVersion_OtherErrorsFail makes sure only 404 is retried.
func TestGetSchemaVersion_OtherErrorsFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	forbidden := avngen.Error{Status: http.StatusForbidden, Message: "Forbidden"}

	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionsGet(ctx, "p", "s", "subj").
		Return(nil, forbidden).
		Once()

	_, err := getSchemaVersion(ctx, mockClient, "p", "s", "subj", 1)
	require.ErrorIs(t, err, forbidden)
}

func TestEqualSchemas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{
			name: "json same different whitespace and key order",
			a:    `{"fields":[],"type":"record"}`,
			b:    `{ "type": "record", "fields": [] }`,
			want: true,
		},
		{
			name: "json different content",
			a:    `{"type":"string"}`,
			b:    `{"type":"int"}`,
			want: false,
		},
		{
			name: "json different whitespace in a string",
			a:    `{"type":"record","doc":"a b"}`,
			b:    `{"type":"record","doc":"a  b"}`,
			want: false,
		},
		{
			name: "protobuf whitespace only",
			a:    "syntax = proto3;\n\nmessage M { string f = 1; }",
			b:    "syntax=proto3;message M{string f=1;}",
			want: true,
		},
		{
			name: "protobuf spacing around punctuation",
			a:    "message M{string f=1;}",
			b:    "message M {\n  string f = 1;\n}\n",
			want: true,
		},
		{
			name: "protobuf words joined",
			a:    "message M { string f = 1; }",
			b:    "message M { stringf = 1; }",
			want: false,
		},
		{
			name: "protobuf different content",
			a:    "syntax = proto3;\nmessage M { string f = 1; }",
			b:    "syntax = proto3;\nmessage M { int32 f = 1; }",
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, equalSchemas(tc.a, tc.b))
		})
	}
}

func newModifyPlanData(t *testing.T, plan, state map[string]any) adapter.ResourceData {
	t.Helper()

	opts := []adapter.ResourceDataOpt{
		adapter.WithPreservePlanValues(),
		adapter.WithTestPlan(plan),
	}
	if state != nil {
		opts = append(opts, adapter.WithTestState(state))
	}
	d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields, opts...)
	require.NoError(t, err)
	if state != nil {
		// Terraform marks the unconfigured computed attributes of a changed resource unknown.
		require.NoError(t, d.SetUnknown("version"))
	}
	return d
}

func TestModifyPlan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const (
		proj    = "p"
		svc     = "s"
		subj    = "subj"
		schemaA = `{"type":"string"}`
		schemaB = `{ "type": "string" }`
	)

	existing := func(schemaPlan, schemaState string, refsPlan, refsState any) (map[string]any, map[string]any) {
		base := map[string]any{
			"id":           fmt.Sprintf("%s/%s/%s", proj, svc, subj),
			"project":      proj,
			"service_name": svc,
			"subject_name": subj,
			"version":      3,
			"schema_type":  "AVRO",
		}
		plan := make(map[string]any, len(base)+2)
		state := make(map[string]any, len(base)+2)
		for k, v := range base {
			plan[k], state[k] = v, v
		}
		plan["schema"], state["schema"] = schemaPlan, schemaState
		if refsPlan != nil {
			plan["references"] = refsPlan
		}
		if refsState != nil {
			state["references"] = refsState
		}
		return plan, state
	}

	t.Run("formatting only unchanged version", func(t *testing.T) {
		t.Parallel()
		plan, state := existing(schemaB, schemaA, nil, nil)
		d := newModifyPlanData(t, plan, state)
		require.NoError(t, modifyPlan(ctx, nil, d))
		require.Equal(t, 3, d.Get("version"))
	})

	t.Run("references changed version unknown", func(t *testing.T) {
		t.Parallel()
		refsState := []any{map[string]any{"name": "a", "subject": "s", "version": 1}}
		refsPlan := []any{map[string]any{"name": "b", "subject": "t", "version": 2}}
		plan, state := existing(schemaA, schemaA, refsPlan, refsState)
		d := newModifyPlanData(t, plan, state)
		require.NoError(t, modifyPlan(ctx, nil, d))
		require.Equal(t, 0, d.Get("version"))
	})

	unknownVersionRefs := []any{map[string]any{
		"name":    "a",
		"subject": "s",
		"version": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	}}

	t.Run("unknown reference version on create", func(t *testing.T) {
		t.Parallel()
		plan := map[string]any{
			"project":      proj,
			"service_name": svc,
			"subject_name": subj,
			"schema_type":  "AVRO",
			"schema":       schemaA,
			"references":   unknownVersionRefs,
		}
		d := newModifyPlanData(t, plan, nil)
		require.NoError(t, modifyPlan(ctx, nil, d))
	})

	t.Run("unknown reference version on update", func(t *testing.T) {
		t.Parallel()
		refsState := []any{map[string]any{"name": "a", "subject": "s", "version": 1}}
		plan, state := existing(schemaA, schemaA, unknownVersionRefs, refsState)
		d := newModifyPlanData(t, plan, state)
		require.NoError(t, modifyPlan(ctx, nil, d))
		require.Equal(t, 0, d.Get("version"))
	})

	t.Run("new resource invalid avro", func(t *testing.T) {
		t.Parallel()
		plan := map[string]any{
			"project":      proj,
			"service_name": svc,
			"subject_name": subj,
			"schema_type":  "AVRO",
			"schema":       `{"type": "nope"}`,
		}
		d := newModifyPlanData(t, plan, nil)
		require.ErrorContains(t, modifyPlan(ctx, nil, d), "schema validation error")
	})

	t.Run("incompatible schema", func(t *testing.T) {
		t.Parallel()
		plan, state := existing(`{"type":"int"}`, schemaA, nil, nil)
		mockClient := avngen.NewMockClient(t)
		mockClient.EXPECT().
			ServiceSchemaRegistryCompatibility(mock.Anything, proj, svc, subj, 3, mock.Anything).
			Return(&kafkaschemaregistry.ServiceSchemaRegistryCompatibilityOut{
				IsCompatible: false,
				Messages:     []string{"boom"},
			}, nil).
			Once()
		d := newModifyPlanData(t, plan, state)
		require.EqualError(
			t,
			modifyPlan(ctx, mockClient, d),
			`subject "subj": schema is not compatible with previous version 3: boom`,
		)
	})

	t.Run("empty schema skipped", func(t *testing.T) {
		t.Parallel()
		plan, state := existing("", schemaA, nil, nil)
		d := newModifyPlanData(t, plan, state)
		require.NoError(t, modifyPlan(ctx, nil, d))
		require.Equal(t, 0, d.Get("version"))
	})

	for _, tc := range []struct {
		name, planType, stateType string
		replace                   bool
	}{
		{name: "configured type change replaces", planType: "AVRO", stateType: "PROTOBUF", replace: true},
		{name: "SDKv2 import without a type stays", planType: "PROTOBUF", stateType: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			plan, state := existing(schemaA, schemaA, nil, nil)
			plan["schema_type"], state["schema_type"] = tc.planType, tc.stateType
			d := newModifyPlanData(t, plan, state)
			require.NoError(t, modifyPlan(ctx, nil, d))
			require.Equal(t, tc.replace, slices.Contains(adapter.RequiresReplaceKeys(d), "schema_type"))
		})
	}
}

func TestExpandReferences(t *testing.T) {
	t.Parallel()

	// Get drops values unknown at plan time, like the version of "c".
	refs, err := expandReferences([]any{
		map[string]any{"name": "b", "subject": "s", "version": 1},
		map[string]any{"name": "a", "subject": "s", "version": 2},
		map[string]any{"name": "c", "subject": "s"},
	})
	require.NoError(t, err)
	require.Equal(
		t,
		[]kafkaschemaregistry.ReferenceIn{
			{Name: "a", Subject: "s", Version: 2},
			{Name: "b", Subject: "s", Version: 1},
			{Name: "c", Subject: "s"},
		},
		refs,
	)
}

func TestCreateView(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := avngen.NewMockClient(t)
	post := client.EXPECT().
		ServiceSchemaRegistrySubjectVersionPost(ctx, "p", "s", "subj", mock.Anything).
		Return(7, nil).
		Once()
	client.EXPECT().
		ServiceSchemaRegistrySubjectVersionsGet(ctx, "p", "s", "subj").
		Return([]int{1}, nil).
		Once()
	client.EXPECT().
		ServiceSchemaRegistrySubjectVersionGet(ctx, "p", "s", "subj", 1).
		Return(&kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut{Id: 7, Version: 1}, nil).
		Once()
	client.EXPECT().
		ServiceSchemaRegistrySubjectConfigPut(
			ctx,
			"p",
			"s",
			"subj",
			&kafkaschemaregistry.ServiceSchemaRegistrySubjectConfigPutIn{Compatibility: kafkaschemaregistry.CompatibilityTypeFull},
		).
		Return(kafkaschemaregistry.CompatibilityTypeFull, nil).
		Once().
		NotBefore(post)

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(map[string]any{
			"project":             "p",
			"service_name":        "s",
			"subject_name":        "subj",
			"schema_type":         "AVRO",
			"schema":              `{"type":"string"}`,
			"compatibility_level": "FULL",
		}),
	)
	require.NoError(t, err)
	require.NoError(t, createView(ctx, client, d))
	require.Equal(t, 1, d.Get("version"))
	require.Equal(t, "p/s/subj", d.ID())
}

func TestUpdateView(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// newData returns an existing resource whose plan sets the schema and compatibility level.
	newData := func(t *testing.T, schema, compatibility string) adapter.ResourceData {
		t.Helper()
		state := map[string]any{
			"id":                  "p/s/subj",
			"project":             "p",
			"service_name":        "s",
			"subject_name":        "subj",
			"version":             3,
			"schema_type":         "AVRO",
			"schema":              `{"type":"string"}`,
			"compatibility_level": "FORWARD",
		}
		plan := maps.Clone(state)
		plan["schema"] = schema
		plan["compatibility_level"] = compatibility
		d, err := adapter.NewResourceData(
			resourceSchemaInternal(),
			ResourceOptions.IdentityFields,
			adapter.WithTestPlan(plan),
			adapter.WithTestState(state),
		)
		require.NoError(t, err)
		return d
	}

	t.Run("formatting only", func(t *testing.T) {
		t.Parallel()
		d := newData(t, `{ "type": "string" }`, "FORWARD")
		require.ErrorIs(t, updateView(ctx, nil, d), adapter.ErrNoChanges)
	})

	t.Run("compatibility only", func(t *testing.T) {
		t.Parallel()
		// The mock fails on any other call, e.g., registering a new version.
		client := avngen.NewMockClient(t)
		client.EXPECT().
			ServiceSchemaRegistrySubjectConfigPut(
				ctx,
				"p",
				"s",
				"subj",
				&kafkaschemaregistry.ServiceSchemaRegistrySubjectConfigPutIn{Compatibility: kafkaschemaregistry.CompatibilityTypeFull},
			).
			Return(kafkaschemaregistry.CompatibilityTypeFull, nil).
			Once()
		d := newData(t, `{"type":"string"}`, "FULL")
		require.NoError(t, updateView(ctx, client, d))
		require.Equal(t, 3, d.Get("version"))
	})
}

func TestReadView_preservesConfiguredSchemaFormatting(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-kafkaschema-readview"
		serviceName = "test-service-kafkaschema-readview"
		subjectName = "subj"
	)
	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	notFound := avngen.Error{Status: http.StatusNotFound, Message: "not found"}

	mockClient.EXPECT().
		ServiceGet(ctx, project, serviceName).
		Return(&service.ServiceGetOut{UserConfig: map[string]any{"schema_registry": true}}, nil).
		Maybe()
	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subjectName, 3).
		Return(&kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut{
			Schema:  `{"type":"string"}`,
			Version: 3,
		}, nil).
		Once()
	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectConfigGet(mock.Anything, project, serviceName, subjectName, mock.Anything).
		Return(kafkaschemaregistry.CompatibilityType(""), notFound).
		Once()

	stateSchema := `{ "type": "string" }`
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestState(map[string]any{
			"id":           fmt.Sprintf("%s/%s/%s", project, serviceName, subjectName),
			"project":      project,
			"service_name": serviceName,
			"subject_name": subjectName,
			"version":      3,
			"schema":       stateSchema,
		}),
	)
	require.NoError(t, err)
	require.NoError(t, readView(ctx, mockClient, d))
	require.Equal(t, stateSchema, d.Get("schema"))
	require.Equal(t, "AVRO", d.Get("schema_type"))
	require.Empty(t, d.Get("compatibility_level"))
	require.Empty(t, d.Get("references"))
}
