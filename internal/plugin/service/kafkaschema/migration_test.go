package kafkaschema_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	sr "github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const schemaAddress = "aiven_kafka_schema.orders"

// FileDescriptorProto for 'syntax = "proto3"; message M { string order_id = 1; }'.
// Its field omits the optional json_name, so its effective JSON name is "orderId".
const descriptorWithField = "CgdtLnByb3RvIhUKAU0SEAoIb3JkZXJfaWQYASABKAliBnByb3RvMw=="

// Karapace's serialized Order.id has proto3_optional without a synthetic oneof.
const descriptorWithOptional = "CgdtLnByb3RvIhYKBU9yZGVyEg0KAmlkGAEgASgJiAEBYgZwcm90bzM="

const sdkProviderConfig = `
terraform {
  required_providers {
    aiven = { source = "aiven/aiven", version = "4.66.0" }
  }
}
`

// Compiled descriptors for demo.M with a map or a field of imported type demo.Dep.
const (
	descriptorWithMap     = "CgdtLnByb3RvEgRkZW1vImsKAU0SKwoGbGFiZWxzGAEgAygLMhMuZGVtby5NLkxhYmVsc0VudHJ5UgZsYWJlbHMaOQoLTGFiZWxzRW50cnkSEAoDa2V5GAEgASgJUgNrZXkSFAoFdmFsdWUYAiABKAlSBXZhbHVlOgI4AWIGcHJvdG8z"
	descriptorWithImport  = "CgdtLnByb3RvEgRkZW1vGglkZXAucHJvdG8iHAoBTRIXCgFkGAEgASgLMgkuZGVtby5EZXBSAWRiBnByb3RvMw=="
	descriptorWithDecimal = "Cgpyb290LnByb3RvEgRkZW1vGhxjb25mbHVlbnQvdHlwZS9kZWNpbWFsLnByb3RvIi8KBU9yZGVyEiYKBXZhbHVlGAEgASgLMhcuY29uZmx1ZW50LnR5cGUuRGVjaW1hbGIGcHJvdG8z"
	// Karapace sets deprecated=false on its serialized enum values.
	descriptorWithEnum    = "CgdtLnByb3RvKg4KAUUSCQoBWBAAGgIIAGIGcHJvdG8z"
	protoWithMeta         = `syntax = "proto3"; import "confluent/meta.proto"; message Order { option (confluent.message_meta) = { params: { key: "region" value: "EU" } }; }`
	protoWithRepeatedMeta = `syntax = "proto3"; import "confluent/meta.proto"; message Order { option (confluent.message_meta) = { params: { key: "region" value: "EU" } doc: "Order" params: { key: "team" value: "billing" } }; }`
	protoWithRPC          = `syntax = "proto3"; message M {} service S { rpc Get(M) returns (M) {} }`
	protoWithGroup        = `syntax = "proto2"; message M { optional group G = 1 { optional int32 x = 1; } optional int32 n = 2; }`
)

func schemaConfig(project, service, subject, schema, attributes string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema" "orders" {
  project      = %q
  service_name = %q
  subject_name = %q
  schema       = %q
  %s
}
`, project, service, subject, schema, attributes)
}

func testSchemaMigration(t *testing.T, project, service string, ready <-chan error) {
	const avro = `{"type":"string"}`
	preCheck := func(t *testing.T) func() {
		return func() { acc.TestAccPreCheck(t); require.NoError(t, <-ready) }
	}
	test := func(t *testing.T, steps ...resource.TestStep) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck: preCheck(t), ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories, Steps: steps,
		})
	}
	t.Run("registry_protobuf", func(t *testing.T) {
		// Karapace prints a double-quoted syntax, messages before enums, and fields
		// before oneofs. The original source also contains a header and inline comment.
		const source = "// Order event\nsyntax = 'proto3';\nenum Status { UNKNOWN = 0; }\nmessage Order {\n  oneof key { string code = 2; }\n  Status status = 1; // State\n}"
		config := schemaConfig(project, service, acc.RandName("format"), source, `schema_type = "PROTOBUF"`)
		test(t, resource.TestStep{Config: config, Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(schemaAddress, "schema", source), resource.TestCheckResourceAttr(schemaAddress, "version", "1"),
		)}, resource.TestStep{Config: config, PlanOnly: true})
	})
	t.Run("serialized_protobuf", func(t *testing.T) {
		config := schemaConfig(project, service, acc.RandName("descriptor"), descriptorWithField, `schema_type = "PROTOBUF"`)
		test(t, resource.TestStep{Config: config, Check: resource.TestCheckResourceAttr(schemaAddress, "schema", descriptorWithField)},
			resource.TestStep{Config: config, PlanOnly: true})
	})
	t.Run("protobuf_options", func(t *testing.T) {
		var config string
		var checks []resource.TestCheckFunc
		for i, source := range []string{protoWithMeta, protoWithRepeatedMeta, protoWithRPC, descriptorWithEnum, descriptorWithOptional} {
			name := fmt.Sprintf("options%d", i)
			config += fmt.Sprintf(`
