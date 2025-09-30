package kafka_test

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

func TestAccAivenMirrorMakerReplicationFlow_basic(t *testing.T) {
	resourceName := "aiven_mirrormaker_replication_flow.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenMirrorMakerReplicationFlowResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMirrorMakerReplicationFlowResource(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenMirrorMakerReplicationFlowAttributes("data.aiven_mirrormaker_replication_flow.flow"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-mm-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "source_cluster", "source"),
					resource.TestCheckResourceAttr(resourceName, "target_cluster", "target"),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "offset_syncs_topic_location", "source"),
					resource.TestCheckResourceAttr(resourceName, "replication_factor", "2"),
					resource.TestCheckResourceAttr(resourceName, "topics_blacklist.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "topics_blacklist.0", ".*[\\-\\.]internal"),
					resource.TestCheckResourceAttr(resourceName, "topics_blacklist.1", ".*\\.replica"),
					resource.TestCheckResourceAttr(resourceName, "topics_blacklist.2", "__.*"),
					resource.TestCheckResourceAttr(resourceName, "config_properties_exclude.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "exactly_once_delivery_enabled", "true"),
				),
			},
			{
				Config: testAccMirrorMakerReplicationFlowResource(rName, `
				config_properties_exclude = [
					"follower\\.replication\\.throttled\\.replicas",
					"leader\\.replication\\.throttled\\.replicas",
					"message\\.timestamp\\.difference\\.max\\.ms",
					"message\\.timestamp\\.type",
					"unclean\\.leader\\.election\\.enable",
					"min\\.insync\\.replicas"
				]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "config_properties_exclude.#", "6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "config_properties_exclude.*", "follower\\.replication\\.throttled\\.replicas"),
					resource.TestCheckTypeSetElemAttr(resourceName, "config_properties_exclude.*", "leader\\.replication\\.throttled\\.replicas"),
					resource.TestCheckTypeSetElemAttr(resourceName, "config_properties_exclude.*", "message\\.timestamp\\.difference\\.max\\.ms"),
					resource.TestCheckTypeSetElemAttr(resourceName, "config_properties_exclude.*", "message\\.timestamp\\.type"),
					resource.TestCheckTypeSetElemAttr(resourceName, "config_properties_exclude.*", "unclean\\.leader\\.election\\.enable"),
					resource.TestCheckTypeSetElemAttr(resourceName, "config_properties_exclude.*", "min\\.insync\\.replicas"),
				),
			},
			{
				// Removes the config
				Config: testAccMirrorMakerReplicationFlowResource(rName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "config_properties_exclude.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAivenMirrorMakerReplicationFlowResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each kafka mirror maker
	// replication flow is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_mirrormaker_replication_flow" {
			continue
		}

		project, serviceName, sourceCluster, targetCluster, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return err
		}

		s, err := c.Services.Get(ctx, project, serviceName)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status != 404 {
				return err
			}
			return nil
		}

		if s.Type == "kafka_mirrormaker" {
			f, err := c.KafkaMirrorMakerReplicationFlow.Get(ctx, project, serviceName, sourceCluster, targetCluster)
			if err != nil {
				var e aiven.Error
				if errors.As(err, &e) && e.Status != 404 {
					return err
				}
			}

			if f != nil {
				return fmt.Errorf("kafka mirror maker replication flow still exists, id %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccMirrorMakerReplicationFlowResource(name, configExclude string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%[1]s"
}

resource "aiven_kafka" "source" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-source-%[2]s"
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
  topic_name   = "test-acc-topic-a-%[2]s"
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka" "target" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-target-%[2]s"
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
  topic_name   = "test-acc-topic-b-%[2]s"
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka_mirrormaker" "mm" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-mm-%[2]s"

  kafka_mirrormaker_user_config {
    ip_filter = ["0.0.0.0/0", "::/0"]

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
}

resource "aiven_mirrormaker_replication_flow" "foo" {
  project                             = data.aiven_project.foo.project
  service_name                        = aiven_kafka_mirrormaker.mm.service_name
  source_cluster                      = "source"
  target_cluster                      = "target"
  enable                              = true
  replication_policy_class            = "org.apache.kafka.connect.mirror.IdentityReplicationPolicy"
  replication_factor                  = 2
  sync_group_offsets_enabled          = true
  sync_group_offsets_interval_seconds = 10
  emit_heartbeats_enabled             = true
  emit_backward_heartbeats_enabled    = true
  offset_syncs_topic_location         = "source"
  exactly_once_delivery_enabled       = true

  topics = [
    ".*",
  ]

  topics_blacklist = [
    ".*[\\-\\.]internal",
    ".*\\.replica",
    "__.*"
  ]

  %[3]s
}

data "aiven_mirrormaker_replication_flow" "flow" {
  project        = data.aiven_project.foo.project
  service_name   = aiven_kafka_mirrormaker.mm.service_name
  source_cluster = aiven_mirrormaker_replication_flow.foo.source_cluster
  target_cluster = aiven_mirrormaker_replication_flow.foo.target_cluster

  depends_on = [aiven_mirrormaker_replication_flow.foo]
}`, acc.ProjectName(), name, configExclude)
}

func testAccCheckAivenMirrorMakerReplicationFlowAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["source_cluster"] != "source" {
			return fmt.Errorf("expected to get a source_cluster from Aiven")
		}

		if a["target_cluster"] != "target" {
			return fmt.Errorf("expected to get a target_cluster from Aiven")
		}

		if a["enable"] != "true" {
			return fmt.Errorf("expected to get a correct enable from Aiven")
		}

		return nil
	}
}

func TestAccAivenMirrorMakerReplicationFlow_invalid_offset_syncs_topic_location(t *testing.T) {
	config := `
resource "aiven_mirrormaker_replication_flow" "foo" {
  project                             = "foo"
  service_name                        = "foo"
  source_cluster                      = "source"
  target_cluster                      = "target"
  enable                              = true
  replication_policy_class            = "org.apache.kafka.connect.mirror.IdentityReplicationPolicy"
  sync_group_offsets_enabled          = true
  sync_group_offsets_interval_seconds = 10
  emit_heartbeats_enabled             = true
  emit_backward_heartbeats_enabled    = true
  offset_syncs_topic_location         = "lol_offset"
}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`expected offset_syncs_topic_location to be one of`),
			},
		},
	})
}
