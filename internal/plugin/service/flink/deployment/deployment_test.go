package deployment_test

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenFlinkApplicationDeployment(t *testing.T) {
	resourceName := "aiven_flink_application_deployment.test"
	projectName := acc.ProjectName()

	t.Run("backward compatibility", func(t *testing.T) {
		serviceName := acc.RandName("flink")
		kafkaServiceName := acc.RandName("kafka")
		config := testAccFlinkApplicationDeploymentConfig(projectName, serviceName, kafkaServiceName, "")
		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig:           config,
				OldProviderVersion: "4.50.0",
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "restart_enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			}),
		})
	})

	t.Run("basic", func(t *testing.T) {
		serviceName := acc.RandName("flink")
		kafkaServiceName := acc.RandName("kafka")
		config := testAccFlinkApplicationDeploymentConfig(projectName, serviceName, kafkaServiceName, "")
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckFlinkApplicationDeploymentDestroy,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttrSet(resourceName, "application_id"),
						resource.TestCheckResourceAttrSet(resourceName, "version_id"),
						resource.TestCheckResourceAttrSet(resourceName, "deployment_id"),
						resource.TestCheckResourceAttr(resourceName, "parallelism", "1"),
						resource.TestCheckResourceAttr(resourceName, "restart_enabled", "true"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
						resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					),
				},
				{
					Config:            config,
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("custom values", func(t *testing.T) {
		serviceName := acc.RandName("flink")
		kafkaServiceName := acc.RandName("kafka")
		config := testAccFlinkApplicationDeploymentConfig(projectName, serviceName, kafkaServiceName, `
  parallelism      = 2
  restart_enabled  = false
`)
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckFlinkApplicationDeploymentDestroy,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttrSet(resourceName, "application_id"),
						resource.TestCheckResourceAttrSet(resourceName, "version_id"),
						resource.TestCheckResourceAttrSet(resourceName, "deployment_id"),
						resource.TestCheckResourceAttr(resourceName, "parallelism", "2"),
						resource.TestCheckResourceAttr(resourceName, "restart_enabled", "false"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
						resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					),
				},
			},
		})
	})
}

func testAccCheckFlinkApplicationDeploymentDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_flink_application_deployment" {
			continue
		}

		project, serviceName, applicationID, deploymentID, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceFlinkGetApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
		if avngen.IsNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("flink application deployment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccFlinkApplicationDeploymentConfig(project, serviceName, kafkaServiceName, deploymentExtra string) string {
	return fmt.Sprintf(`
resource "aiven_kafka" "kafka" {
  project      = %[1]q
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = %[3]q
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

resource "aiven_flink" "flink" {
  project      = %[1]q
  cloud_name   = "google-europe-west1"
  plan         = "business-4"
  service_name = %[2]q

  flink_user_config {
    number_of_task_slots = 10
  }
}

resource "aiven_service_integration" "flink_to_kafka" {
  project                  = %[1]q
  integration_type         = "flink"
  destination_service_name = aiven_flink.flink.service_name
  source_service_name      = aiven_kafka.kafka.service_name
}

resource "aiven_flink_application" "app" {
  project      = %[1]q
  service_name = aiven_flink.flink.service_name
  name         = "test-app"
}

resource "aiven_flink_application_version" "version" {
  project        = %[1]q
  service_name   = aiven_flink.flink.service_name
  application_id = aiven_flink_application.app.application_id
  statement      = <<EOT
    INSERT INTO kafka_known_pizza SELECT * FROM kafka_pizza WHERE shop LIKE '%%Luigis Pizza%%'
  EOT
  sink {
    create_table   = <<EOT
      CREATE TABLE kafka_known_pizza (
        shop STRING,
        name STRING
      ) WITH (
        'connector' = 'kafka',
        'properties.bootstrap.servers' = '',
        'scan.startup.mode' = 'earliest-offset',
        'topic' = 'sink_topic',
        'value.format' = 'json'
      )
    EOT
    integration_id = aiven_service_integration.flink_to_kafka.integration_id
  }
  source {
    create_table   = <<EOT
      CREATE TABLE kafka_pizza (
        shop STRING,
        name STRING
      ) WITH (
        'connector' = 'kafka',
        'properties.bootstrap.servers' = '',
        'scan.startup.mode' = 'earliest-offset',
        'topic' = 'source_topic',
        'value.format' = 'json'
      )
    EOT
    integration_id = aiven_service_integration.flink_to_kafka.integration_id
  }
}

resource "aiven_flink_application_deployment" "test" {
  project        = %[1]q
  service_name   = aiven_flink.flink.service_name
  application_id = aiven_flink_application.app.application_id
  version_id     = aiven_flink_application_version.version.application_version_id
  %[4]s
}
`, project, serviceName, kafkaServiceName, deploymentExtra)
}
