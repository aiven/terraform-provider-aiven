package kafkaschema_test

import (
	"context"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	sr "github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const (
	registryProject     = "test-project"
	registrationService = "kafka-registration"
	registrySchema      = `{"type":"string"}`
)

// Only the registry is simulated. Terraform plans, applies and refreshes each subject
// through the provider. Entries retain their reference sequence, just as in a registry fingerprint.
type versionRegistry struct {
	client                      *avngen.MockClient
	subject                     string
	versions                    []sr.ServiceSchemaRegistrySubjectVersionGetOut
	submissions                 []string
	compatibilityWrites         []sr.CompatibilityType
	level                       sr.CompatibilityType
	readsAfterPut, pendingReads []sr.CompatibilityType
	deniedLevel                 sr.CompatibilityType
	render                      func(string) string
	serviceType                 string
	disabled                    bool
	poweredOff                  bool
	omitNextVersion             bool
	omitNextService             bool
	omitNextVersions            bool
	failAfterLevel              sr.CompatibilityType
	dependencies                []sr.ServiceSchemaRegistrySubjectVersionGetOut
}

func (r *versionRegistry) register(schema string, kind sr.SchemaType, refs []sr.ReferenceIn) int {
	if r.render != nil {
		schema = r.render(schema)
	}
	out := make([]sr.ReferenceOut, len(refs))
	for i, ref := range refs {
		out[i] = sr.ReferenceOut(ref)
	}
	for _, v := range r.versions {
		if v.Schema == schema && v.SchemaType == kind && slices.Equal(v.References, out) {
			return v.Id
		}
	}
	n := len(r.versions) + 1
	r.versions = append(r.versions, sr.ServiceSchemaRegistrySubjectVersionGetOut{
		Id: 100 + n, Version: n, Subject: r.subject, Schema: schema, SchemaType: kind, References: out,
	})
	return 100 + n
}

