package serviceintegration_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenServiceIntegration_should_fail(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceIntegrationShouldFailResource(),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("endpoint id should have the following format: project_name/endpoint_id"),
			},
		},
	})
}

func TestAccAivenServiceIntegration_preexisting_read_replica(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationPreexistingReadReplica(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceIntegrationAttributes("data.aiven_service_integration.int"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "read_replica"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-source-pg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-sink-pg-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegration_logs(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationLogs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceIntegrationAttributes("data.aiven_service_integration.int"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "logs"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-source-pg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-sink-os-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegration_mm(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationMirrorMakerResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "integration_type", "kafka_mirrormaker"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-source-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-mm-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegration_kafka_connect(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationKafkaConnectResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceIntegrationAttributes("data.aiven_service_integration.int"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "kafka_connect"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-kafka-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-kafka-con-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegration_basic(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceIntegrationAttributes("data.aiven_service_integration.int"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "metrics"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-pg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-influxdb-%s", rName)),
				),
			},
		},
	})
}

func testAccServiceIntegrationResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar-pg" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-pg-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

resource "aiven_influxdb" "bar-influxdb" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-influxdb-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  influxdb_user_config {
    public_access {
      influxdb = true
    }
  }
}

resource "aiven_service_integration" "bar" {
  project                  = data.aiven_project.foo.project
  integration_type         = "metrics"
  source_service_name      = aiven_pg.bar-pg.service_name
  destination_service_name = aiven_influxdb.bar-influxdb.service_name

  depends_on = [aiven_pg.bar-pg, aiven_influxdb.bar-influxdb]
}

data "aiven_service_integration" "int" {
  project                  = aiven_service_integration.bar.project
  integration_type         = aiven_service_integration.bar.integration_type
  source_service_name      = aiven_service_integration.bar.source_service_name
  destination_service_name = aiven_service_integration.bar.destination_service_name

  depends_on = [aiven_service_integration.bar]
}`, acc.ProjectName(), name, name)
}

func testAccServiceIntegrationKafkaConnectResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "kafka1" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-kafka-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_kafka_connect" "kafka_connect1" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-kafka-con-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_connect_user_config {
    kafka_connect {
      consumer_isolation_level = "read_committed"
    }

    public_access {
      kafka_connect = true
    }
  }
}

resource "aiven_service_integration" "bar" {
  project                  = data.aiven_project.foo.project
  integration_type         = "kafka_connect"
  source_service_name      = aiven_kafka.kafka1.service_name
  destination_service_name = aiven_kafka_connect.kafka_connect1.service_name

  kafka_connect_user_config {
    kafka_connect {
      group_id             = "connect"
      status_storage_topic = "__connect_status"
      offset_storage_topic = "__connect_offsets"
    }
  }
}

data "aiven_service_integration" "int" {
  project                  = aiven_service_integration.bar.project
  integration_type         = aiven_service_integration.bar.integration_type
  source_service_name      = aiven_service_integration.bar.source_service_name
  destination_service_name = aiven_service_integration.bar.destination_service_name

  depends_on = [aiven_service_integration.bar]
}`, acc.ProjectName(), name, name)
}

func testAccServiceIntegrationMirrorMakerResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "source" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-source-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

resource "aiven_kafka_topic" "source" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.source.service_name
  topic_name   = "test-acc-topic-a-%s"
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka" "target" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-target-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

resource "aiven_kafka_topic" "target" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.target.service_name
  topic_name   = "test-acc-topic-b-%s"
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka_mirrormaker" "mm" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-mm-%s"

  kafka_mirrormaker_user_config {
    ip_filter = ["0.0.0.0/0"]

    kafka_mirrormaker {
      refresh_groups_interval_seconds = 600
      refresh_topics_enabled          = true
      refresh_topics_interval_seconds = 600
    }
  }
}

resource "aiven_service_integration" "bar" {
  project                  = data.aiven_project.foo.project
  integration_type         = "kafka_mirrormaker"
  source_service_name      = aiven_kafka.source.service_name
  destination_service_name = aiven_kafka_mirrormaker.mm.service_name

  kafka_mirrormaker_user_config {
    cluster_alias = "source"
  }
}

