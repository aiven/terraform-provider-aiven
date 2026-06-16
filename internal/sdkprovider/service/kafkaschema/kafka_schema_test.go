package kafkaschema_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// TestAccAivenKafkaSchema_schema_registry_lifecycle tests the complete lifecycle of managing
// Kafka schemas when Schema Registry is enabled and disabled.
func TestAccAivenKafkaSchema_schema_registry_lifecycle(t *testing.T) {
	resourceName := "aiven_kafka_schema.foo"
	sName := acc.RandName("kafka-schema")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories:  acc.TestProtoV6ProviderFactories,
		PreventPostDestroyRefresh: true,
		CheckDestroy:              testAccCheckAivenKafkaSchemaResourceDestroy,
		Steps: []resource.TestStep{
			{
				// create schema with schema registry enabled
				Config: testAccKafkaSchemaWithSchemaRegistryConfig(sName, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_name", sName),
					resource.TestCheckResourceAttr(resourceName, "subject_name", "kafka-schema-lifecycle"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				// disable registry while keeping schema in config
				// provider should handle 403 error and remove schema from state
				Config:             testAccKafkaSchemaWithSchemaRegistryConfig(sName, false, true),
				ExpectNonEmptyPlan: true, // schema removed from state, plan will want to recreate it
			},
			{
				// re-enable registry, keep schema in config, schema should be recreated
				Config: testAccKafkaSchemaWithSchemaRegistryConfig(sName, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_name", sName),
					resource.TestCheckResourceAttr(resourceName, "subject_name", "kafka-schema-lifecycle"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				// disable registry and remove schema resource simultaneously, provider should handle 403 error during deletion
				Config: testAccKafkaSchemaWithSchemaRegistryConfig(sName, false, false),
			},
			{
				// re-enable registry without schema resource
				// ensures service is in clean state after delete test
				Config: testAccKafkaSchemaWithSchemaRegistryConfig(sName, true, false),
			},
			{
				// add schema resource back
				// Note: Version is "2" because Schema Registry maintains version history;
				// the schema deleted in Step 4 was version 1, so recreating it assigns version 2
				Config: testAccKafkaSchemaWithSchemaRegistryConfig(sName, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_name", sName),
					resource.TestCheckResourceAttr(resourceName, "subject_name", "kafka-schema-lifecycle"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
				),
			},
		},
	})
}

func TestAccAivenKafkaSchema_avroReferencesPlan(t *testing.T) {
	refSubject := acc.RandName("ref")
	depSubject := acc.RandName("dep")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccKafkaSchemaAVROReferencesPlan(refSubject, depSubject),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAivenKafkaSchema(t *testing.T) {
	projectName := acc.ProjectName()
	serviceName := acc.RandName("kafka")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("kafka"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
		acc.WithUserConfig(map[string]any{"schema_registry": true}),
	)

	t.Run("basic", func(t *testing.T) {
		resourceName := "aiven_kafka_schema.foo"
		subjectName := acc.RandName("basic")

		// This test can't run in Parallel
		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				acc.TestAccPreCheck(t)
				require.NoError(t, <-serviceIsReady)
			},
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaSchemaResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaSchemaResource(projectName, serviceName, subjectName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema"),
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "subject_name", subjectName),
						resource.TestCheckResourceAttr(resourceName, "version", "1"),
						resource.TestCheckResourceAttr(resourceName, "schema_type", "AVRO"),
					),
				},
				{
					Config: testAccKafkaSchemaResourceGoodUpdate(projectName, serviceName, subjectName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema"),
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "subject_name", subjectName),
						resource.TestCheckResourceAttr(resourceName, "version", "2"),
						resource.TestCheckResourceAttr(resourceName, "schema_type", "AVRO"),
					),
				},
				{
					Config: testAccKafkaSchemaResource(projectName, serviceName, subjectName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema"),
						resource.TestCheckResourceAttr(resourceName, "version", "1"),
						resource.TestCheckResourceAttr(resourceName, "schema_type", "AVRO"),
					),
				},
				{
					Config:      testAccKafkaSchemaResourceInvalidUpdate(projectName, serviceName, subjectName),
					ExpectError: regexp.MustCompile("schema is not compatible with previous version"),
				},
			},
		})
	})

	// checks that compatibility_level doesn't appear in plan after KafkaSchema import
	t.Run("import_compatibility_level", func(t *testing.T) {
		resourceName := "aiven_kafka_schema.schema"
		subjectName := acc.RandName("import")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() {
				acc.TestAccPreCheck(t)
				require.NoError(t, <-serviceIsReady)
			},
			ProtoV6ProviderFactories:  acc.TestProtoV6ProviderFactories,
			CheckDestroy:              testAccCheckAivenKafkaSchemaResourceDestroy,
			PreventPostDestroyRefresh: true,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaSchemaImportCompatibilityLevel(projectName, serviceName, subjectName),
				},
				{
					ResourceName:       resourceName,
					ExpectNonEmptyPlan: false, // compatibility_level doesn't appear in plan
					ImportState:        true,
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return "", fmt.Errorf("expected resource %q to be present in the state", resourceName)
						}

						v := rs.Primary.Attributes["compatibility_level"]
						if v != "FULL_TRANSITIVE" {
							return "", fmt.Errorf(`expected resource %q to have compatibility_level = "FULL_TRANSITIVE", got %q`, resourceName, v)
						}
						return rs.Primary.ID, nil
					},
				},
			},
		})
	})

	// test for JSON and Protobuf schema Kafka Schema resource
	t.Run("json_protobuf_basic", func(t *testing.T) {
		resourceName := "aiven_kafka_schema.foo"
		resourceName2 := "aiven_kafka_schema.bar"
		fooSubject := acc.RandName("json")
		barSubject := acc.RandName("protobuf")

		// This test can't run in parallel
		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				acc.TestAccPreCheck(t)
				require.NoError(t, <-serviceIsReady)
			},
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaSchemaResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaSchemaJSONProtobufResource(projectName, serviceName, fooSubject, barSubject),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema"),
						testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema2"),
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "subject_name", fooSubject),
						resource.TestCheckResourceAttr(resourceName, "version", "1"),
						resource.TestCheckResourceAttr(resourceName, "schema_type", "JSON"),
						resource.TestCheckResourceAttr(resourceName2, "project", projectName),
						resource.TestCheckResourceAttr(resourceName2, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName2, "subject_name", barSubject),
						resource.TestCheckResourceAttr(resourceName2, "version", "1"),
						resource.TestCheckResourceAttr(resourceName2, "schema_type", "PROTOBUF"),
					),
				},
			},
		})
	})

	t.Run("references", func(t *testing.T) {
		resourceName := "aiven_kafka_schema.dep"
		avroResourceName := "aiven_kafka_schema.avro_dep"
		dataSourceName := "data.aiven_kafka_schema.dep"
		refSubject := acc.RandName("ref") + ".proto"
		depSubject := acc.RandName("dep")
		avroRefSubject := acc.RandName("ref")
		avroDepSubject := acc.RandName("dep")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() {
				acc.TestAccPreCheck(t)
				require.NoError(t, <-serviceIsReady)
			},
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaSchemaResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaSchemaReferencesResource(projectName, serviceName, refSubject, depSubject),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "subject_name", depSubject),
						resource.TestCheckResourceAttr(resourceName, "version", "1"),
						resource.TestCheckResourceAttr(resourceName, "schema_type", "PROTOBUF"),
						resource.TestCheckTypeSetElemNestedAttrs(resourceName, "references.*", map[string]string{
							"name":    refSubject,
							"subject": refSubject,
							"version": "1",
						}),
						resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "references.*", map[string]string{
							"name":    refSubject,
							"subject": refSubject,
							"version": "1",
						}),
					),
				},
				{
					Config:             testAccKafkaSchemaReferencesResource(projectName, serviceName, refSubject, depSubject),
					ExpectNonEmptyPlan: false,
				},
				{
					Config: testAccKafkaSchemaAVROReferencesResource(projectName, serviceName, avroRefSubject, avroDepSubject),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(avroResourceName, "project", projectName),
						resource.TestCheckResourceAttr(avroResourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(avroResourceName, "subject_name", avroDepSubject),
						resource.TestCheckResourceAttr(avroResourceName, "version", "1"),
						resource.TestCheckResourceAttr(avroResourceName, "schema_type", "AVRO"),
					),
				},
				{
					Config: testAccKafkaSchemaAVROReferencesCompatibleResource(projectName, serviceName, avroRefSubject, avroDepSubject),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(avroResourceName, "version", "2"),
						resource.TestCheckResourceAttr(avroResourceName, "schema_type", "AVRO"),
					),
				},
				// Schemas with references skip compatibility checks in CustomizeDiff, so an
				// incompatible change still produces a non-empty plan instead of a plan error.
				{
					Config:             testAccKafkaSchemaAVROReferencesIncompatibleResource(projectName, serviceName, avroRefSubject, avroDepSubject),
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
				{
					Config:      testAccKafkaSchemaAVROReferencesIncompatibleResource(projectName, serviceName, avroRefSubject, avroDepSubject),
					ExpectError: regexp.MustCompile("Incompatible schema, compatibility_mode=BACKWARD. Incompatibilities: reader type: string not compatible with writer type: int"),
				},
			},
		})
	})

	t.Run("historical_references_destroy", func(t *testing.T) {
		resourceName := "aiven_kafka_schema.dep"
		refSubject := acc.RandName("ref") + ".proto"
		depSubject := acc.RandName("dep")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() {
				acc.TestAccPreCheck(t)
				require.NoError(t, <-serviceIsReady)
			},
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaSchemaReferencesDestroy(projectName, serviceName, refSubject),
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaSchemaReferencesResource(projectName, serviceName, refSubject, depSubject),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "version", "1"),
						resource.TestCheckResourceAttr(resourceName, "schema_type", "PROTOBUF"),
						resource.TestCheckTypeSetElemNestedAttrs(resourceName, "references.*", map[string]string{
							"name":    refSubject,
							"subject": refSubject,
							"version": "1",
						}),
					),
				},
				{
					Config: testAccKafkaSchemaReferencesRemovedResource(projectName, serviceName, refSubject, depSubject),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "version", "2"),
						resource.TestCheckResourceAttr(resourceName, "references.#", "0"),
					),
				},
			},
		})
	})
}