func (r *versionRegistry) testCase(t *testing.T, steps ...resource.TestStep) resource.TestCase {
	t.Helper()
	client := avngen.NewMockClient(t)
	r.client = client
	api := client.EXPECT()
	const project, serviceName = registryProject, registrationService
	if r.subject == "" {
		r.subject = "orders"
	}
	subject := r.subject
	schemautil.ServiceReadyForget(project, serviceName)
	t.Cleanup(func() { schemautil.ServiceReadyForget(project, serviceName) })
	if r.serviceType == "" {
		r.serviceType = "kafka"
	}
	api.ServiceGet(mock.Anything, project, serviceName).RunAndReturn(
		func(context.Context, string, string, ...[2]string) (*service.ServiceGetOut, error) {
			if r.omitNextService {
				r.omitNextService = false
				return nil, avngen.Error{Status: http.StatusNotFound}
			}
			state := service.ServiceStateTypeRunning
			if r.poweredOff {
				state = service.ServiceStateTypePoweroff
			}
			return &service.ServiceGetOut{
				ServiceType: r.serviceType, State: state,
				UserConfig: map[string]any{"schema_registry": r.serviceType == "kafka" && !r.disabled},
			}, nil
		},
	).Maybe()
	api.ServiceGet(mock.Anything, project, "missing-kafka").
		Return(nil, avngen.Error{Status: http.StatusNotFound}).Maybe()
	api.ServiceList(mock.Anything, project).
		Return([]service.ServiceOut{{ServiceName: serviceName, ServiceType: "kafka"}}, nil).Maybe()
	api.ServiceSchemaRegistrySubjectVersionsGet(mock.Anything, project, "missing-kafka", subject).
		Return(nil, avngen.Error{Status: http.StatusNotFound}).Maybe()
	api.ServiceSchemaRegistryCompatibility(mock.Anything, project, serviceName, subject, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, _, _, _ string, _ int, request *sr.ServiceSchemaRegistryCompatibilityIn) (*sr.ServiceSchemaRegistryCompatibilityOut, error) {
			if (request.SchemaType == "" || request.SchemaType == sr.SchemaTypeAvro) && request.Schema == `{"type":"object"}` {
				return nil, avngen.Error{Status: http.StatusUnprocessableEntity, Message: "invalid Avro object"}
			}
			return &sr.ServiceSchemaRegistryCompatibilityOut{IsCompatible: true}, nil
		}).Maybe()
	api.ServiceSchemaRegistrySubjectVersionPost(mock.Anything, project, serviceName, subject, mock.Anything).
		RunAndReturn(func(_ context.Context, _, _, _ string, in *sr.ServiceSchemaRegistrySubjectVersionPostIn) (int, error) {
			r.submissions = append(r.submissions, in.Schema)
			var refs []sr.ReferenceIn
			if in.References != nil {
				refs = *in.References
			}
			kind := in.SchemaType
			if kind == "" {
				kind = sr.SchemaTypeAvro
			}
			return r.register(in.Schema, kind, refs), nil
		}).Maybe()
	api.ServiceSchemaRegistrySubjectVersionsGet(mock.Anything, project, serviceName, subject).
		RunAndReturn(func(context.Context, string, string, string) ([]int, error) {
			if r.omitNextVersions {
				r.omitNextVersions = false
				return nil, avngen.Error{Status: http.StatusNotFound}
			}
			var versions []int
			for _, v := range r.versions {
				versions = append(versions, v.Version)
			}
			return versions, nil
		}).Maybe()
	api.ServiceSchemaRegistrySubjectVersionGet(mock.Anything, project, serviceName, subject, mock.Anything).
		RunAndReturn(func(_ context.Context, _, _, _ string, version int) (*sr.ServiceSchemaRegistrySubjectVersionGetOut, error) {
			if r.failAfterLevel != "" && r.level == r.failAfterLevel {
				return nil, avngen.Error{Status: 500, Message: "refresh after PUT failed"}
			}
			if r.omitNextVersion || version < 1 || version > len(r.versions) {
				r.omitNextVersion = false
				return nil, avngen.Error{Status: http.StatusNotFound}
			}
			v := r.versions[version-1]
			return &v, nil
		}).Maybe()
	for _, dep := range r.dependencies {
		api.ServiceSchemaRegistrySubjectVersionGet(mock.Anything, project, serviceName, dep.Subject, dep.Version).
			Return(&dep, nil).Maybe()
	}
	api.ServiceSchemaRegistrySubjectConfigGet(mock.Anything, project, serviceName, subject, mock.Anything).
		RunAndReturn(func(context.Context, string, string, string, ...[2]string) (sr.CompatibilityType, error) {
			level := r.level
			if len(r.pendingReads) > 0 {
				level, r.pendingReads = r.pendingReads[0], r.pendingReads[1:]
			}
			if level == "" {
				return "", avngen.Error{Status: http.StatusNotFound}
			}
			return level, nil
		}).Maybe()
	api.ServiceSchemaRegistrySubjectConfigPut(mock.Anything, project, serviceName, subject, mock.Anything).
		RunAndReturn(func(_ context.Context, _, _, _ string, in *sr.ServiceSchemaRegistrySubjectConfigPutIn) (sr.CompatibilityType, error) {
			r.compatibilityWrites = append(r.compatibilityWrites, in.Compatibility)
			if in.Compatibility == r.deniedLevel {
				return "", avngen.Error{Status: http.StatusForbidden, Message: "compatibility write denied"}
			}
			r.level = in.Compatibility
			r.pendingReads = slices.Clone(r.readsAfterPut)
			return r.level, nil
		}).Maybe()
	api.ServiceSchemaRegistrySubjectDelete(mock.Anything, project, serviceName, subject, mock.Anything).
		RunAndReturn(func(context.Context, string, string, string, ...[2]string) error {
			r.versions = nil
			return nil
		}).Maybe()
	return resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"aiven": providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client}),
		},
		PreventPostDestroyRefresh: true,
		Steps:                     steps,
	}
}

func registrationConfig(schema, attrs string) string {
	return schemaConfig(registryProject, registrationService, "orders", schema, attrs)
}

// Apply uses the saved plan even if a concurrent writer changes the subject after planning.
type afterRegistrationPlan func()

func (f afterRegistrationPlan) CheckPlan(context.Context, plancheck.CheckPlanRequest, *plancheck.CheckPlanResponse) {
	f()
}