resource "aiven_service_integration" "i2" {
  project                  = data.aiven_project.foo.project
  integration_type         = "kafka_mirrormaker"
  source_service_name      = aiven_kafka.target.service_name
  destination_service_name = aiven_kafka_mirrormaker.mm.service_name

  kafka_mirrormaker_user_config {
    cluster_alias = "target"
  }
}`, acc.ProjectName(), name, name, name, name, name)
}

func testAccServiceIntegrationLogs(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "source" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-source-pg-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "00:00:00"
}

resource "aiven_opensearch" "sink" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-sink-os-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "00:00:00"
}

resource "aiven_service_integration" "bar" {
  project                  = data.aiven_project.foo.project
  integration_type         = "logs"
  source_service_name      = resource.aiven_pg.source.service_name
  destination_service_name = resource.aiven_opensearch.sink.service_name
  logs_user_config {
    elasticsearch_index_days_max = "2"
    elasticsearch_index_prefix   = "logs"
  }
}

data "aiven_service_integration" "int" {
  project                  = aiven_service_integration.bar.project
  integration_type         = aiven_service_integration.bar.integration_type
  source_service_name      = aiven_service_integration.bar.source_service_name
  destination_service_name = aiven_service_integration.bar.destination_service_name

  depends_on = [aiven_service_integration.bar]
}`, acc.ProjectName(), name, name)
}

func testAccServiceIntegrationPreexistingReadReplica(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "source" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-source-pg-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "00:00:00"
}

resource "aiven_pg" "sink" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-sink-pg-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "00:00:00"
  service_integrations {
    integration_type    = "read_replica"
    source_service_name = resource.aiven_pg.source.service_name
  }
}

resource "aiven_service_integration" "bar" {
  project                  = data.aiven_project.foo.project
  integration_type         = "read_replica"
  source_service_name      = resource.aiven_pg.source.service_name
  destination_service_name = resource.aiven_pg.sink.service_name
}

data "aiven_service_integration" "int" {
  project                  = aiven_service_integration.bar.project
  integration_type         = aiven_service_integration.bar.integration_type
  source_service_name      = aiven_service_integration.bar.source_service_name
  destination_service_name = aiven_service_integration.bar.destination_service_name

  depends_on = [aiven_service_integration.bar]
}`, acc.ProjectName(), name, name)
}

func testAccServiceIntegrationShouldFailResource() string {
	return `
resource "aiven_service_integration" "bar" {
  project                 = "test"
  integration_type        = "kafka_mirrormaker"
  source_endpoint_id      = "test"
  destination_endpoint_id = "test"

  kafka_mirrormaker_user_config {
    cluster_alias = "source"
  }
}`
}

func testAccCheckAivenServiceIntegrationResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each aiven_service_integration is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_service_integration" {
			continue
		}

		projectName, integrationID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		i, err := c.ServiceIntegrations.Get(ctx, projectName, integrationID)
		if common.IsCritical(err) {
			return err
		}
		if i != nil {
			return fmt.Errorf("common integration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAivenServiceIntegrationAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project from Aiven")
		}

		if a["integration_type"] == "" {
			return fmt.Errorf("expected to get an integration_type from Aiven")
		}

		if a["source_service_name"] == "" {
			return fmt.Errorf("expected to get a source_service_name from Aiven")
		}

		if a["destination_service_name"] == "" {
			return fmt.Errorf("expected to get a source_service_name from Aiven")
		}

		return nil
	}
}

func testAccServiceIntegrationDatadogWithUserConfig(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "postgres" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-postgres-%s"
}

resource "aiven_service_integration_endpoint" "datadog" {
  project       = data.aiven_project.foo.project
  endpoint_name = "postgres-datadog-%s"
  endpoint_type = "datadog"

  datadog_user_config {
    datadog_api_key = "11111111222233334444555555555555"

    datadog_tags {
      tag = "bar:foo"
    }
  }
}

resource "aiven_service_integration" "postgres-datadog" {
  project                 = data.aiven_project.foo.project
  integration_type        = "datadog"
  source_service_name     = aiven_pg.postgres.service_name
  destination_endpoint_id = aiven_service_integration_endpoint.datadog.id

  datadog_user_config {
    datadog_tags {
      tag     = "lol:bar"
      comment = "my custom config"
    }
  }
}
`, acc.ProjectName(), name, name)
}

func TestAccAivenServiceIntegration_datadog_with_user_config_creates(t *testing.T) {
	resourceName := "aiven_service_integration.postgres-datadog"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationDatadogWithUserConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "integration_type", "datadog"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-postgres-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "datadog_user_config.0.datadog_tags.0.tag", "lol:bar"),
					resource.TestCheckResourceAttr(resourceName, "datadog_user_config.0.datadog_tags.0.comment", "my custom config"),
				),
			},
		},
	})
}