func testAccKafkaSchemaAVROReferencesPlan(refSubject, depSubject string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema" "dep" {
  project      = %[1]q
  service_name = "kafka-schema-references-plan"
  subject_name = %[3]q
  schema_type  = "AVRO"

  references {
    name    = "io.aiven.RefRecord"
    subject = %[2]q
    version = 1
  }

  schema = <<EOT
    {
      "type": "record",
      "name": "DepRecord",
      "namespace": "io.aiven",
      "fields": [
        {
          "name": "ref",
          "type": "io.aiven.RefRecord"
        }
      ]
    }
  EOT
}
`, acc.ProjectName(), refSubject, depSubject)
}

func testAccKafkaSchemaAVROReferencesResource(projectName, serviceName, refSubject, depSubject string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_configuration" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "avro_ref" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[3]q
  schema_type  = "AVRO"

  schema = <<EOT
    {
      "type": "record",
      "name": "RefRecord",
      "namespace": "io.aiven",
      "fields": [
        {
          "name": "id",
          "type": "string"
        }
      ]
    }
  EOT
}

resource "aiven_kafka_schema" "avro_dep" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[4]q
  schema_type  = "AVRO"

  references {
    name    = "io.aiven.RefRecord"
    subject = aiven_kafka_schema.avro_ref.subject_name
    version = aiven_kafka_schema.avro_ref.version
  }

  schema = <<EOT
    {
      "type": "record",
      "name": "DepRecord",
      "namespace": "io.aiven",
      "fields": [
        {
          "name": "value",
          "type": "int",
          "default": 0
        },
        {
          "name": "ref",
          "type": "io.aiven.RefRecord"
        }
      ]
    }
  EOT
}`, projectName, serviceName, refSubject, depSubject)
}

