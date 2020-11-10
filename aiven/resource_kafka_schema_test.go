package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

func init() {
	resource.AddTestSweepers("aiven_kafka_schema", &resource.Sweeper{
		Name: "aiven_kafka_schema",
		F:    sweepKafkaSchemas,
	})
}

func sweepKafkaSchemas(region string) error {
	client, err := sharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projects, err := conn.Projects.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of projects : %s", err)
	}

	for _, project := range projects {
		if project.Name == os.Getenv("AIVEN_PROJECT_NAME") {
			services, err := conn.Services.List(project.Name)
			if err != nil {
				return fmt.Errorf("error retrieving a list of services for a project `%s`: %s", project.Name, err)
			}

			for _, service := range services {
				if service.Type != "kafka" {
					continue
				}

				schemaList, err := conn.KafkaSubjectSchemas.List(project.Name, service.Name)
				if err != nil {
					if err.(aiven.Error).Status == 403 {
						continue
					}

					return fmt.Errorf("error retrieving a list of kafka schemas for a service `%s`: %s", service.Name, err)
				}

				for _, s := range schemaList.Subjects {
					err = conn.KafkaSubjectSchemas.Delete(project.Name, service.Name, s)
					if err != nil {
						return fmt.Errorf("error destroying kafka schema `%s` during sweep: %s", s, err)
					}
				}
			}
		}
	}

	return nil
}

func TestAccAivenKafkaSchema_basic(t *testing.T) {
	resourceName := "aiven_kafka_schema.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenKafkaSchemaResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaSchemaResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenKafkaSchemaAttributes("data.aiven_kafka_schema.schema"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "subject_name", fmt.Sprintf("kafka-schema-%s", rName)),
				),
			},
		},
	})
}

func testAccCheckAivenKafkaSchemaResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each aiven_kafka_schema is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_service" {
			continue
		}

		projectName, serviceName := splitResourceID2(rs.Primary.ID)
		_, err := c.Services.Get(projectName, serviceName)
		if err != nil {
			if err.(aiven.Error).Status == 404 {
				return nil
			}

			return err
		}

		schemaList, err := c.KafkaSubjectSchemas.List(projectName, serviceName)
		if err != nil {
			if err.(aiven.Error).Status == 404 {
				return nil
			}

			return err
		}

		for _, s := range schemaList.KafkaSchemaSubjects.Subjects {
			versions, err := c.KafkaSubjectSchemas.GetVersions(projectName, serviceName, s)
			if err != nil {
				if err.(aiven.Error).Status == 404 {
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

func testAccKafkaSchemaResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
			project = "%s"
		}

		resource "aiven_service" "bar" {
			project = data.aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "business-4"
			service_name = "test-acc-sr-%s"
			service_type = "kafka"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			kafka_user_config {
				schema_registry = true
				kafka_version = "2.4"

				kafka {
				  group_max_session_timeout_ms = 70000
				  log_retention_bytes = 1000000000
				}
			}
		}
		
		resource "aiven_kafka_schema_configuration" "foo" {
			project = data.aiven_project.foo.project
			service_name = aiven_service.bar.service_name
			compatibility_level = "BACKWARD"
		}

		resource "aiven_kafka_schema" "foo" {
			project = data.aiven_project.foo.project
			service_name = aiven_service.bar.service_name
			subject_name = "kafka-schema-%s"
			
			schema = <<EOT
				{
					"doc": "example",
					"fields": [{
						"default": 5,
						"doc": "my test number",
						"name": "test",
						"namespace": "test",
						"type": "int"
					}],
					"name": "example",
					"namespace": "example",
					"type": "record"
				}
			EOT
		}

		data "aiven_kafka_schema" "schema" {
			project = aiven_kafka_schema.foo.project
			service_name = aiven_kafka_schema.foo.service_name
			subject_name = aiven_kafka_schema.foo.subject_name

			depends_on = [aiven_kafka_schema.foo]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccCheckAivenKafkaSchemaAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven)")
		}

		if a["subject_name"] == "" {
			return fmt.Errorf("expected to get a subject_name from Aiven)")
		}

		if a["schema"] == "" {
			return fmt.Errorf("expected to get a schema from Aiven)")
		}

		if a["compatibility_level"] != "" {
			return fmt.Errorf("expected to get a corect compatibility_level from Aiven)")
		}

		return nil
	}
}