func testAccServiceIntegrationClickhouseKafkaUserConfig(prefix, project, tables string) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}

resource "aiven_kafka" "kafka" {
  project                 = data.aiven_project.project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "%[1]s-kafka"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    // Enables Kafka Schemas
    schema_registry = true
  }
}

resource "aiven_clickhouse" "clickhouse" {
  project                 = data.aiven_project.project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-16"
  service_name            = "%[1]s-clickhouse"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_service_integration" "clickhouse_kafka_source" {
  project                  = data.aiven_project.project.project
  integration_type         = "clickhouse_kafka"
  source_service_name      = aiven_kafka.kafka.service_name
  destination_service_name = aiven_clickhouse.clickhouse.service_name

  clickhouse_kafka_user_config {
    dynamic "tables" {
      for_each = %[3]s
      content {
        data_format   = "AvroConfluent"
        name          = tables.key
        group_name    = tables.key
        num_consumers = 1
        topics {
          name = tables.key
        }
        columns {
          name = tables.key
          type = "LowCardinality(String)"
        }
      }
    }
  }
}
`, prefix, project, tables)
}

func TestAccAivenServiceIntegration_clickhouse_kafka_user_config_creates(t *testing.T) {
	project := acc.ProjectName()
	prefix := "test-acc-" + acctest.RandString(7)
	resourceName := "aiven_service_integration.clickhouse_kafka_source"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationClickhouseKafkaUserConfig(prefix, project, "[]"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "integration_type", "clickhouse_kafka"),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", prefix+"-kafka"),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", prefix+"-clickhouse"),
					resource.TestCheckResourceAttr(resourceName, "clickhouse_kafka_user_config.0.tables.#", "0"),
				),
			},
			{
				// Tests table removal: adds three
				Config: testAccServiceIntegrationClickhouseKafkaUserConfig(prefix, project, `["foo", "bar", "baz"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "clickhouse_kafka_user_config.0.tables.#", "3"),
				),
			},
			{
				// Tests table removal: leaves one
				Config: testAccServiceIntegrationClickhouseKafkaUserConfig(prefix, project, `["bar"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "clickhouse_kafka_user_config.0.tables.#", "1"),
				),
			},
		},
	})
}

func testAccServiceIntegrationClickhousePostgresUserConfig(prefix, project string) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}

resource "aiven_pg" "pg" {
  project      = data.aiven_project.project.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "%[1]s-pg"
}

resource "aiven_clickhouse" "clickhouse" {
  project                 = data.aiven_project.project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-16"
  service_name            = "%[1]s-clickhouse"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_service_integration" "clickhouse_pg_source" {
  project                  = data.aiven_project.project.project
  integration_type         = "clickhouse_postgresql"
  source_service_name      = aiven_pg.pg.service_name
  destination_service_name = aiven_clickhouse.clickhouse.service_name

  clickhouse_postgresql_user_config {
    databases {
      database = "defaultdb"
      schema   = "public"
    }
  }
}
`, prefix, project)
}

func TestAccAivenServiceIntegration_clickhouse_postgres_user_config_creates(t *testing.T) {
	project := acc.ProjectName()
	prefix := "test-acc-" + acctest.RandString(7)
	resourceName := "aiven_service_integration.clickhouse_pg_source"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationClickhousePostgresUserConfig(prefix, project),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "integration_type", "clickhouse_postgresql"),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", prefix+"-pg"),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", prefix+"-clickhouse"),
				),
			},
		},
	})
}

func testAccServiceIntegrationAutoscaler(prefix string, includeDiskSpace bool) string {
	additionalDiskSpace := ""
	if includeDiskSpace {
		additionalDiskSpace = `additional_disk_space = "30GiB"`
	}

	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[1]q
}

resource "aiven_pg" "test_pg" {
  project      = data.aiven_project.project.project
  cloud_name   = "google-europe-north1"
  service_name = "%[2]s-pg"
  plan         = "startup-4"
  %[3]s
}

resource "aiven_service_integration_endpoint" "test_endpoint" {
  project       = data.aiven_project.project.project
  endpoint_name = "%[2]s-autoscaler"
  endpoint_type = "autoscaler"

  autoscaler_user_config {
    autoscaling {
      cap_gb = 200
      type   = "autoscale_disk"
    }
  }
}