func testAccKafkaSchemaAVROReferencesCompatibleResource(projectName, serviceName, refSubject, depSubject string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_configuration" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "avro_ref" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[3]q
  schema_type  = "AVRO"

  schema = <<EOT
    {
      "type": "record",
      "name": "RefRecord",
      "namespace": "io.aiven",
      "fields": [
        {
          "name": "id",
          "type": "string"
        }
      ]
    }
  EOT
}

resource "aiven_kafka_schema" "avro_dep" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[4]q
  schema_type  = "AVRO"

  references {
    name    = "io.aiven.RefRecord"
    subject = aiven_kafka_schema.avro_ref.subject_name
    version = aiven_kafka_schema.avro_ref.version
  }

  schema = <<EOT
    {
      "type": "record",
      "name": "DepRecord",
      "namespace": "io.aiven",
      "fields": [
        {
          "name": "value",
          "type": "int",
          "default": 0
        },
        {
          "name": "ref",
          "type": "io.aiven.RefRecord"
        },
        {
          "name": "note",
          "type": "string",
          "default": "ok"
        }
      ]
    }
  EOT
}`, projectName, serviceName, refSubject, depSubject)
}

func testAccKafkaSchemaAVROReferencesIncompatibleResource(projectName, serviceName, refSubject, depSubject string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_configuration" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "avro_ref" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[3]q
  schema_type  = "AVRO"

  schema = <<EOT
    {
      "type": "record",
      "name": "RefRecord",
      "namespace": "io.aiven",
      "fields": [
        {
          "name": "id",
          "type": "string"
        }
      ]
    }
  EOT
}

resource "aiven_kafka_schema" "avro_dep" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[4]q
  schema_type  = "AVRO"

  references {
    name    = "io.aiven.RefRecord"
    subject = aiven_kafka_schema.avro_ref.subject_name
    version = aiven_kafka_schema.avro_ref.version
  }

  schema = <<EOT
    {
      "type": "record",
      "name": "DepRecord",
      "namespace": "io.aiven",
      "fields": [
        {
          "name": "value",
          "type": "string",
          "default": "not-compatible"
        },
        {
          "name": "ref",
          "type": "io.aiven.RefRecord"
        }
      ]
    }
  EOT
}`, projectName, serviceName, refSubject, depSubject)
}

