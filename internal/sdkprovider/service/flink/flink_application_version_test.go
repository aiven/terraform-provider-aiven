package flink_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenFlinkApplicationVersion_basic(t *testing.T) {
	resourceName := "aiven_flink_application_version.foo"
	resourceNameDeployment := "aiven_flink_application_deployment.foobar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenFlinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlinkApplicationVersionResource(rName),
				Check: resource.ComposeTestCheckFunc(
					checkAivenFlinkApplicationVersionAttributes("data.aiven_flink_application_version.bar"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(
						resourceName,
						"service_name",
						fmt.Sprintf("test-acc-flink-%s", rName),
					),
					resource.TestCheckResourceAttr(resourceName, "sink.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(
						resourceNameDeployment, "project", os.Getenv("AIVEN_PROJECT_NAME"),
					),
					resource.TestCheckResourceAttr(
						resourceNameDeployment,
						"service_name",
						fmt.Sprintf("test-acc-flink-%s", rName),
					),
				),
			},
		},
	})
}

func testAccCheckAivenFlinkDestroy(s *terraform.State) error {
	client := acc.GetTestAivenClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_flink_application_version" {
			continue
		}

		project, serviceName, applicationID, version, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return err
		}
		v, err := client.FlinkApplicationVersions.Get(project, serviceName, applicationID, version)
		if err != nil && !aiven.IsNotFound(err) {
			return err
		}

		if v != nil {
			return fmt.Errorf("flink application version (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccFlinkApplicationVersionResource(r string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%[1]s"
}

resource "aiven_flink" "foo" {
  project                 = data.aiven_project.foo.project
  service_name            = "test-acc-flink-%[2]s"
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "04:00:00"
}

resource "aiven_kafka" "kafka" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "business-4"
  service_name = "test-acc-kafka-%[2]s"
}

resource "aiven_service_integration" "flink_to_kafka" {
  project                  = data.aiven_project.foo.project
  integration_type         = "flink"
  destination_service_name = aiven_flink.foo.service_name
  source_service_name      = aiven_kafka.kafka.service_name
}

resource "aiven_kafka_topic" "source" {
  project      = aiven_kafka.kafka.project
  service_name = aiven_kafka.kafka.service_name
  partitions   = 2
  replication  = 3
  topic_name   = "source_topic"
}

resource "aiven_kafka_topic" "sink" {
  project      = aiven_kafka.kafka.project
  service_name = aiven_kafka.kafka.service_name
  partitions   = 2
  replication  = 3
  topic_name   = "sink_topic"
}

resource "aiven_flink_application" "foo" {
  project      = data.aiven_project.foo.project
  service_name = aiven_flink.foo.service_name
  name         = "test-acc-flink-application"
}

resource "aiven_flink_application_version" "foo" {
  project        = data.aiven_project.foo.project
  service_name   = aiven_flink.foo.service_name
  application_id = aiven_flink_application.foo.application_id
  statement      = "INSERT INTO kafka_known_pizza SELECT * FROM kafka_pizza WHERE shop LIKE 'Luigis Pizza'"
  sink {
    create_table   = "CREATE TABLE kafka_known_pizza (shop STRING,name STRING) WITH ('connector' = 'kafka','properties.bootstrap.servers' = '','scan.startup.mode' = 'earliest-offset','topic' = 'test_out','value.format' = 'json')"
    integration_id = aiven_service_integration.flink_to_kafka.integration_id
  }
  source {
    create_table   = "CREATE TABLE kafka_pizza (shop STRING, name STRING) WITH ('connector' = 'kafka','properties.bootstrap.servers' = '','scan.startup.mode' = 'earliest-offset','topic' = 'test','value.format' = 'json')"
    integration_id = aiven_service_integration.flink_to_kafka.integration_id
  }
}

resource "aiven_flink_application_deployment" "foobar" {
  project        = data.aiven_project.foo.project
  service_name   = aiven_flink.foo.service_name
  application_id = aiven_flink_application.foo.application_id
  version_id     = data.aiven_flink_application_version.bar.application_version_id
}

data "aiven_flink_application_version" "bar" {
  project                = data.aiven_project.foo.project
  service_name           = aiven_flink.foo.service_name
  application_id         = aiven_flink_application.foo.application_id
  application_version_id = aiven_flink_application_version.foo.application_version_id
}
`, os.Getenv("AIVEN_PROJECT_NAME"), r)
}

func checkAivenFlinkApplicationVersionAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rn, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rn.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		if rn.Primary.Attributes["project"] == "" {
			return fmt.Errorf("no project is set")
		}

		if rn.Primary.Attributes["service_name"] == "" {
			return fmt.Errorf("no service_name is set")
		}

		if rn.Primary.Attributes["application_id"] == "" {
			return fmt.Errorf("no application_id is set")
		}

		if rn.Primary.Attributes["statement"] == "" {
			return fmt.Errorf("no statement is set")
		}

		if rn.Primary.Attributes["source.#"] == "" {
			return fmt.Errorf("no source are set")
		}

		if rn.Primary.Attributes["sink.#"] == "" {
			return fmt.Errorf("no sink are set")
		}

		return nil
	}
}