func TestKafkaSchemaRegistration(t *testing.T) {
	const (
		updatedSchema = `{"type":"long"}`
		full          = `compatibility_level = "FULL"`
		none          = `compatibility_level = "NONE"`
	)
	t.Run("retry_after_unconfirmed_create_preserves_history", func(t *testing.T) {
		r := &versionRegistry{subject: "orders", readsAfterPut: []sr.CompatibilityType{"", ""}}
		r.register(updatedSchema, sr.SchemaTypeAvro, nil)
		r.register(registrySchema, sr.SchemaTypeAvro, nil)
		config := registrationConfig(registrySchema, full+"\ntimeouts { create = \"2s\" }")
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: config, ExpectError: regexp.MustCompile("context deadline exceeded")},
			resource.TestStep{Config: config, PreConfig: func() { r.readsAfterPut, r.pendingReads = nil, nil }, Check: resource.TestCheckResourceAttr(schemaAddress, "version", "2")},
		))
	})
	t.Run("registered_source_and_external_drift", func(t *testing.T) {
		const header = "// Order events\n"
		const protoSchema = `syntax = "proto3"; message Order { string id = 1; }`
		source := header + protoSchema
		config := registrationConfig(source, `schema_type = "PROTOBUF"`)
		r := &versionRegistry{render: func(s string) string { return strings.TrimPrefix(s, header) }}
		const external = `syntax = "proto3"; message Order { int32 units = 1; }`
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: config, Check: resource.TestCheckResourceAttr(schemaAddress, "schema", source)},
			resource.TestStep{Config: config, PlanOnly: true},
			resource.TestStep{
				PreConfig: func() {
					r.versions[0].Schema = external
					r.versions[0].Id++
				},
				RefreshState: true, ExpectNonEmptyPlan: true,
				Check: resource.TestCheckResourceAttr(schemaAddress, "schema", external),
			},
			resource.TestStep{
				PreConfig:    func() { r.versions[0].Schema, r.versions[0].Id = protoSchema, r.versions[0].Id-1 },
				RefreshState: true, ExpectNonEmptyPlan: true,
				Check: resource.TestCheckResourceAttr(schemaAddress, "schema", protoSchema),
			},
		))
	})
	t.Run("known_level_on_formatting_update", func(t *testing.T) {
		r := &versionRegistry{level: sr.CompatibilityTypeFull}
		config := registrationConfig(`{ "type" : "string" }`, "") + `
resource "terraform_data" "by_level" {
  for_each = toset([aiven_kafka_schema.orders.compatibility_level])
  input    = each.key
}
`
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: registrationConfig(registrySchema, "")},
			resource.TestStep{
				Config: config, PlanOnly: true, ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(`terraform_data.by_level["FULL"]`, plancheck.ResourceActionCreate),
				}},
			},
		))
		require.Equal(t, []string{registrySchema}, r.submissions)
	})
	t.Run("null_level_on_formatting_update", func(t *testing.T) {
		r := &versionRegistry{}
		formatted := registrationConfig(`{ "type" : "string" }`, "")
		keyed := formatted + `
resource "terraform_data" "by_level" {
  for_each = toset([coalesce(aiven_kafka_schema.orders.compatibility_level, "default")])
  input    = each.key
}
`
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: registrationConfig(registrySchema, "")},
			resource.TestStep{
				Config: keyed, PlanOnly: true, ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(`terraform_data.by_level["default"]`, plancheck.ResourceActionCreate),
				}},
			},
			resource.TestStep{Config: formatted, Check: resource.TestCheckResourceAttr(schemaAddress, "version", "1")},
		))
		require.Equal(t, []string{registrySchema}, r.submissions)
		require.Empty(t, r.compatibilityWrites)
	})
	t.Run("package_comment_keeps_version", func(t *testing.T) {
		const inline = "syntax = \"proto3\"; package events; // Order event\nmessage Order {}"
		const leading = "syntax = \"proto3\"; package events;\n// Order event\nmessage Order {}"
		r := &versionRegistry{render: func(string) string { return leading }}
		attrs := `schema_type = "PROTOBUF"`
		keyed := registrationConfig(leading, attrs) + `
resource "terraform_data" "by_version" {
  for_each = toset([tostring(aiven_kafka_schema.orders.version)])
  input    = each.key
}
`
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: registrationConfig(inline, attrs)},
			resource.TestStep{Config: keyed, PlanOnly: true, ExpectNonEmptyPlan: true},
			resource.TestStep{Config: registrationConfig(leading, attrs), Check: resource.TestCheckResourceAttr(schemaAddress, "schema", leading)},
		))
		require.Equal(t, []string{inline}, r.submissions)
	})
	t.Run("move_to_missing_service", func(t *testing.T) {
		r := &versionRegistry{}
		original := registrationConfig(registrySchema, "")
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: original},
			resource.TestStep{
				Config:   schemaConfig(registryProject, "missing-kafka", "orders", registrySchema, ""),
				PlanOnly: true, ExpectNonEmptyPlan: true, ExpectError: regexp.MustCompile("destination service"),
			},
			resource.TestStep{Config: original, PlanOnly: true},
		))
	})
	t.Run("move_after_incomplete_destination_version_list", func(t *testing.T) {
		const dest = "payments"
		original := registrationConfig(updatedSchema, "")
		r := &versionRegistry{}
		tc := r.testCase(t, resource.TestStep{Config: original}, resource.TestStep{
			Config:   schemaConfig(registryProject, registrationService, dest, updatedSchema, ""),
			PlanOnly: true, ExpectNonEmptyPlan: true, ExpectError: regexp.MustCompile(`subject "payments": schema is not compatible`),
		}, resource.TestStep{Config: original, PlanOnly: true})
		api := r.client.EXPECT()
		var destinationListed bool
		api.ServiceSchemaRegistrySubjectVersionsGet(mock.Anything, registryProject, registrationService, dest).
			RunAndReturn(func(context.Context, string, string, string) ([]int, error) {
				if !destinationListed {
					return nil, avngen.Error{Status: http.StatusNotFound}
				}
				return []int{1}, nil
			})
		api.ServiceSchemaRegistrySubjects(mock.Anything, registryProject, registrationService).
			RunAndReturn(func(context.Context, string, string) ([]string, error) {
				destinationListed = true
				return []string{"orders", dest}, nil
			}).Maybe()
		api.ServiceSchemaRegistryCompatibility(mock.Anything, registryProject, registrationService, dest, 1, mock.Anything).
			Return(&sr.ServiceSchemaRegistryCompatibilityOut{IsCompatible: false}, nil).Maybe()
		api.ServiceSchemaRegistrySubjectVersionGet(mock.Anything, registryProject, registrationService, dest, 1).
			Return(&sr.ServiceSchemaRegistrySubjectVersionGetOut{Schema: registrySchema, SchemaType: sr.SchemaTypeAvro}, nil).Maybe()
		resource.UnitTest(t, tc)
		require.Equal(t, []string{updatedSchema}, r.submissions)
	})
	t.Run("move_to_unresolved_subject", func(t *testing.T) {
		original := registrationConfig(registrySchema, "")
		proposed := strings.Replace(original, `subject_name = "orders"`, "subject_name = terraform_data.dest.output", 1) +
			`resource "terraform_data" "dest" { input = "destination" }`
		r := &versionRegistry{}
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: original},
			resource.TestStep{Config: proposed, PlanOnly: true, ExpectNonEmptyPlan: true, ExpectError: regexp.MustCompile("subject_name must be known before replacing")},
			resource.TestStep{Config: original, PlanOnly: true},
		))
	})
	t.Run("unknown_payload_on_replacement", func(t *testing.T) {
		r := &versionRegistry{}
		original := registrationConfig(registrySchema, `schema_type = "JSON"`)
		proposed := strings.Replace(registrationConfig("payload", `schema_type = "AVRO"`), `schema       = "payload"`, "schema       = terraform_data.payload.output", 1) +
			`resource "terraform_data" "payload" { input = "bad-avro" }`
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: original},
			resource.TestStep{Config: proposed, PlanOnly: true, ExpectNonEmptyPlan: true, ExpectError: regexp.MustCompile("known before replacing")},
			resource.TestStep{Config: original, PlanOnly: true},
		))
	})
	t.Run("unknown_type_on_replacement", func(t *testing.T) {
		const schema = `{"type":"object"}`
		original := registrationConfig(schema, `schema_type = "JSON"`)
		proposed := registrationConfig(schema, "schema_type = terraform_data.kind.output") + `resource "terraform_data" "kind" { input = "AVRO" }`
		r := &versionRegistry{}
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: original},
			resource.TestStep{Config: proposed, PlanOnly: true, ExpectNonEmptyPlan: true, ExpectError: regexp.MustCompile("invalid Avro object")},
			resource.TestStep{Config: original, PlanOnly: true},
		))
	})
	t.Run("replace_referenced_schema_with_unknown_payload", func(t *testing.T) {
		attrs := "schema_type = \"PROTOBUF\"\n" + refBlock("dep.proto", "dependency", 1)
		const source = `syntax = "proto3"; import "dep.proto"; message Order { Dep value = 1; }`
		original := registrationConfig(source, attrs)
		proposed := strings.Replace(schemaConfig(registryProject, registrationService, "new-orders", "payload", attrs),
			`schema       = "payload"`, "schema       = terraform_data.payload.output", 1) +
			`resource "terraform_data" "payload" { input = ` + strconv.Quote(source) + " }"
		r := &versionRegistry{}
		resource.UnitTest(t, r.testCase(t, resource.TestStep{Config: original},
			resource.TestStep{Config: proposed, PlanOnly: true, ExpectNonEmptyPlan: true, ConfigPlanChecks: resource.ConfigPlanChecks{
				PostApplyPreRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(schemaAddress, plancheck.ResourceActionDestroyBeforeCreate)},
			}}))
	})
	t.Run("unknown_references_on_replacement", func(t *testing.T) {
		const schema = `{"type":"object"}`
		original := registrationConfig(schema, `schema_type = "JSON"`)
		proposed := registrationConfig(schema, `schema_type = "AVRO"
  dynamic "references" {
    for_each = terraform_data.refs.output
    content {
      name    = "R"
      subject = "ref"
      version = 1
    }
  }`) + `resource "terraform_data" "refs" { input = [] }`
		r := &versionRegistry{}
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: original},
			resource.TestStep{Config: proposed, PlanOnly: true, ExpectNonEmptyPlan: true, ExpectError: regexp.MustCompile("schema validation error")},
			resource.TestStep{Config: original, PlanOnly: true},
		))
	})
	t.Run("serialized_descriptor_without_json_name", func(t *testing.T) {
		const source = `syntax = "proto3"; message M { string order_id = 1; }`
		r := &versionRegistry{render: func(string) string { return source }}
		config := registrationConfig(descriptorWithField, `schema_type = "PROTOBUF"`)
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: config, Check: resource.TestCheckResourceAttr(schemaAddress, "schema", descriptorWithField)},
			resource.TestStep{Config: config, PlanOnly: true},
		))
	})
	const mapSource = `syntax = "proto3"; package demo; message M { repeated .demo.M.LabelsEntry labels = 1; message LabelsEntry { option map_entry = true; string key = 1; string value = 2; } }`
	const commonProto = `syntax = "proto3"; package demo; message Common {}`
	t.Run("format_registry_map_text", func(t *testing.T) {
		r := &versionRegistry{render: strings.TrimSpace}
		attrs := `schema_type = "PROTOBUF"`
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: registrationConfig(mapSource, attrs)},
			resource.TestStep{Config: registrationConfig("\n"+mapSource, attrs), Check: resource.TestCheckResourceAttr(schemaAddress, "schema", "\n"+mapSource)},
		))
		require.Len(t, r.submissions, 1)
	})
	for _, tc := range []struct {
		name, submitted, observed, attrs string
		dependencies                     []sr.ServiceSchemaRegistrySubjectVersionGetOut
	}{
		{"aggregate_option_delimiters", protoWithMeta, strings.Replace(protoWithMeta, `"region" value:`, `"region", value:`, 1), "", nil},
		{"repeated_aggregate_option", protoWithRepeatedMeta, `syntax = "proto3"; import "confluent/meta.proto"; message Order { option (confluent.message_meta) = { params: [{ key: "region", value: "EU" }, { key: "team", value: "billing" }], doc: "Order" }; }`, "", nil},
		{"empty_rpc_options", protoWithRPC, strings.Replace(protoWithRPC, "returns (M) {}", "returns (M);", 1), "", nil},
		{"serialized_enum_options", descriptorWithEnum, `syntax = "proto3"; enum E { X = 0; }`, "", nil},
		{"serialized_optional", descriptorWithOptional, `syntax = "proto3"; message Order { optional string id = 1; }`, "", nil},
		{"serialized_map", descriptorWithMap, mapSource, "", nil},
		{"registry_map_as_source", "\n" + mapSource, mapSource, "", nil},
		{
			"registry_builtin_import", descriptorWithDecimal,
			`syntax = "proto3"; package demo; import "confluent/type/decimal.proto"; message Order { .confluent.type.Decimal value = 1; }`, "", nil,
		},
		{
			"shared_transitive_import", "Cgpyb290LnByb3RvEgRkZW1vGgdhLnByb3RvGgdiLnByb3RvIh8KBU9yZGVyEhYKBXZhbHVlGAEgASgLMgcuZGVtby5BYgZwcm90bzM=",
			`syntax = "proto3"; package demo; import "a.proto"; import "b.proto"; message Order { .demo.A value = 1; }`, refBlock("a.proto", "a", 1) + "\n" + refBlock("b.proto", "b", 1),
			[]sr.ServiceSchemaRegistrySubjectVersionGetOut{
				{Subject: "a", Version: 1, Schema: `syntax = "proto3"; package demo; import "common.proto"; message A { Common c = 1; }`, References: []sr.ReferenceOut{{Name: "common.proto", Subject: "common-a", Version: 1}}},
				{Subject: "b", Version: 1, Schema: `syntax = "proto3"; package demo; import "common.proto"; message B { Common c = 1; }`, References: []sr.ReferenceOut{{Name: "common.proto", Subject: "common-b", Version: 1}}},
				{Subject: "common-a", Version: 1, Schema: commonProto},
				{Subject: "common-b", Version: 1, Schema: commonProto},
			},
		},
		{
			"serialized_import", descriptorWithImport, `syntax = "proto3"; package demo; import "dep.proto"; message M { .demo.Dep d = 1; }`, refBlock("dep.proto", "dep-subject", 1),
			[]sr.ServiceSchemaRegistrySubjectVersionGetOut{{Subject: "dep-subject", Version: 1, SchemaType: sr.SchemaTypeProtobuf, Schema: `syntax = "proto3"; package demo; message Dep {}`}},
		},
		{"proto2_group_order", protoWithGroup, `syntax = "proto2"; message M { optional int32 n = 2; optional group G = 1 { optional int32 x = 1; } }`, "", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &versionRegistry{render: func(string) string { return tc.observed }, dependencies: tc.dependencies}
			config := registrationConfig(tc.submitted, `schema_type = "PROTOBUF"`+"\n"+tc.attrs)
			resource.UnitTest(t, r.testCase(t,
				resource.TestStep{Config: config, Check: resource.TestCheckResourceAttr(schemaAddress, "schema", tc.submitted)},
				resource.TestStep{Config: config, PlanOnly: true},
			))
		})
	}
	t.Run("compatibility_convergence_and_removal", func(t *testing.T) {
		r := &versionRegistry{readsAfterPut: []sr.CompatibilityType{"", sr.CompatibilityTypeBackward}}
		const timeouts = "\ntimeouts {\n create = \"30s\"\n update = \"30s\"\n}\n"
		config := registrationConfig(registrySchema, full+timeouts)
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: config, Check: resource.TestCheckResourceAttr(schemaAddress, "compatibility_level", "FULL")},
			resource.TestStep{Config: registrationConfig(registrySchema, none+timeouts), Check: resource.TestCheckResourceAttr(schemaAddress, "compatibility_level", "NONE")},
			resource.TestStep{
				PreConfig: func() { r.level = "" }, RefreshState: true, ExpectNonEmptyPlan: true,
				Check: resource.TestCheckNoResourceAttr(schemaAddress, "compatibility_level"),
			},
		))
	})
	t.Run("partial_update", func(t *testing.T) {
		r := &versionRegistry{deniedLevel: sr.CompatibilityTypeNone}
		ref := "\n" + refBlock("Ref", "ref", 1)
		tc := r.testCase(t,
			resource.TestStep{Config: registrationConfig(registrySchema, full)},
			resource.TestStep{Config: registrationConfig(updatedSchema, none+ref), ExpectError: regexp.MustCompile("compatibility write denied")},
			resource.TestStep{
				Config:           registrationConfig(updatedSchema, full+ref),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(schemaAddress, "version", "2"),
					resource.TestCheckResourceAttr(schemaAddress, "schema", updatedSchema),
					resource.TestCheckResourceAttr(schemaAddress, "references.#", "1"),
					resource.TestCheckResourceAttr(schemaAddress, "compatibility_level", "FULL"),
				),
			},
			resource.TestStep{
				Config: registrationConfig(updatedSchema, none+ref), PreConfig: func() { r.deniedLevel = "" },
				Check: resource.TestCheckResourceAttr(schemaAddress, "version", "2"),
			},
		)
		// The first plan after the 403 must use the state saved on error, before a
		// refresh could reconstruct the schema and references from the version.
		tc.AdditionalCLIOptions = &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}
		resource.UnitTest(t, tc)
	})
	t.Run("confirmed_compatibility_survives_read_error", func(t *testing.T) {
		r := &versionRegistry{failAfterLevel: sr.CompatibilityTypeBackward}
		initial := registrationConfig(registrySchema, full)
		tc := r.testCase(t,
			resource.TestStep{Config: initial},
			resource.TestStep{Config: registrationConfig(registrySchema, `compatibility_level = "BACKWARD"`), ExpectError: regexp.MustCompile("refresh after PUT failed")},
			resource.TestStep{Config: initial, PreConfig: func() { r.failAfterLevel = "" }, ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(schemaAddress, plancheck.ResourceActionUpdate)},
			}},
		)
		tc.AdditionalCLIOptions = &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}
		resource.UnitTest(t, tc)
		require.Equal(t, sr.CompatibilityTypeFull, r.level)
	})
	for _, tc := range []struct{ name, attrs, want string }{
		{"unmanaged_compatibility_at_apply", "", "BACKWARD"},
		{"configured_compatibility_at_apply", full, "FULL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &versionRegistry{level: sr.CompatibilityTypeFull}
			attrs := tc.attrs + "\ntimeouts {\n update = \"20s\"\n}\n"
			resource.UnitTest(t, r.testCase(t,
				resource.TestStep{Config: registrationConfig(registrySchema, attrs), Check: resource.TestCheckResourceAttr(schemaAddress, "compatibility_level", "FULL")},
				resource.TestStep{
					Config: registrationConfig(updatedSchema, attrs),
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						afterRegistrationPlan(func() { r.level = sr.CompatibilityTypeBackward }),
					}},
					Check: resource.TestCheckResourceAttr(schemaAddress, "compatibility_level", tc.want),
				},
			))
		})
	}
	t.Run("rollback_to_sdk_version", func(t *testing.T) {
		r := &versionRegistry{}
		// SDKv2's schema.Set.List() order for these three references.
		r.register(registrySchema, sr.SchemaTypeAvro, []sr.ReferenceIn{
			{Name: "A", Subject: "ref-A", Version: 1}, {Name: "C", Subject: "ref-C", Version: 1}, {Name: "B", Subject: "ref-B", Version: 1},
		})
		refs := "\n" + refBlock("B", "ref-B", 1) + "\n" + refBlock("A", "ref-A", 1) + "\n" + refBlock("C", "ref-C", 1)
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: registrationConfig(updatedSchema, refs), Check: resource.TestCheckResourceAttr(schemaAddress, "version", "2")},
			resource.TestStep{Config: registrationConfig(registrySchema, refs), Check: resource.TestCheckResourceAttr(schemaAddress, "version", "1")},
		))
	})
	t.Run("omitted_ignored_compatibility_at_apply", func(t *testing.T) {
		r := &versionRegistry{}
		config := registrationConfig(updatedSchema, `lifecycle { ignore_changes = [compatibility_level] }`)
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: registrationConfig(registrySchema, full)},
			resource.TestStep{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				afterRegistrationPlan(func() { r.level = sr.CompatibilityTypeBackward }),
			}}, Check: resource.TestCheckResourceAttr(schemaAddress, "version", "2")},
			resource.TestStep{RefreshState: true, Check: resource.TestCheckResourceAttr(schemaAddress, "compatibility_level", "BACKWARD")},
		))
		require.Equal(t, []sr.CompatibilityType{sr.CompatibilityTypeFull}, r.compatibilityWrites)
	})
	for _, level := range []sr.CompatibilityType{"", sr.CompatibilityTypeFull} {
		t.Run("ignored_compatibility_"+string(level), func(t *testing.T) {
			r := &versionRegistry{level: level}
			attrs := ""
			if level != "" {
				attrs = full
			}
			ignored := none + "\nlifecycle { ignore_changes = [compatibility_level] }\ntimeouts { update = \"2s\" }"
			resource.UnitTest(t, r.testCase(t,
				resource.TestStep{Config: registrationConfig(registrySchema, attrs)},
				resource.TestStep{Config: registrationConfig(updatedSchema, ignored), Check: resource.TestCheckResourceAttr(schemaAddress, "version", "2")},
			))
			require.Equal(t, level, r.level)
			require.NotContains(t, r.compatibilityWrites, sr.CompatibilityTypeNone)
		})
	}
	t.Run("retry_upsert_preserves_other_versions", func(t *testing.T) {
		r := &versionRegistry{deniedLevel: sr.CompatibilityTypeFull, subject: "orders"}
		r.register(updatedSchema, sr.SchemaTypeAvro, nil)
		r.register(registrySchema, sr.SchemaTypeAvro, nil)
		config := registrationConfig(registrySchema, full)
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: config, ExpectError: regexp.MustCompile("compatibility write denied")},
			resource.TestStep{Config: config, PreConfig: func() { r.deniedLevel = "" }, Check: resource.TestCheckResourceAttr(schemaAddress, "version", "2")},
		))
	})
	t.Run("registry_dropped_enum_reservation", func(t *testing.T) {
		const source = `syntax = "proto3"; enum Kind { reserved 5; EMPTY = 0; }`
		r := &versionRegistry{render: func(s string) string { return strings.Replace(s, "reserved 5; ", "", 1) }}
		config := registrationConfig(source, `schema_type = "PROTOBUF"`)
		// Retrying a rejected registration must not destroy the existing subject.
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: config, ExpectError: regexp.MustCompile("registered PROTOBUF schema differs")},
			resource.TestStep{Config: config, PlanOnly: true, ExpectNonEmptyPlan: true, ConfigPlanChecks: resource.ConfigPlanChecks{
				PostApplyPreRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(schemaAddress, plancheck.ResourceActionCreate)},
			}},
		))
	})
	t.Run("large_integer_bound", func(t *testing.T) {
		r := &versionRegistry{}
		const before = `{"type":"integer","maximum":9007199254740992}`
		const after = `{"type":"integer","maximum":9007199254740993}`
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: registrationConfig(before, `schema_type = "JSON"`)},
			resource.TestStep{Config: registrationConfig(after, `schema_type = "JSON"`), Check: resource.TestCheckResourceAttr(schemaAddress, "version", "2")},
		))
		require.Equal(t, []string{before, after}, r.submissions)
	})
	t.Run("transient_version_not_found", func(t *testing.T) {
		r := &versionRegistry{}
		config := registrationConfig(registrySchema, "timeouts { read = \"2s\" }")
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: config},
			resource.TestStep{Config: config, PlanOnly: true, PreConfig: func() { r.omitNextVersion = true }},
		))
	})
	t.Run("transient_parent_not_found", func(t *testing.T) {
		r := &versionRegistry{}
		config := registrationConfig(registrySchema, "")
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: config},
			resource.TestStep{Config: config, PlanOnly: true, PreConfig: func() { r.omitNextService = true }},
		))
	})
	t.Run("destroy_after_transient_parent_not_found", func(t *testing.T) {
		r := &versionRegistry{}
		config := registrationConfig(registrySchema, "")
		tc := r.testCase(t,
			resource.TestStep{Config: config},
			resource.TestStep{Config: config, Destroy: true, PreConfig: func() { r.omitNextService = true }},
		)
		tc.AdditionalCLIOptions = &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}
		resource.UnitTest(t, tc)
		require.Empty(t, r.versions)
	})
	t.Run("transient_parent_and_versions_not_found", func(t *testing.T) {
		r := &versionRegistry{}
		config := registrationConfig(registrySchema, "")
		tc := r.testCase(t,
			resource.TestStep{Config: config},
			resource.TestStep{Config: config, Destroy: true, PreConfig: func() { r.omitNextService, r.omitNextVersions = true, true }},
		)
		tc.AdditionalCLIOptions = &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}
		resource.UnitTest(t, tc)
		require.Empty(t, r.versions)
	})
	t.Run("disabled_registry_on_powered_off_parent", func(t *testing.T) {
		r := &versionRegistry{}
		config := registrationConfig(registrySchema, "")
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: config},
			resource.TestStep{
				Config: config, PlanOnly: true, ExpectNonEmptyPlan: true,
				PreConfig: func() { r.disabled, r.poweredOff = true, true },
				ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(schemaAddress, plancheck.ResourceActionCreate),
				}},
			},
		))
	})
	t.Run("registry_enabled_after_readiness_check", func(t *testing.T) {
		r := &versionRegistry{disabled: true}
		tc := r.testCase(t, resource.TestStep{
			Config: registrationConfig(registrySchema, "timeouts { create = \"2s\" }"),
			Check:  resource.TestCheckResourceAttr(schemaAddress, "version", "1"),
		})
		client := avngen.NewMockClient(t)
		client.EXPECT().ServiceGet(mock.Anything, registryProject, registrationService).
			Return(&service.ServiceGetOut{ServiceType: "kafka", UserConfig: map[string]any{"schema_registry": false}}, nil).Once()
		require.NoError(t, schemautil.CheckServiceReady(t.Context(), client, registryProject, registrationService, "kafka"))
		r.disabled = false
		resource.UnitTest(t, tc)
	})
}