func testAccKafkaSchemaReferencesResource(projectName, serviceName, refSubject, depSubject string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_configuration" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  compatibility_level = "NONE"
}

resource "aiven_kafka_schema" "ref" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[3]q
  schema_type  = "PROTOBUF"

  schema = <<EOT
syntax = "proto3";

message OtherRecord {
  int32 other_id = 1;
}
EOT
}

resource "aiven_kafka_schema" "dep" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[4]q
  schema_type  = "PROTOBUF"

  references {
    name    = %[3]q
    subject = aiven_kafka_schema.ref.subject_name
    version = aiven_kafka_schema.ref.version
  }

  schema = <<EOT
syntax = "proto3";
import %[3]q;

message MyRecord {
  string f1 = 1;
  OtherRecord f2 = 2;
}
EOT
}

data "aiven_kafka_schema" "dep" {
  project      = aiven_kafka_schema.dep.project
  service_name = aiven_kafka_schema.dep.service_name
  subject_name = aiven_kafka_schema.dep.subject_name
}`, projectName, serviceName, refSubject, depSubject)
}

func testAccKafkaSchemaReferencesRemovedResource(projectName, serviceName, refSubject, depSubject string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_configuration" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  compatibility_level = "NONE"
}

resource "aiven_kafka_schema" "ref" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[3]q
  schema_type  = "PROTOBUF"

  schema = <<EOT
syntax = "proto3";

message OtherRecord {
  int32 other_id = 1;
}
EOT
}

resource "aiven_kafka_schema" "dep" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[4]q
  schema_type  = "PROTOBUF"

  schema = <<EOT
syntax = "proto3";

message MyRecord {
  string f1 = 1;
  string f2 = 2;
}
EOT

  depends_on = [aiven_kafka_schema.ref]
}

data "aiven_kafka_schema" "dep" {
  project      = aiven_kafka_schema.dep.project
  service_name = aiven_kafka_schema.dep.service_name
  subject_name = aiven_kafka_schema.dep.subject_name
}`, projectName, serviceName, refSubject, depSubject)
}