resource "aiven_kafka_schema" %q {
  project = %q
  service_name = %q
  subject_name = %q
  schema_type = "PROTOBUF"
  schema = %q
}
`, name, project, service, acc.RandName("options"), source)
			checks = append(checks, resource.TestCheckResourceAttr("aiven_kafka_schema."+name, "schema", source))
		}
		test(t, resource.TestStep{Config: config, Check: resource.ComposeTestCheckFunc(checks...)}, resource.TestStep{Config: config, PlanOnly: true})
	})
	t.Run("registry_protobuf_inputs", func(t *testing.T) {
		config := schemaConfig(project, service, acc.RandName("decimal"), descriptorWithDecimal, `schema_type = "PROTOBUF"`)
		config += fmt.Sprintf(`
resource "aiven_kafka_schema" "original" {
  project      = %q
  service_name = %q
  subject_name = %q
  schema_type  = "PROTOBUF"
  schema       = %q
}

data "aiven_kafka_schema" "original" {
  project      = aiven_kafka_schema.original.project
  service_name = aiven_kafka_schema.original.service_name
  subject_name = aiven_kafka_schema.original.subject_name
}

resource "aiven_kafka_schema" "copy" {
  project      = %q
  service_name = %q
  subject_name = %q
  schema_type  = "PROTOBUF"
  schema       = format("\n%%s", data.aiven_kafka_schema.original.schema)
}
`, project, service, acc.RandName("map"), descriptorWithMap, project, service, acc.RandName("map-copy"))
		test(t, resource.TestStep{Config: config, Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(schemaAddress, "schema", descriptorWithDecimal),
			resource.TestCheckResourceAttr("aiven_kafka_schema.copy", "version", "1"),
		)}, resource.TestStep{Config: config, PlanOnly: true})
	})
	t.Run("protected_replacement", func(t *testing.T) {
		original, destination := acc.RandName("original"), acc.RandName("destination")
		const policy = `compatibility_level = "BACKWARD"`
		dest := fmt.Sprintf(`
resource "aiven_kafka_schema" "destination" {
  project             = %q
  service_name        = %q
  subject_name        = %q
  schema              = jsonencode({ type = "int" })
  compatibility_level = "BACKWARD"
}
`, project, service, destination)
		config := func(subject, attrs string) string { return dest + schemaConfig(project, service, subject, avro, attrs) }
		rejected := regexp.MustCompile("schema is not compatible")
		test(t,
			resource.TestStep{Config: config(original, policy)},
			resource.TestStep{Config: config(original, policy+"\nschema_type = \"JSON\""), PlanOnly: true, ExpectNonEmptyPlan: true, ExpectError: rejected},
			resource.TestStep{Config: config(destination, policy), PlanOnly: true, ExpectNonEmptyPlan: true, ExpectError: rejected},
			resource.TestStep{Config: config(original, policy), PlanOnly: true},
		)
	})
	t.Run("planned_schema_type", func(t *testing.T) {
		subject := acc.RandName("planned")
		known := schemaConfig(project, service, subject, avro, "") + `
resource "terraform_data" "by_kind" {
  for_each = toset([aiven_kafka_schema.orders.schema_type])
  input    = each.key
}
`
		unknown := schemaConfig(project, service, subject, `syntax = "proto3"; message M {}`, `schema_type = terraform_data.kind.output`) + `