func TestKafkaSchemaEscapedID(t *testing.T) {
	t.Run("preserve_import_spelling", func(t *testing.T) {
		r := &versionRegistry{subject: "orders/v1"}
		config := schemaConfig(registryProject, registrationService, r.subject, registrySchema, "")
		id := registryProject + "/" + registrationService + "/orders/v1"
		resource.UnitTest(t, r.testCase(t,
			resource.TestStep{Config: config},
			resource.TestStep{ResourceName: schemaAddress, ImportState: true, ImportStateId: id, ImportStateCheck: func(states []*terraform.InstanceState) error {
				require.Len(t, states, 1)
				require.Equal(t, id, states[0].ID)
				return nil
			}},
		))
	})
	for _, tc := range []struct{ subject, escaped string }{
		{"orders/v1", "orders%2Fv1"},
		{"orders%2Fv1", "orders%252Fv1"},
	} {
		t.Run(tc.escaped, func(t *testing.T) {
			r := &versionRegistry{subject: tc.subject}
			id := registryProject + "/" + registrationService + "/" + tc.escaped
			config := schemaConfig(registryProject, registrationService, tc.subject, registrySchema, "") + `
data "aiven_kafka_schema" "orders" {
  project      = aiven_kafka_schema.orders.project
  service_name = aiven_kafka_schema.orders.service_name
  subject_name = aiven_kafka_schema.orders.subject_name
}
`
			resource.UnitTest(t, r.testCase(t,
				resource.TestStep{Config: config, Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(schemaAddress, "id", id),
					resource.TestCheckResourceAttr("data.aiven_kafka_schema.orders", "id", id),
				)},
				resource.TestStep{ResourceName: schemaAddress, ImportState: true, ImportStateVerify: true},
			))
		})
	}
}