func testAccCheckAivenKafkaSchemaReferencesDestroy(projectName, serviceName, refSubject string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		c, err := acc.GetTestGenAivenClient()
		if err != nil {
			return err
		}

		err = c.ServiceSchemaRegistrySubjectDelete(
			context.Background(),
			projectName,
			serviceName,
			refSubject,
			kafkaschemaregistry.ServiceSchemaRegistrySubjectDeletePermanent(true),
		)
		if avngen.IsNotFound(err) {
			return nil
		}

		return err
	}
}

func testAccCheckAivenKafkaSchemaResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each aiven_kafka_schema is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_kafka" {
			continue
		}

		projectName, serviceName, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceGet(ctx, projectName, serviceName)
		if err != nil {
			if avngen.IsNotFound(err) {
				return nil
			}

			return err
		}

		subjects, err := c.ServiceSchemaRegistrySubjects(ctx, projectName, serviceName)
		if err != nil {
			if avngen.IsNotFound(err) {
				return nil
			}

			return err
		}

		for _, subject := range subjects {
			versions, err := c.ServiceSchemaRegistrySubjectVersionsGet(ctx, projectName, serviceName, subject)
			if err != nil {
				if avngen.IsNotFound(err) {
					return nil
				}

				return err
			}

			if len(versions) > 0 {
				return fmt.Errorf("kafka schema (%s) still exists", subject)
			}
		}

	}

	return nil
}

// testAccKafkaSchemaJSONProtobufResource is a test resource for JSON and Protobuf schema Kafka Schema resource.
func testAccKafkaSchemaJSONProtobufResource(projectName, serviceName, fooSubject, barSubject string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_configuration" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "foo" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[3]q
  schema_type  = "JSON"

  schema = <<EOT
    {
      "type": "object",
      "title": "example",
      "description": "example",
      "properties": {
        "test": {
          "type": "integer",
          "title": "my test number",
          "default": 5
        }
      }
    }
  EOT
}

data "aiven_kafka_schema" "schema" {
  project      = aiven_kafka_schema.foo.project
  service_name = aiven_kafka_schema.foo.service_name
  subject_name = aiven_kafka_schema.foo.subject_name

  depends_on = [aiven_kafka_schema.foo]
}

resource "aiven_kafka_schema" "bar" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[4]q
  schema_type  = "PROTOBUF"

  schema = <<EOT
    syntax = "proto3";

    message Example {
      int32 test = 5;
    }
  EOT
}

data "aiven_kafka_schema" "schema2" {
  project      = aiven_kafka_schema.bar.project
  service_name = aiven_kafka_schema.bar.service_name
  subject_name = aiven_kafka_schema.bar.subject_name

  depends_on = [aiven_kafka_schema.bar]
}`, projectName, serviceName, fooSubject, barSubject)
}

func testAccKafkaSchemaWithSchemaRegistryConfig(sName string, schemaRegistry bool, includeSchema bool) string {
	config := fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    schema_registry = %t

    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}
`, sName, schemaRegistry)

	if includeSchema {
		config += `
resource "aiven_kafka_schema_configuration" "foo" {
  project             = aiven_kafka.bar.project
  service_name        = aiven_kafka.bar.service_name
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "foo" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = "kafka-schema-lifecycle"

  schema = <<EOT
    {
      "doc": "example",
      "fields": [
        {
          "default": 5,
          "doc": "my test number",
          "name": "test",
          "namespace": "test",
          "type": "int"
        }
      ],
      "name": "example",
      "namespace": "example",
      "type": "record"
    }
  EOT
}
`
	}

	return fmt.Sprintf(config, acc.ProjectName())
}