resource "aiven_service_integration" "test_autoscaler" {
  project                 = data.aiven_project.project.project
  integration_type        = "autoscaler"
  source_service_name     = aiven_pg.test_pg.service_name
  destination_endpoint_id = aiven_service_integration_endpoint.test_endpoint.id
}
`, acc.ProjectName(), prefix, additionalDiskSpace)
}

func TestAccAivenServiceIntegration_autoscaler(t *testing.T) {
	project := acc.ProjectName()
	prefix := "test-acc-" + acctest.RandString(7)
	resourceName := "aiven_service_integration.test_autoscaler"
	endpointResourceName := "aiven_service_integration_endpoint.test_endpoint"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { /* Add necessary pre-checks here */ },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationAutoscaler(prefix, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "integration_type", "autoscaler"),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("%s-pg", prefix)),
					resource.TestCheckResourceAttrSet(resourceName, "destination_endpoint_id"),
					resource.TestCheckResourceAttr(endpointResourceName, "project", project),
					resource.TestCheckResourceAttr(endpointResourceName, "endpoint_name", fmt.Sprintf("%s-autoscaler", prefix)),
					resource.TestCheckResourceAttr(endpointResourceName, "endpoint_type", "autoscaler"),
					resource.TestCheckResourceAttr(endpointResourceName, "autoscaler_user_config.0.autoscaling.0.cap_gb", "200"),
				),
			},
			{
				Config:      testAccServiceIntegrationAutoscaler(prefix, true),
				ExpectError: regexp.MustCompile(schemautil.ErrAutoscalerConflict.Error()),
			},
		},
	})
}

func TestAccAivenServiceIntegration_destination_service_name(t *testing.T) {
	projectMetrics := acc.ProjectName()
	projectServices := os.Getenv("AIVEN_PROJECT_NAME_SECONDARY")
	if projectServices == "" {
		t.Skip("AIVEN_PROJECT_NAME_SECONDARY is not set")
	}

	thanosToGrafanaName := "aiven_service_integration.thanos_to_grafana"
	kafkaToThanosName := "aiven_service_integration.kafka_to_thanos"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAivenServiceIntegrationDestinationServiceName(projectMetrics, projectServices),
				Check: resource.ComposeTestCheckFunc(
					// Same source and destination project names
					resource.TestCheckResourceAttr(thanosToGrafanaName, "integration_type", "dashboard"),
					resource.TestCheckResourceAttr(thanosToGrafanaName, "source_service_project", projectMetrics),
					resource.TestCheckResourceAttr(thanosToGrafanaName, "destination_service_project", projectMetrics),

					// Different source and destination project names
					resource.TestCheckResourceAttr(kafkaToThanosName, "integration_type", "metrics"),
					resource.TestCheckResourceAttr(kafkaToThanosName, "source_service_project", projectServices),
					resource.TestCheckResourceAttr(kafkaToThanosName, "destination_service_project", projectMetrics),
				),
			},
		},
	})
}

func testAccAivenServiceIntegrationDestinationServiceName(projectMetrics, projectService string) string {
	return fmt.Sprintf(`
data "aiven_project" "metrics" {
  project = %[1]q
}

data "aiven_project" "services" {
  project = %[2]q
}

resource "aiven_grafana" "grafana" {
  project                 = data.aiven_project.metrics.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-1"
  service_name            = "test-acc-grafana-%[3]s"
  maintenance_window_dow  = "sunday"
  maintenance_window_time = "10:00:00"

  grafana_user_config {
    alerting_enabled = true
  }
}

resource "aiven_thanos" "thanos" {
  project                 = data.aiven_project.metrics.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-thanos-%[3]s"
  maintenance_window_dow  = "sunday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_kafka" "kafka_service" {
  project                 = data.aiven_project.services.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-kafka-%[3]s"
  maintenance_window_dow  = "sunday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    schema_registry = true
    kafka_rest      = true

    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

resource "aiven_service_integration" "thanos_to_grafana" {
  project                     = data.aiven_project.metrics.project
  integration_type            = "dashboard"
  source_service_name         = aiven_grafana.grafana.service_name // project "metrics"
  destination_service_name    = aiven_thanos.thanos.service_name   // project "metrics"
  destination_service_project = aiven_thanos.thanos.project
}

resource "aiven_service_integration" "kafka_to_thanos" {
  project                     = data.aiven_project.services.project
  integration_type            = "metrics"
  source_service_name         = aiven_kafka.kafka_service.service_name // project "services"
  destination_service_name    = aiven_thanos.thanos.service_name       // project "metrics"
  destination_service_project = aiven_thanos.thanos.project
}
	`, projectMetrics, projectService, acc.RandStr())
}
