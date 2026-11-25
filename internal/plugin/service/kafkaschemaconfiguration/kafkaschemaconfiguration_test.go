package kafkaschemaconfiguration_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenKafkaSchemaConfiguration_basic(t *testing.T) {
	resourceName := "aiven_kafka_schema_configuration.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories:  acc.TestProtoV6ProviderFactories,
		PreventPostDestroyRefresh: true,
		CheckDestroy:              testAccCheckAivenKafkaSchemaConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaSchemaConfigurationResource(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compatibility_level", "BACKWARD"),
				),
			},
			{
				Config: testAccKafkaSchemaConfigurationResource(rName, false, false),
			},
			{
				Config: testAccKafkaSchemaConfigurationResource(rName, true, false),
			},
			{
				Config: testAccKafkaSchemaConfigurationResource(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compatibility_level", "BACKWARD"),
				),
			},
		},
	})
}

func TestAccAivenKafkaSchemaConfiguration_backwardCompat(t *testing.T) {
	projectName := acc.ProjectName()
	serviceName := acc.RandName("kafka")
	resourceName := "aiven_kafka_schema_configuration.foo"
	datasourceName := "data.aiven_kafka_schema_configuration.config"

	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("kafka"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
		acc.WithUserConfig(map[string]any{"schema_registry": true}),
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			PreConfig: func() { require.NoError(t, <-serviceIsReady) },
			TFConfig:  testAccKafkaSchemaConfigurationBackwardCompatConfig(projectName, serviceName),
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "project", projectName),
				resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
				resource.TestCheckResourceAttr(resourceName, "compatibility_level", "FORWARD"),
				resource.TestCheckResourceAttr(datasourceName, "compatibility_level", "FORWARD"),
			),
		}),
	})
}

func testAccCheckAivenKafkaSchemaConfigurationResourceDestroy(_ *terraform.State) error {
	return nil
}

func testAccKafkaSchemaConfigurationResource(name string, schemaRegistry bool, includeConfig bool) string {
	config := fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
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
`, acc.ProjectName(), name, schemaRegistry)

	if includeConfig {
		config += `
resource "aiven_kafka_schema_configuration" "foo" {
  project             = data.aiven_project.foo.project
  service_name        = aiven_kafka.bar.service_name
  compatibility_level = "BACKWARD"
}`
	}
	return config
}

func testAccKafkaSchemaConfigurationBackwardCompatConfig(projectName, serviceName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_configuration" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  compatibility_level = "FORWARD"
}

data "aiven_kafka_schema_configuration" "config" {
  project      = aiven_kafka_schema_configuration.foo.project
  service_name = aiven_kafka_schema_configuration.foo.service_name
}
`, projectName, serviceName)
}