func testAccKafkaSchemaImportCompatibilityLevel(projectName, serviceName, subjectName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema" "schema" {
  project             = %q
  service_name        = %q
  subject_name        = %q
  compatibility_level = "FULL_TRANSITIVE"

  schema = <<EOT
    {
      "doc": "example",
      "fields": [
        {
          "default": 5,
          "doc": "my test number",
          "name": "test",
          "namespace": "test",
          "type": "int"
        }
      ],
      "name": "example",
      "namespace": "example",
      "type": "record"
    }
  EOT
}
`, projectName, serviceName, subjectName)
}

func testAccKafkaSchemaResource(projectName, serviceName, subjectName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_configuration" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "foo" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[3]q

  schema = <<EOT
    {
      "doc": "example",
      "fields": [
        {
          "default": 5,
          "doc": "my test number",
          "name": "test",
          "namespace": "test",
          "type": "int"
        }
      ],
      "name": "example",
      "namespace": "example",
      "type": "record"
    }
  EOT
}

data "aiven_kafka_schema" "schema" {
  project      = aiven_kafka_schema.foo.project
  service_name = aiven_kafka_schema.foo.service_name
  subject_name = aiven_kafka_schema.foo.subject_name

  depends_on = [aiven_kafka_schema.foo]
}`, projectName, serviceName, subjectName)
}

func testAccKafkaSchemaResourceInvalidUpdate(projectName, serviceName, subjectName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_configuration" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "foo" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[3]q

  schema = <<EOT
    {
      "doc": "example",
      "fields": [
        {
          "default": "foo",
          "doc": "my test string",
          "name": "test",
          "namespace": "test",
          "type": "string"
        }
      ],
      "name": "example",
      "namespace": "example",
      "type": "record"
    }
  EOT
}

data "aiven_kafka_schema" "schema" {
  project      = aiven_kafka_schema.foo.project
  service_name = aiven_kafka_schema.foo.service_name
  subject_name = aiven_kafka_schema.foo.subject_name

  depends_on = [aiven_kafka_schema.foo]
}`, projectName, serviceName, subjectName)
}

func testAccKafkaSchemaResourceGoodUpdate(projectName, serviceName, subjectName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_configuration" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "foo" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = %[3]q

  schema = <<EOT
    {
      "doc": "example",
      "fields": [
        {
          "default": 5,
          "doc": "my test number",
          "name": "test",
          "namespace": "test",
          "type": "int"
        },
        {
          "default": "str",
          "doc": "my test string",
          "name": "test_2",
          "namespace": "test",
          "type": "string"
        }
      ],
      "name": "example",
      "namespace": "example",
      "type": "record"
    }
  EOT
}

data "aiven_kafka_schema" "schema" {
  project      = aiven_kafka_schema.foo.project
  service_name = aiven_kafka_schema.foo.service_name
  subject_name = aiven_kafka_schema.foo.subject_name

  depends_on = [aiven_kafka_schema.foo]
}`, projectName, serviceName, subjectName)
}

func testAccCheckAivenKafkaSchemaAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["subject_name"] == "" {
			return fmt.Errorf("expected to get a subject_name from Aiven")
		}

		if a["schema"] == "" {
			return fmt.Errorf("expected to get a schema from Aiven")
		}

		if a["compatibility_level"] != "" {
			return fmt.Errorf("expected to get a corect compatibility_level from Aiven")
		}

		return nil
	}
}

const invalidAvroSchemaConfig = `
resource "aiven_kafka_schema" "foo" {
  project      = "foo"
  service_name = "bar"
  subject_name = "baz"

  schema = <<EOT
    {
	  "name": "foo",
	  "type": "record",
	  "fields": [
		{
		  "name": "foo",
		  "type": "enum",
		  "symbols": ["foo", "bar"]
		}
	  ]
	}
  EOT
}`

func TestAccAivenKafkaSchema_invalid_avro_schema(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenKafkaSchemaResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      invalidAvroSchemaConfig,
				ExpectError: regexp.MustCompile(`Error: schema validation error: avro: unknown type: enum`),
			},
		},
	})
}
