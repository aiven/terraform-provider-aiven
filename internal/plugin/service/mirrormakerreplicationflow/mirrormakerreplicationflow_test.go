package mirrormakerreplicationflow_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenMirrorMakerReplicationFlow(t *testing.T) {
	projectName := acc.ProjectName()
	resourceName := "aiven_mirrormaker_replication_flow.foo"
	datasourceName := "data.aiven_mirrormaker_replication_flow.flow"

	// The two Kafka services and the MirrorMaker service are slow to provision and
	// the flow only needs them to exist, so they are created once and shared by all
	// subtests. Each subtest manages the integrations, topics and the flow itself.
	sourceName := acc.RandName("kafka-source")
	targetName := acc.RandName("kafka-target")
	mmName := acc.RandName("mm")

	servicesAreReady := []<-chan error{
		acc.CreateTestService(
			t,
			projectName,
			sourceName,
			acc.WithServiceType("kafka"),
			acc.WithPlan("startup-4"),
			acc.WithCloud("google-europe-west1"),
		),
		acc.CreateTestService(
			t,
			projectName,
			targetName,
			acc.WithServiceType("kafka"),
			acc.WithPlan("startup-4"),
			acc.WithCloud("google-europe-west1"),
		),
		acc.CreateTestService(
			t,
			projectName,
			mmName,
			acc.WithServiceType("kafka_mirrormaker"),
			acc.WithPlan("startup-4"),
			acc.WithCloud("google-europe-west1"),
		),
	}
	for _, ready := range servicesAreReady {
		require.NoError(t, <-ready)
	}

	t.Run("basic", func(t *testing.T) {
		topicName := acc.RandName("topic")

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenMirrorMakerReplicationFlowResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccMirrorMakerReplicationFlowResource(projectName, sourceName, targetName, mmName, topicName, ""),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", mmName),
						resource.TestCheckResourceAttr(resourceName, "source_cluster", "source"),
						resource.TestCheckResourceAttr(resourceName, "target_cluster", "target"),
						resource.TestCheckResourceAttr(resourceName, "enable", "true"),
						resource.TestCheckResourceAttr(resourceName, "offset_syncs_topic_location", "source"),
						resource.TestCheckResourceAttr(resourceName, "replication_factor", "2"),
						resource.TestCheckResourceAttr(resourceName, "topics_blacklist.#", "3"),
						resource.TestCheckResourceAttr(resourceName, "topics_blacklist.0", ".*[\\-\\.]internal"),
						resource.TestCheckResourceAttr(resourceName, "topics_blacklist.1", ".*\\.replica"),
						resource.TestCheckResourceAttr(resourceName, "topics_blacklist.2", "__.*"),
						// Never configured, so it is not sent and not stored: the
						// flow keeps the exclusions the API applies on its own.
						resource.TestCheckNoResourceAttr(resourceName, "config_properties_exclude.#"),
						resource.TestCheckResourceAttr(resourceName, "exactly_once_delivery_enabled", "true"),
						resource.TestCheckResourceAttr(resourceName, "follower_fetching_enabled", "true"),

						// Data source reads the same flow back by its composite key.
						resource.TestCheckResourceAttr(datasourceName, "source_cluster", "source"),
						resource.TestCheckResourceAttr(datasourceName, "target_cluster", "target"),
						resource.TestCheckResourceAttr(datasourceName, "enable", "true"),
					),
				},
				{
					Config: testAccMirrorMakerReplicationFlowResource(projectName, sourceName, targetName, mmName, topicName, `
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
					// Removing the config clears it on the backend: here the empty
					// string is sent, and the cleared value reads back as absent.
					Config: testAccMirrorMakerReplicationFlowResource(projectName, sourceName, targetName, mmName, topicName, ``),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckNoResourceAttr(resourceName, "config_properties_exclude.#"),
					),
				},
				{
					Config:            testAccMirrorMakerReplicationFlowResource(projectName, sourceName, targetName, mmName, topicName, ``),
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	// A configuration that omits every optional field must apply cleanly and leave
	// a follow-up plan empty. The API returns only the fields that were set on the
	// flow, so the omitted ones come back absent and stay null: the ".*" default
	// the spec documents for topics is MirrorMaker's runtime default, it is not
	// stored on the flow and never echoed back.
	t.Run("omitted_fields", func(t *testing.T) {
		topicName := acc.RandName("topic")

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenMirrorMakerReplicationFlowResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccMirrorMakerReplicationFlowMinimal(projectName, sourceName, targetName, mmName, topicName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "enable", "true"),
						// Absent in the response means null, not the empty list the
						// API treats as "no value": Terraform tells the two apart
						// and an empty list is a value the configuration never
						// asked for.
						resource.TestCheckNoResourceAttr(resourceName, "topics.#"),
					),
				},
				{
					// No drift once what the flow reports is in state.
					Config:   testAccMirrorMakerReplicationFlowMinimal(projectName, sourceName, targetName, mmName, topicName),
					PlanOnly: true,
				},
			},
		})
	})

	// Dropping a list from the configuration must clear it on the backend, the way
	// the SDK did. topics and topics_blacklist are the only fields here that can be
	// cleared at all: their request fields are `*[]string`, so the empty array the
	// adapter sends for a removed list survives `omitempty` and reaches the API.
	t.Run("removing_topics_clears_them", func(t *testing.T) {
		topicName := acc.RandName("topic")

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenMirrorMakerReplicationFlowResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccMirrorMakerReplicationFlowResource(projectName, sourceName, targetName, mmName, topicName, ""),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "topics.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "topics_blacklist.#", "3"),
					),
				},
				{
					// The same flow, now configured without the two lists.
					Config: testAccMirrorMakerReplicationFlowMinimal(projectName, sourceName, targetName, mmName, topicName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckNoResourceAttr(resourceName, "topics.#"),
						resource.TestCheckNoResourceAttr(resourceName, "topics_blacklist.#"),
					),
				},
			},
		})
	})

	// Verifies that state created by the previous SDK-based provider version is
	// compatible with the Plugin Framework version.
	t.Run("backward_compat", func(t *testing.T) {
		topicName := acc.RandName("topic")

		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acc.TestAccPreCheck(t) },
			CheckDestroy: testAccCheckAivenMirrorMakerReplicationFlowResourceDestroy,
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig: testAccMirrorMakerReplicationFlowResource(projectName, sourceName, targetName, mmName, topicName, ""),
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "source_cluster", "source"),
					resource.TestCheckResourceAttr(resourceName, "target_cluster", "target"),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "topics_blacklist.#", "3"),
				),
			}),
		})
	})

	// The same, for a configuration that leaves the optional lists out. Keeping
	// them plain optional only works if the SDK stored nothing for them either:
	// had it stored an empty list, the first plan after the upgrade would want to
	// turn that into null. The checks run against both providers, so the first
	// step pins down what the SDK actually wrote.
	t.Run("backward_compat_omitted_fields", func(t *testing.T) {
		topicName := acc.RandName("topic")

		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acc.TestAccPreCheck(t) },
			CheckDestroy: testAccCheckAivenMirrorMakerReplicationFlowResourceDestroy,
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig: testAccMirrorMakerReplicationFlowMinimal(projectName, sourceName, targetName, mmName, topicName),
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "topics.#"),
					resource.TestCheckNoResourceAttr(resourceName, "topics_blacklist.#"),
				),
			}),
		})
	})
}

// Validation only, so it needs no services and stays out of the shared-service test.
func TestAccAivenMirrorMakerReplicationFlow_invalid_offset_syncs_topic_location(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Attribute offset_syncs_topic_location value must be one of`),
				Config: `
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
`,
			},
		},
	})
}