resource "terraform_data" "kind" { input = "PROTOBUF" }
`
		test(t, resource.TestStep{Config: known, PlanOnly: true, ExpectNonEmptyPlan: true},
			resource.TestStep{Config: unknown, Check: resource.TestCheckResourceAttr(schemaAddress, "schema_type", "PROTOBUF")})
	})
	t.Run("planned_reference_set", func(t *testing.T) {
		subject := acc.RandName("orders")
		ref := acc.RandName("ref")
		reference := fmt.Sprintf(`
resource "aiven_kafka_schema" "ref" {
  project      = %q
  service_name = %q
  subject_name = %q
  schema       = jsonencode({ type = "record", name = "Ref", fields = [] })
}
`, project, service, ref)
		const source = `{"type":"record","name":"Order","fields":[]}`
		config := func(attrs string) string { return reference + schemaConfig(project, service, subject, source, attrs) }
		updated := config(`
  dynamic "references" {
    for_each = toset(terraform_data.names.output)
    content {
      name    = "Ref"
      subject = references.value
      version = aiven_kafka_schema.ref.version
    }
  }`) + `
resource "terraform_data" "names" { input = [aiven_kafka_schema.ref.subject_name] }
`
		test(t, resource.TestStep{Config: config(""), Check: resource.TestCheckResourceAttr(schemaAddress, "version", "1")},
			resource.TestStep{Config: updated, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectUnknownValue(schemaAddress, tfjsonpath.New("version")),
			}}, Check: resource.TestCheckResourceAttr(schemaAddress, "version", "2")})
	})
	t.Run("formatting_without_compatibility_override", func(t *testing.T) {
		subject := acc.RandName("format-level")
		initial := schemaConfig(project, service, subject, avro, "")
		formatted := schemaConfig(project, service, subject, `{ "type" : "string" }`, "")
		keyed := formatted + `
