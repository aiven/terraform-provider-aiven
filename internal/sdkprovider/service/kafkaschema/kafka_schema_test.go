package kafkaschema_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// TestAccAivenKafkaSchema_import_compatibility_level
// checks that compatibility_level doesn't appear in plan after KafkaSchema import
func TestAccAivenKafkaSchema_import_compatibility_level(t *testing.T) {
	project := acc.ProjectName()
	serviceName := "test-acc-sr-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aiven_kafka_schema.schema"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories:  acc.TestProtoV6ProviderFactories,
		CheckDestroy:              testAccCheckAivenKafkaSchemaResourceDestroy,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Creates resources
				Config: testAccKafkaSchemaImportCompatibilityLevel(project, serviceName, "test-subject"),
			},
			{
				// Imports the schema
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
}

func testAccKafkaSchemaImportCompatibilityLevel(project, serviceName, subjectName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka" "kafka" {
  project      = %q
  service_name = %q
  cloud_name   = "google-europe-west1"
  plan         = "startup-2"

  kafka_user_config {
    schema_registry = true
  }
}

resource "aiven_kafka_schema" "schema" {
  project             = aiven_kafka.kafka.project
  service_name        = aiven_kafka.kafka.service_name
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
`, project, serviceName, subjectName)
}

// TestAccAivenKafkaSchema_json_protobuf_basic is a test for JSON and Protobuf schema Kafka Schema resource.
func TestAccAivenKafkaSchema_json_protobuf_basic(t *testing.T) {
	resourceName := "aiven_kafka_schema.foo"
	resourceName2 := "aiven_kafka_schema.bar"

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenKafkaSchemaResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaSchemaJSONProtobufResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema"),
					testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema2"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(
						resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName),
					),
					resource.TestCheckResourceAttr(
						resourceName, "subject_name", fmt.Sprintf("kafka-schema-%s-foo", rName),
					),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName2, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(
						resourceName2, "service_name", fmt.Sprintf("test-acc-sr-%s", rName),
					),
					resource.TestCheckResourceAttr(
						resourceName2, "subject_name", fmt.Sprintf("kafka-schema-%s-bar", rName),
					),
					resource.TestCheckResourceAttr(resourceName2, "version", "1"),
					resource.TestCheckResourceAttr(resourceName2, "schema_type", "PROTOBUF"),
				),
			},
		},
	})
}

func TestAccAivenKafkaSchema_basic(t *testing.T) {
	resourceName := "aiven_kafka_schema.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenKafkaSchemaResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaSchemaResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "subject_name", fmt.Sprintf("kafka-schema-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_type", "AVRO"),
				),
			},
			{
				Config: testAccKafkaSchemaResourceGoodUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "subject_name", fmt.Sprintf("kafka-schema-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "schema_type", "AVRO"),
				),
			},
			// Reverts changes and gets version=1
			{
				Config: testAccKafkaSchemaResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_type", "AVRO"),
				),
			},
			{
				Config:      testAccKafkaSchemaResourceInvalidUpdate(rName),
				ExpectError: regexp.MustCompile("schema is not compatible with previous version"),
			},
		},
	})
}

func testAccCheckAivenKafkaSchemaResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

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

		_, err = c.Services.Get(ctx, projectName, serviceName)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status == 404 {
				return nil
			}

			return err
		}

		schemaList, err := c.KafkaSubjectSchemas.List(ctx, projectName, serviceName)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status == 404 {
				return nil
			}

			return err
		}

		for _, s := range schemaList.KafkaSchemaSubjects.Subjects {
			versions, err := c.KafkaSubjectSchemas.GetVersions(ctx, projectName, serviceName, s)
			if err != nil {
				var e aiven.Error
				if errors.As(err, &e) && e.Status == 404 {
					return nil
				}

				return err
			}

			if len(versions.Versions) > 0 {
				return fmt.Errorf("kafka schema (%s) still exists", s)
			}
		}

	}

	return nil
}

// testAccKafkaSchemaJSONProtobufResource is a test resource for JSON and Protobuf schema Kafka Schema resource.
func testAccKafkaSchemaJSONProtobufResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%[1]s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    schema_registry = true

    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

resource "aiven_kafka_schema_configuration" "foo" {
  project             = aiven_kafka.bar.project
  service_name        = aiven_kafka.bar.service_name
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "foo" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = "kafka-schema-%[2]s-foo"
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
  subject_name = "kafka-schema-%[2]s-bar"
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
}`, acc.ProjectName(), name)
}

func testAccKafkaSchemaResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    schema_registry = true

    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

resource "aiven_kafka_schema_configuration" "foo" {
  project             = aiven_kafka.bar.project
  service_name        = aiven_kafka.bar.service_name
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "foo" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = "kafka-schema-%s"

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
}`, acc.ProjectName(), name, name)
}

func testAccKafkaSchemaResourceInvalidUpdate(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  kafka_user_config {
    schema_registry = true

    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

resource "aiven_kafka_schema_configuration" "foo" {
  project             = aiven_kafka.bar.project
  service_name        = aiven_kafka.bar.service_name
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "foo" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = "kafka-schema-%s"

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
}`, acc.ProjectName(), name, name)
}

func testAccKafkaSchemaResourceGoodUpdate(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  kafka_user_config {
    schema_registry = true

    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

resource "aiven_kafka_schema_configuration" "foo" {
  project             = aiven_kafka.bar.project
  service_name        = aiven_kafka.bar.service_name
  compatibility_level = "BACKWARD"
}

resource "aiven_kafka_schema" "foo" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
  subject_name = "kafka-schema-%s"

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
}`, acc.ProjectName(), name, name)
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