func testAccCheckAivenMirrorMakerReplicationFlowResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("failed to instantiate GenAiven client: %w", err)
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_mirrormaker_replication_flow" {
			continue
		}

		project, serviceName, sourceCluster, targetCluster, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceKafkaMirrorMakerGetReplicationFlow(ctx, project, serviceName, sourceCluster, targetCluster)
		if err != nil {
			// The MirrorMaker service is shared and outlives the test, so a 404
			// here means the flow itself is gone.
			if avngen.IsNotFound(err) {
				continue
			}

			return err
		}

		return fmt.Errorf("kafka mirror maker replication flow (%s) still exists", rs.Primary.ID)
	}

	return nil
}

// testAccMirrorMakerReplicationFlowResource builds a config against the shared
// services: it only manages the topics, the cluster-alias integrations and the flow.
func testAccMirrorMakerReplicationFlowResource(project, sourceName, targetName, mmName, topicName, configExclude string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "source" {
  project      = %[1]q
  service_name = %[2]q
  topic_name   = "%[5]s-a"
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka_topic" "target" {
  project      = %[1]q
  service_name = %[3]q
  topic_name   = "%[5]s-b"
  partitions   = 3
  replication  = 2
}

resource "aiven_service_integration" "source" {
  project                  = %[1]q
  integration_type         = "kafka_mirrormaker"
  source_service_name      = %[2]q
  destination_service_name = %[4]q

  kafka_mirrormaker_user_config {
    cluster_alias = "source"
  }
}

resource "aiven_service_integration" "target" {
  project                  = %[1]q
  integration_type         = "kafka_mirrormaker"
  source_service_name      = %[3]q
  destination_service_name = %[4]q

  kafka_mirrormaker_user_config {
    cluster_alias = "target"
  }
}

resource "aiven_mirrormaker_replication_flow" "foo" {
  project                             = %[1]q
  service_name                        = %[4]q
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
  follower_fetching_enabled           = true

  topics = [
    ".*",
  ]

  topics_blacklist = [
    ".*[\\-\\.]internal",
    ".*\\.replica",
    "__.*"
  ]

  %[6]s

  # The integrations declare the "source" and "target" cluster aliases the flow
  # refers to. Nothing else orders them now that the services are not managed here.
  depends_on = [
    aiven_service_integration.source,
    aiven_service_integration.target,
  ]
}

data "aiven_mirrormaker_replication_flow" "flow" {
  project        = %[1]q
  service_name   = %[4]q
  source_cluster = aiven_mirrormaker_replication_flow.foo.source_cluster
  target_cluster = aiven_mirrormaker_replication_flow.foo.target_cluster

  depends_on = [aiven_mirrormaker_replication_flow.foo]
}`, project, sourceName, targetName, mmName, topicName, configExclude)
}

// testAccMirrorMakerReplicationFlowMinimal sets only the required fields and omits
// the ones the API leaves unset when they are not sent (topics, topics_blacklist,
// replication_factor), so the Optional+Computed handling can be exercised.
func testAccMirrorMakerReplicationFlowMinimal(project, sourceName, targetName, mmName, topicName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "source" {
  project      = %[1]q
  service_name = %[2]q
  topic_name   = "%[5]s-a"
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka_topic" "target" {
  project      = %[1]q
  service_name = %[3]q
  topic_name   = "%[5]s-b"
  partitions   = 3
  replication  = 2
}

resource "aiven_service_integration" "source" {
  project                  = %[1]q
  integration_type         = "kafka_mirrormaker"
  source_service_name      = %[2]q
  destination_service_name = %[4]q

  kafka_mirrormaker_user_config {
    cluster_alias = "source"
  }
}

resource "aiven_service_integration" "target" {
  project                  = %[1]q
  integration_type         = "kafka_mirrormaker"
  source_service_name      = %[3]q
  destination_service_name = %[4]q

  kafka_mirrormaker_user_config {
    cluster_alias = "target"
  }
}

resource "aiven_mirrormaker_replication_flow" "foo" {
  project                     = %[1]q
  service_name                = %[4]q
  source_cluster              = "source"
  target_cluster              = "target"
  enable                      = true
  replication_policy_class    = "org.apache.kafka.connect.mirror.IdentityReplicationPolicy"
  offset_syncs_topic_location = "source"

  depends_on = [
    aiven_service_integration.source,
    aiven_service_integration.target,
  ]
}`, project, sourceName, targetName, mmName, topicName)
}