resource "terraform_data" "by_level" {
  for_each = toset([coalesce(aiven_kafka_schema.orders.compatibility_level, "default")])
  input    = each.key
}
`
		test(t, resource.TestStep{Config: initial}, resource.TestStep{
			Config: keyed, PlanOnly: true, ExpectNonEmptyPlan: true,
			ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(`terraform_data.by_level["default"]`, plancheck.ResourceActionCreate),
			}},
		}, resource.TestStep{Config: formatted, Check: resource.TestCheckResourceAttr(schemaAddress, "version", "1")})
	})
	t.Run("replacement_validation", func(t *testing.T) {
		subject := acc.RandName("validate")
		original := schemaConfig(project, service, subject, avro, "")
		steps := []resource.TestStep{{Config: original, Check: resource.TestCheckResourceAttr(schemaAddress, "version", "1")}}
		for _, tc := range []struct{ project, service, subject, schema, attrs, diagnostic string }{
			{project, acc.RandName("absent"), subject + "-new", avro, `lifecycle { ignore_changes = [service_name] }`, "destination service"},
			{project + "!", service, subject, avro, "", "project must match pattern"},
			{project, service + " ", subject, avro, "", "service_name must match pattern"},
			{project, service, subject + "-new", "", "", "schema validation"},
			{project, service, subject, `{"type":"strng"}`, `schema_type = "JSON"`, "schema validity"},
			{project, service, subject + "-new", avro, refBlock("", subject, 1), "length must be"},
			{project, service, subject + "-new", avro, refBlock("Ref", "", 1), "length must be"},
			{project, service, subject + "-new", avro, refBlock("Ref", subject, -1), "at least 0"},
		} {
			steps = append(steps, resource.TestStep{
				Config: schemaConfig(tc.project, tc.service, tc.subject, tc.schema, tc.attrs), PlanOnly: true, ExpectNonEmptyPlan: true, ExpectError: regexp.MustCompile(tc.diagnostic),
			})
		}
		test(t, append(steps, resource.TestStep{Config: original, PlanOnly: true})...)
	})
	for _, tc := range []struct {
		name, original, newer string
		kind                  sr.SchemaType
		fromSDK               bool
	}{
		{
			"move_to_existing_version", `{"type":"record","name":"MoveEvent","fields":[{"name":"quantity","type":"int"}]}`,
			`{"type":"record","name":"MoveEvent","fields":[{"name":"quantity","type":"long"}]}`, sr.SchemaTypeAvro, false,
		},
		{
			"move_from_sdk_to_registered_protobuf", `syntax = "proto3"; enum Status { READY = 0; } message MoveEvent { Status state = 1; }`,
			`syntax = "proto3"; message MoveEvent { string state = 1; }`, sr.SchemaTypeProtobuf, true,
		},
		{
			"move_from_sdk_to_registered_descriptor", descriptorWithField,
			`syntax = "proto3"; message M { int32 order_id = 1; }`, sr.SchemaTypeProtobuf, true,
		},
		{
			"move_from_sdk_to_registered_json", `{"type":"number","maximum":1.50}`,
			// A distinct schema avoids sharing a registry ID with the Avro string schema.
			`{"type":"string","enum":["decimal-history-v2"]}`, sr.SchemaTypeJson, true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source, dest := acc.RandName("source"), acc.RandName("dest")
			config := func(subject string) string {
				return schemaConfig(project, service, subject, tc.original, fmt.Sprintf("schema_type = %q", tc.kind))
			}
			create := resource.TestStep{Config: config(source), ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories}
			if tc.fromSDK {
				t.Setenv(resource.EnvTfAccProviderNamespace, "aiven")
				create = resource.TestStep{
					Config:             sdkProviderConfig + config(source),
					ExpectNonEmptyPlan: tc.kind == sr.SchemaTypeProtobuf, // SDK keeps the registry's rendered Protobuf text.
					ExternalProviders:  map[string]resource.ExternalProvider{"aiven": acc.ExternalAivenProvider(t, "4.66.0")},
				}
			}
			move := resource.TestStep{
				Config: config(dest), ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
				PreConfig: func() {
					c, err := acc.GetTestGenAivenClient()
					require.NoError(t, err)
					t.Cleanup(func() {
						if err := c.ServiceSchemaRegistrySubjectDelete(context.Background(), project, service, dest); !avngen.IsNotFound(err) {
							require.NoError(t, err)
						}
					})
					_, err = c.ServiceSchemaRegistrySubjectVersionPost(t.Context(), project, service, dest, &sr.ServiceSchemaRegistrySubjectVersionPostIn{Schema: tc.original, SchemaType: tc.kind})
					require.NoError(t, err)
					_, err = c.ServiceSchemaRegistrySubjectConfigPut(t.Context(), project, service, dest, &sr.ServiceSchemaRegistrySubjectConfigPutIn{Compatibility: sr.CompatibilityTypeNone})
					require.NoError(t, err)
					_, err = c.ServiceSchemaRegistrySubjectVersionPost(t.Context(), project, service, dest, &sr.ServiceSchemaRegistrySubjectVersionPostIn{Schema: tc.newer, SchemaType: tc.kind})
					require.NoError(t, err)
					_, err = c.ServiceSchemaRegistrySubjectConfigPut(t.Context(), project, service, dest, &sr.ServiceSchemaRegistrySubjectConfigPutIn{Compatibility: sr.CompatibilityTypeBackward})
					require.NoError(t, err)
				},
				Check: resource.TestCheckResourceAttr(schemaAddress, "version", "1"),
			}
			resource.Test(t, resource.TestCase{PreCheck: preCheck(t), Steps: []resource.TestStep{create, move}})
		})
	}
	t.Run("sdk_import_without_refresh", func(t *testing.T) {
		t.Setenv(resource.EnvTfAccProviderNamespace, "aiven")
		subject := acc.RandName("import-json")
		schema := fmt.Sprintf(`{"type":"string","description":%q}`, subject)
		config := func(attrs string) string { return schemaConfig(project, service, subject, schema, attrs) }
		resource.Test(t, resource.TestCase{
			PreCheck: preCheck(t), AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
			Steps: []resource.TestStep{
				{
					Config: sdkProviderConfig + config(""), ExternalProviders: map[string]resource.ExternalProvider{"aiven": acc.ExternalAivenProvider(t, "4.66.0")},
					ResourceName: schemaAddress, ImportState: true, ImportStatePersist: true, ImportStateId: schemautil.BuildResourceID(project, service, subject),
					PreConfig: func() {
						c, err := acc.GetTestGenAivenClient()
						require.NoError(t, err)
						_, err = c.ServiceSchemaRegistrySubjectVersionPost(t.Context(), project, service, subject, &sr.ServiceSchemaRegistrySubjectVersionPostIn{Schema: schema, SchemaType: sr.SchemaTypeJson})
						require.NoError(t, err)
						// Permit a type change here so this test isolates SDK import planning.
						_, err = c.ServiceSchemaRegistrySubjectConfigPut(t.Context(), project, service, subject, &sr.ServiceSchemaRegistrySubjectConfigPutIn{Compatibility: sr.CompatibilityTypeNone})
						require.NoError(t, err)
						t.Cleanup(func() {
							if err := c.ServiceSchemaRegistrySubjectDelete(context.Background(), project, service, subject); !avngen.IsNotFound(err) {
								require.NoError(t, err)
							}
						})
					}, ImportStateCheck: func(items []*terraform.InstanceState) error {
						if len(items) != 1 || items[0].Attributes["schema_type"] != "" || items[0].Attributes["version"] != "1" {
							return fmt.Errorf("expected an SDK import with version 1 and no saved type")
						}
						return nil
					},
				},
				{
					Config: config(`schema_type = "AVRO"`), ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories, PlanOnly: true, ExpectNonEmptyPlan: true,
					ExpectError: regexp.MustCompile("refresh-only"),
				},
				{
					Config: config(`compatibility_level = "FULL"`), ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
					Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(schemaAddress, "schema_type", "JSON"),
						resource.TestCheckResourceAttr(schemaAddress, "version", "1"), resource.TestCheckResourceAttr(schemaAddress, "compatibility_level", "FULL")),
				},
			},
		})
	})
	t.Run("move_deleted_sdk_import", func(t *testing.T) {
		t.Setenv(resource.EnvTfAccProviderNamespace, "aiven")
		source, dest, dependency := acc.RandName("old"), acc.RandName("new"), acc.RandName("dep")
		const sourceSchema = `syntax = "proto3"; import "dep.proto"; message Order { Dep payload = 1; }`
		config := func(subject string) string {
			return schemaConfig(project, service, subject, sourceSchema, `schema_type = "PROTOBUF"`+"\n"+refBlock("dep.proto", dependency, 1))
		}
		resource.Test(t, resource.TestCase{
			PreCheck: preCheck(t), AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
			Steps: []resource.TestStep{
				{
					Config: sdkProviderConfig + config(source), ExternalProviders: map[string]resource.ExternalProvider{"aiven": acc.ExternalAivenProvider(t, "4.66.0")},
					ResourceName: schemaAddress, ImportState: true, ImportStatePersist: true, ImportStateId: schemautil.BuildResourceID(project, service, source),
					PreConfig: func() {
						c, err := acc.GetTestGenAivenClient()
						require.NoError(t, err)
						t.Cleanup(func() {
							for _, subject := range []string{source, dest, dependency} {
								for _, permanent := range []bool{false, true} {
									err := c.ServiceSchemaRegistrySubjectDelete(context.Background(), project, service, subject, sr.ServiceSchemaRegistrySubjectDeletePermanent(permanent))
									if !avngen.IsNotFound(err) {
										require.NoError(t, err)
									}
								}
							}
						})
						_, err = c.ServiceSchemaRegistrySubjectVersionPost(t.Context(), project, service, dependency, &sr.ServiceSchemaRegistrySubjectVersionPostIn{Schema: `syntax = "proto3"; message Dep {}`, SchemaType: sr.SchemaTypeProtobuf})
						require.NoError(t, err)
						refs := []sr.ReferenceIn{{Name: "dep.proto", Subject: dependency, Version: 1}}
						_, err = c.ServiceSchemaRegistrySubjectVersionPost(t.Context(), project, service, source, &sr.ServiceSchemaRegistrySubjectVersionPostIn{Schema: sourceSchema, SchemaType: sr.SchemaTypeProtobuf, References: &refs})
						require.NoError(t, err)
					},
				},
				{
					Config: config(dest), ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
					PreConfig: func() {
						c, err := acc.GetTestGenAivenClient()
						require.NoError(t, err)
						require.NoError(t, c.ServiceSchemaRegistrySubjectDelete(t.Context(), project, service, source))
					},
					Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(schemaAddress, "subject_name", dest), resource.TestCheckResourceAttr(schemaAddress, "version", "1")),
				},
			},
		})
	})
}

func refBlock(name, subject string, version int) string {
	return fmt.Sprintf("references {\n name = %q\n subject = %q\n version = %d\n}", name, subject, version)
}
