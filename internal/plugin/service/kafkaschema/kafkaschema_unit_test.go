package kafkaschema

import (
	"context"
	"encoding/base64"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/bufbuild/protocompile/options"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

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
	require.Equal(t, 2, version.Version)
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
			a:    "syntax = \"proto3\";\n\nmessage M { string f = 1; }",
			b:    "syntax=\"proto3\";message M{string f=1;}",
			want: true,
		},
		{
			name: "protobuf spacing around punctuation",
			a:    "syntax=\"proto3\";message M{string f=1;}",
			b:    "syntax = \"proto3\"; message M {\n  string f = 1;\n}\n",
			want: true,
		},
		{
			name: "protobuf words joined",
			a:    "message M { string f = 1; }",
			b:    "message M { stringf = 1; }",
			want: false,
		},
		{
			name: "protobuf line comment separator",
			a:    "syntax=\"proto3\";\n//Order\nmessage Order {}",
			b:    "syntax=\"proto3\";\n// Order\nmessage Order {}",
			want: true,
		},
		{
			name: "comment after syntax documents next declaration",
			a:    "syntax=\"proto3\"; // Order\nmessage Order {}",
			b:    "syntax=\"proto3\";\n// Order\nmessage Order {}",
			want: true,
		},
		{
			name: "protobuf comment contents",
			a:    "syntax=\"proto3\";\n//Order  ID\nmessage Order {}",
			b:    "syntax=\"proto3\";\n// Order ID\nmessage Order {}",
			want: false,
		},
		{
			name: "protobuf different content",
			a:    "syntax = \"proto3\";\nmessage M { string f = 1; }",
			b:    "syntax = \"proto3\";\nmessage M { int32 f = 1; }",
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

func TestSameRegistrationText(t *testing.T) {
	t.Parallel()
	require.True(t, sameRegistrationText(`{"type":"number","maximum":1.0}`, `{ "maximum": 1.0, "type": "number" }`))
	require.False(t, sameRegistrationText(`{"type":"number","maximum":1.0}`, `{"type":"number","maximum":1}`))
	require.True(t, sameRegistrationText(`{"maximum":1.50}`, `{"maximum":1.5}`))
	require.True(t, sameRegistrationText(`{"maximum":15e-1}`, `{"maximum":1.5}`))
	require.False(t, sameRegistrationText(`{"maximum":0.10000000000000000001}`, `{"maximum":0.1}`))
	require.False(t, sameRegistrationText(`{"maximum":-0.0}`, `{"maximum":0.0}`))
	require.True(t, sameRegistrationText(`{"maximum":-0e3}`, `{"maximum":-0.0}`))
	require.True(t, sameRegistrationText(`{"maximum":-0}`, `{"maximum":0}`))
	const header = "syntax = \"proto3\";\n"
	const enum = "// Delivery states\nenum Status { READY = 0; }\n"
	const message = "// Event\nmessage M { Status state = 1; }\n"
	require.True(t, sameRegistrationText(header+enum+message, header+message+enum))
	require.False(t, sameRegistrationText(header+enum+message, header+`message M { Status state = 1; }`+"\n// Event\n"+enum))
	require.False(t, sameRegistrationText(header+`message A {} message B {}`, header+`message B {} message A {}`))
	const descriptor = "CgdtLnByb3RvIhUKAU0SEAoIb3JkZXJfaWQYASABKAliBnByb3RvMw=="
	require.True(t, sameRegistrationText(descriptor, header+`message M { string order_id = 1; }`))
	require.False(t, sameRegistrationText(descriptor, header+"message M {\n// Order identifier\nstring order_id = 1; }"))
	require.False(t, sameRegistrationText(descriptor, header+`message M { string order_id = 1 [json_name = "orderId"]; }`))
}

func TestSerializedRegistrationText(t *testing.T) {
	const source = `syntax = "proto3";
package orders;
option java_package = "orders";
option go_package = "example/orders";
message Order {
  reserved 3, 5 to 7;
  optional string key = 1;
  oneof payload { bytes value = 2; }
  message Nested {}
  enum Status { READY = 0 [deprecated = true]; }
}`
	parsed, err := parseProto(source)
	require.NoError(t, err)
	_, err = options.InterpretUnlinkedOptions(parsed)
	require.NoError(t, err)
	data, err := proto.Marshal(parsed.FileDescriptorProto())
	require.NoError(t, err)
	compiled := base64.StdEncoding.EncodeToString(data)
	require.True(t, sameRegistrationText(compiled, source))
	require.False(t, sameRegistrationText(compiled, strings.Replace(source, "reserved 3, 5 to 7;", "reserved 3; reserved 5 to 7;", 1)))
	require.False(t, sameRegistrationText(compiled, strings.Replace(source, "optional string key", "// Partition key\noptional string key", 1)))
	require.False(t, sameRegistrationText(compiled, strings.Replace(source,
		"option java_package = \"orders\";\noption go_package = \"example/orders\";",
		"option go_package = \"example/orders\";\noption java_package = \"orders\";", 1)))
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

// Terraform carries the private plan into a second RPC with null prior state on replacement.
func planWithReplacement(t *testing.T, ctx context.Context, client avngen.Client, first adapter.ResourceData, config map[string]any) error {
	t.Helper()
	if err := modifyPlan(ctx, client, first); err != nil {
		return err
	}
	pending, err := first.GetPrivate(ctx, replacementSourceKey)
	require.NoError(t, err)
	if len(pending) == 0 {
		return nil
	}
	create := newModifyPlanData(t, config, nil)
	require.NoError(t, create.SetPrivate(ctx, replacementSourceKey, pending))
	return modifyPlan(ctx, client, create)
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
	t.Run("incompatible type replacement", func(t *testing.T) {
		plan, state := existing(schemaA, schemaA, nil, nil)
		plan["schema_type"] = "JSON"
		client := avngen.NewMockClient(t)
		client.EXPECT().ServiceSchemaRegistryCompatibility(mock.Anything, proj, svc, subj, 3, mock.Anything).
			Return(&kafkaschemaregistry.ServiceSchemaRegistryCompatibilityOut{IsCompatible: false}, nil)
		require.ErrorContains(t, planWithReplacement(t, ctx, client, newModifyPlanData(t, plan, state), plan), "is not compatible")
	})
	t.Run("incompatible rename to existing subject", func(t *testing.T) {
		plan, state := existing(schemaA, schemaA, nil, nil)
		plan["subject_name"] = "destination"
		client := avngen.NewMockClient(t)
		client.EXPECT().ServiceSchemaRegistrySubjectVersionsGet(mock.Anything, proj, svc, "destination").Return([]int{1, 2}, nil)
		client.EXPECT().ServiceSchemaRegistryCompatibility(mock.Anything, proj, svc, "destination", 2, mock.Anything).
			Return(&kafkaschemaregistry.ServiceSchemaRegistryCompatibilityOut{IsCompatible: false}, nil)
		client.EXPECT().ServiceSchemaRegistrySubjectVersionGet(mock.Anything, proj, svc, "destination", 1).
			Return(&kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut{Schema: schemaA, References: []kafkaschemaregistry.ReferenceOut{{Name: "dep", Subject: "dep", Version: 1}}}, nil)
		client.EXPECT().ServiceSchemaRegistrySubjectVersionGet(mock.Anything, proj, svc, "destination", 2).
			Return(&kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut{Schema: `{"type":"int"}`}, nil)
		require.ErrorContains(t, planWithReplacement(t, ctx, client, newModifyPlanData(t, plan, state), plan), `subject "destination": schema is not compatible with previous version 2`)
	})
	t.Run("move without type validates the create default", func(t *testing.T) {
		plan, state := existing(schemaA, schemaA, nil, nil)
		state["schema_type"], plan["schema_type"], plan["subject_name"] = "JSON", "JSON", "destination"
		client := avngen.NewMockClient(t)
		client.EXPECT().ServiceSchemaRegistrySubjectVersionsGet(mock.Anything, proj, svc, "destination").Return([]int{1}, nil)
		client.EXPECT().ServiceSchemaRegistrySubjectVersionGet(mock.Anything, proj, svc, "destination", 1).
			Return(&kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut{Schema: schemaA, SchemaType: "JSON"}, nil)
		client.EXPECT().ServiceSchemaRegistryCompatibility(mock.Anything, proj, svc, "destination", 1, mock.Anything).
			RunAndReturn(func(_ context.Context, _, _, _ string, _ int, in *kafkaschemaregistry.ServiceSchemaRegistryCompatibilityIn) (*kafkaschemaregistry.ServiceSchemaRegistryCompatibilityOut, error) {
				return &kafkaschemaregistry.ServiceSchemaRegistryCompatibilityOut{IsCompatible: in.SchemaType == "JSON"}, nil
			})
		config := maps.Clone(plan)
		delete(config, "schema_type")
		require.ErrorContains(t, planWithReplacement(t, ctx, client, newModifyPlanData(t, plan, state), config), "is not compatible")
	})
	t.Run("move to a registry being enabled", func(t *testing.T) {
		plan, state := existing(schemaA, schemaA, nil, nil)
		plan["service_name"] = "another-kafka"
		client := avngen.NewMockClient(t)
		client.EXPECT().ServiceSchemaRegistrySubjectVersionsGet(mock.Anything, proj, "another-kafka", subj).
			Return(nil, avngen.Error{Status: http.StatusForbidden})
		client.EXPECT().ServiceGet(mock.Anything, proj, "another-kafka").
			Return(&service.ServiceGetOut{ServiceType: "kafka", State: service.ServiceStateTypeRunning, UserConfig: map[string]any{"schema_registry": false}}, nil).Maybe()
		client.EXPECT().ServiceSchemaRegistryCompatibility(mock.Anything, proj, svc, subj, 3, mock.Anything).
			Return(&kafkaschemaregistry.ServiceSchemaRegistryCompatibilityOut{IsCompatible: true}, nil).Maybe()
		require.NoError(t, planWithReplacement(t, ctx, client, newModifyPlanData(t, plan, state), plan))
	})
	t.Run("SDK import with an explicitly conflicting type", func(t *testing.T) {
		plan, state := existing(schemaA, schemaA, nil, nil)
		state["schema_type"] = ""
		client := avngen.NewMockClient(t)
		client.EXPECT().ServiceSchemaRegistrySubjectVersionGet(mock.Anything, proj, svc, subj, 3).
			Return(&kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut{SchemaType: "JSON"}, nil)
		client.EXPECT().ServiceSchemaRegistryCompatibility(mock.Anything, proj, svc, subj, 3, mock.Anything).
			Return(&kafkaschemaregistry.ServiceSchemaRegistryCompatibilityOut{IsCompatible: true}, nil).Maybe()
		require.ErrorContains(t, modifyPlan(ctx, client, newModifyPlanData(t, plan, state)), "refresh-only")
	})

	t.Run("known empty avro is rejected", func(t *testing.T) {
		t.Parallel()
		plan, state := existing("", schemaA, nil, nil)
		d := newModifyPlanData(t, plan, state)
		require.ErrorContains(t, modifyPlan(ctx, nil, d), "schema validation error")
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
			client := avngen.NewMockClient(t)
			if tc.stateType == "" {
				client.EXPECT().ServiceSchemaRegistrySubjectVersionGet(ctx, proj, svc, subj, 3).
					Return(&kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut{SchemaType: "PROTOBUF"}, nil).Once()
			}
			if tc.replace {
				client.EXPECT().ServiceSchemaRegistryCompatibility(ctx, proj, svc, subj, 3, mock.Anything).
					Return(&kafkaschemaregistry.ServiceSchemaRegistryCompatibilityOut{IsCompatible: true}, nil).Once()
			}
			require.NoError(t, planWithReplacement(t, ctx, client, d, plan))
			require.Equal(t, tc.replace, slices.Contains(adapter.RequiresReplaceKeys(d), "schema_type"))
		})
	}
}

func TestConfirmDestinationVersions(t *testing.T) {
	const project, name, subject = "project", "kafka", "payments"
	notFound := avngen.Error{Status: http.StatusNotFound}
	denied := avngen.Error{Status: http.StatusForbidden}
	for _, tc := range []struct {
		name        string
		subjects    []string
		listErr     error
		versions    []int
		versionsErr error
		diagnostic  string
	}{
		{name: "listed subject still has no visible versions", subjects: []string{subject}, diagnostic: "not yet visible"},
		{name: "scoped read recovers when listing is denied", listErr: denied, versions: []int{1, 2}},
		{name: "cannot confirm absence", listErr: denied, versionsErr: notFound, diagnostic: "unable to confirm"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			api := client.EXPECT()
			api.ServiceSchemaRegistrySubjectVersionsGet(mock.Anything, project, name, subject).Return(nil, notFound).Once()
			api.ServiceGet(mock.Anything, project, name).Return(&service.ServiceGetOut{
				ServiceType: "kafka", State: service.ServiceStateTypeRunning, UserConfig: map[string]any{"schema_registry": true},
			}, nil)
			api.ServiceSchemaRegistrySubjects(mock.Anything, project, name).Return(tc.subjects, tc.listErr)
			api.ServiceSchemaRegistrySubjectVersionsGet(mock.Anything, project, name, subject).Return(tc.versions, tc.versionsErr)
			versions, err := replacementDestinationVersions(t.Context(), client, project, name, subject)
			if tc.diagnostic != "" {
				require.ErrorContains(t, err, tc.diagnostic)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.versions, versions)
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
	client.EXPECT().ServiceGet(ctx, "p", "s").Return(&service.ServiceGetOut{ServiceType: "kafka"}, nil).Once()
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
		Return(&kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut{Id: 7, Version: 1, Schema: `{"type":"string"}`}, nil).
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
		adapter.WithTestConfig(map[string]any{"compatibility_level": "FULL"}),
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
			adapter.WithTestConfig(map[string]any{"compatibility_level": compatibility}),
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
		client.EXPECT().ServiceGet(ctx, "p", "s").Return(&service.ServiceGetOut{ServiceType: "kafka"}, nil).Once()
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
		Return(&service.ServiceGetOut{ServiceType: "kafka", UserConfig: map[string]any{"schema_registry": true}}, nil).
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

func TestRegisteredProtobufDeclarations(t *testing.T) {
	t.Parallel()
	const syntax = "syntax = \"proto3\";\n"
	t.Run("serialized optional presence", func(t *testing.T) {
		const proto3 = "CgdtLnByb3RvIhYKBU9yZGVyEg0KAmlkGAEgASgJiAEBYgZwcm90bzM="
		const proto2 = "CgdtLnByb3RvIhYKBU9yZGVyEg0KAmlkGAEgASgJiAEBYgZwcm90bzI="
		require.True(t, sameProtobufDeclarations(proto3, syntax+`message Order { optional string id = 1; }`, nil))
		require.False(t, sameProtobufDeclarations(proto3, syntax+`message Order { string id = 1; }`, nil))
		require.False(t, sameProtobufDeclarations(proto3, syntax+`message Order { oneof key { string id = 1; } }`, nil))
		require.True(t, sameProtobufDeclarations(proto2, `syntax = "proto2"; message Order { optional string id = 1; }`, nil))
		require.False(t, sameProtobufDeclarations(proto2, `syntax = "proto2"; message Order { required string id = 1; }`, nil))
	})
	t.Run("descriptor field JSON name", func(t *testing.T) {
		// FileDescriptorProto for M.order_id with no explicit JSON name.
		const descriptor = "CgdtLnByb3RvIhUKAU0SEAoIb3JkZXJfaWQYASABKAliBnByb3RvMw=="
		require.True(t, sameProtobufDeclarations(descriptor, syntax+`message M { string order_id = 1; }`, nil))
		require.False(t, sameProtobufDeclarations(descriptor, syntax+`message M { string order_id = 1 [json_name = "id"]; }`, nil))
	})
	t.Run("proto2 group placement", func(t *testing.T) {
		const proto2 = `syntax = "proto2"; `
		for _, tc := range []struct {
			name, before, after string
			equal               bool
		}{
			{"inside oneof", `message M { oneof key { group G = 1 { optional int32 x = 1; } int32 n = 2; } }`, `message M { oneof key { int32 n = 2; group G = 1 { optional int32 x = 1; } } }`, true},
			// G and N are both addressable nested messages; swapping their indexes changes the wire contract.
			{"changed nested message index", `message M { message N {} optional group G = 1 { optional int32 x = 1; } }`, `message M { optional group G = 1 { optional int32 x = 1; } message N {} }`, false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				_, err := parseProto(proto2 + tc.before)
				require.NoError(t, err)
				require.Equal(t, tc.equal, sameProtobufDeclarations(proto2+tc.before, proto2+tc.after, nil))
			})
		}
	})
	for _, tc := range []struct {
		name, submitted, registry string
		same                      bool
	}{
		{"map entries and nested messages", "message M { message N {} map<string, string> labels = 1; }", "message M { map<string, string> labels = 1; message N {} }", true},
		{"public imports grouped", `import public "a.proto"; import "b.proto"; message M {}`, `import "b.proto"; import public "a.proto"; message M {}`, true},
		{"referenced schema.proto", `import "schema.proto"; message M { External e = 1; } // header`, `import "schema.proto"; message M { External e = 1; }`, true},
		{"public import changed", `import public "a.proto"; import "b.proto";`, `import "a.proto"; import public "b.proto";`, false},
		{"weak import changed", `import weak "a.proto";`, `import "a.proto";`, false},
		{"message index changed", "message A {} message B {}", "message B {} message A {}", false},
		{"enum reservation dropped", "enum E { reserved 5; X = 0; }", "enum E { X = 0; }", false},
		{"changed aggregate option", `message M { option (meta) = { params: { key: "region" value: "EU" } }; }`, `message M { option (meta) = { params: { key: "region", value: "US" } }; }`, false},
		{"repeated option order changed", `message M { option (meta) = { tags: "A" tags: "B" }; }`, `message M { option (meta) = { tags: ["B","A"] }; }`, false},
		{"nonempty RPC option lost", `message M {} service S { rpc Get(M) returns (M) { option deprecated = true; } }`, `message M {} service S { rpc Get(M) returns (M); }`, false},
		{"deprecated enum value lost", "enum E { X = 0 [deprecated = true]; }", "enum E { X = 0; }", false},
		{"explicit packing changed", "message M { repeated int32 x = 1 [packed = false]; }", "message M { repeated int32 x = 1; }", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.same, sameProtobufDeclarations(syntax+tc.submitted, syntax+tc.registry, nil))
		})
	}
}
