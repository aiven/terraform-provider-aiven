package topic_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/kafkatopicrepository"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenKafkaTopic(t *testing.T) {
	accEnabled, _ := strconv.ParseBool(os.Getenv(resource.EnvTfAcc))
	if !accEnabled {
		t.Skipf("Set '%s=1' to run this acceptance test", resource.EnvTfAcc)
	}

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	orgName := acc.OrganizationName()
	projectName := acc.ProjectName()
	kafkaName := acc.RandName("kafka")

	// Creates shared Kafka
	cleanup, err := acc.CreateTestService(
		t.Context(),
		projectName,
		kafkaName,
		acc.WithServiceType("kafka"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	require.NoError(t, err)
	defer cleanup()

	t.Run("basic", func(t *testing.T) {
		defer kafkatopicrepository.ForgetService(projectName, kafkaName)

		resourceName := "aiven_kafka_topic.foo"
		topic2ResourceName := "aiven_kafka_topic.topic2"
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaTopicResource(orgName, projectName, kafkaName, rName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenKafkaTopicAttributes("data.aiven_kafka_topic.topic"),
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", kafkaName),
						resource.TestCheckResourceAttr(resourceName, "topic_name", fmt.Sprintf("test-acc-topic-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "partitions", "3"),
						resource.TestCheckResourceAttr(resourceName, "replication", "2"),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
						resource.TestCheckResourceAttr(resourceName, "config.0.retention_bytes", "1234"),
						resource.TestCheckResourceAttr(resourceName, "config.0.segment_bytes", "1610612736"),
						resource.TestCheckResourceAttr(topic2ResourceName, "topic_description", fmt.Sprintf("test-acc-topic2-desc-%s", rName)),
						resource.TestCheckResourceAttrSet(topic2ResourceName, "owner_user_group_id"),
					),
				},
			},
		})
	})

	// The test proves that TF can create hundreds of topics without hitting any server issues
	t.Run("many_topics", func(t *testing.T) {
		defer kafkatopicrepository.ForgetService(projectName, kafkaName)

		topicName := acc.RandName("topic")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafka451TopicResource(projectName, kafkaName, topicName),
					Check: resource.ComposeTestCheckFunc(
						func(state *terraform.State) error {
							assert.Len(t, state.RootModule().Resources, 451)
							return nil
						},
					),
				},
			},
		})
	})

	// This test proves the "termination_protection" field doesn't mess up with the state in various scenarios,
	// despite being a virtual field (not sent to or returned from the API).
	t.Run("termination_protection", func(t *testing.T) {
		defer kafkatopicrepository.ForgetService(projectName, kafkaName)

		topicName := acc.RandName("topic")
		resourceName := "aiven_kafka_topic.foo"
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaTopicTerminationProtectionResource(projectName, kafkaName, topicName, false),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					),
				},
				{
					Config: testAccKafkaTopicTerminationProtectionResource(projectName, kafkaName, topicName, true),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: filepath.Join(projectName, kafkaName, topicName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
					),
				},
				{
					RefreshState: true,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
					),
				},
				{
					Config: testAccKafkaTopicTerminationProtectionResource(projectName, kafkaName, topicName, false),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					),
				},
			},
		})
	})

	// TestAccAivenKafkaTopic_recreate validates that topic is recreated if it is missing
	// Kafka loses all topics on turn off/on, then TF recreates topics. This test imitates the case.
	t.Run("recreate_missing", func(t *testing.T) {
		defer kafkatopicrepository.ForgetService(projectName, kafkaName)

		topicResource := "aiven_kafka_topic.topic"
		topicName := acc.RandName("topic")
		kafkaID := fmt.Sprintf("%s/%s", projectName, kafkaName)
		topicID := fmt.Sprintf("%s/%s", kafkaID, topicName)

		config := testAccAivenKafkaTopicResourceRecreateMissing(projectName, kafkaName, topicName)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					// Step 1: setups resources, creates the state
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(topicResource, "id", topicID),
					),
				},
				{
					// Step 2: deletes topic, then runs apply, same config & checks
					PreConfig: func() {
						ctx := t.Context()
						err := client.ServiceKafkaTopicDelete(ctx, projectName, kafkaName, topicName)
						require.NoError(t, err)

						// Makes sure topic does not exist
						err = retry.Do(
							func() error {
								_, err := client.ServiceKafkaTopicGet(ctx, projectName, kafkaName, topicName)
								return err
							},
							retry.Context(ctx),
							retry.Delay(5*time.Second),
							retry.Attempts(10),
							retry.RetryIf(avngen.IsNotFound),
							retry.LastErrorOnly(true),
						)
						require.NoError(t, err)

						// We need to remove it from reps cache
						assert.NoError(t, kafkatopicrepository.ForgetTopic(projectName, kafkaName, topicName))
					},
					// Now plan shows a diff
					ExpectNonEmptyPlan: true,
					RefreshState:       true,
				},
				{
					// Step 3: recreates the topic
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						// Saved in state
						resource.TestCheckResourceAttr(topicResource, "id", topicID),
						func(_ *terraform.State) error {
							// Topic exists and active
							ctx := t.Context()
							return retry.Do(
								func() error {
									topic, err := client.ServiceKafkaTopicGet(ctx, projectName, kafkaName, topicName)
									if topic != nil {
										assert.NotNil(t, "ACTIVE", topic.State)
									}
									return err
								},
								retry.Context(ctx),
								retry.Delay(5*time.Second),
								retry.Attempts(10),
								retry.RetryIf(avngen.IsNotFound),
								retry.LastErrorOnly(true),
							)
						},
					),
				},
			},
		})
	})

	t.Run("import_missing", func(t *testing.T) {
		topicName := acc.RandName("topic")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					// Tries to import non-existing topic
					// Must fail not create
					Config:        testAccAivenKafkaTopicResourceImportMissing(projectName, kafkaName, topicName),
					ResourceName:  "aiven_kafka_topic.topic",
					ImportState:   true,
					ImportStateId: fmt.Sprintf("%s/%s/%s", projectName, kafkaName, topicName),
					ExpectError:   regexp.MustCompile(`Topic not found`),
				},
			},
		})
	})

	t.Run("conflicts_if_exists", func(t *testing.T) {
		defer kafkatopicrepository.ForgetService(projectName, kafkaName)

		topicName := acc.RandName("topic")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAivenKafkaTopicConflictsIfExists(projectName, kafkaName, topicName),
					ExpectError: regexp.MustCompile(`Topic conflict, already exists`),
				},
			},
		})
	})

	// The Plugin Framework doesn't support computed+optional
	// This test proves that when the block is removed from the tf file, it disappears from the state too.
	t.Run("config_block_is removed", func(t *testing.T) {
		defer kafkatopicrepository.ForgetService(projectName, kafkaName)

		topicName := acc.RandName("topic")
		resourceName := "aiven_kafka_topic.foo"
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "aiven_kafka_topic" "foo" {
  project      = %q
  service_name = %q
  topic_name   = %q
  partitions   = 3
  replication  = 2

  config {
    flush_ms       = 10
    cleanup_policy = "compact"
  }
}
`, projectName, kafkaName, topicName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "config.#", "1"),
					),
				},
				{
					Config: fmt.Sprintf(`
resource "aiven_kafka_topic" "foo" {
  project      = %q
  service_name = %q
  topic_name   = %q
  partitions   = 3
  replication  = 2
}
`, projectName, kafkaName, topicName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "config.#", "0"),
					),
				},
			},
		})
	})
}

func testAccKafka451TopicResource(projectName, kafkaName, topicName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "foo" {
  count = 451

  project      = %q
  service_name = %q
  topic_name   = "%s-${count.index}"
  partitions   = 3
  replication  = 2
}`, projectName, kafkaName, topicName)
}

func testAccKafkaTopicResource(orgName, projectName, kafkaName, rName string) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = %[1]q
}

data "aiven_kafka" "bar" {
  project      = %[2]q
  service_name = %[3]q
}

resource "aiven_organization_user_group" "foo" {
  name            = "test-acc-u-grp-%[4]s"
  organization_id = data.aiven_organization.foo.id
  description     = "test"
}

resource "aiven_kafka_topic" "foo" {
  project      = data.aiven_kafka.bar.project
  service_name = data.aiven_kafka.bar.service_name
  topic_name   = "test-acc-topic-%[4]s"
  partitions   = 3
  replication  = 2

  config {
    flush_ms                  = 10
    cleanup_policy            = "compact"
    min_cleanable_dirty_ratio = 0.01
    delete_retention_ms       = 50000
    retention_bytes           = 1234
    segment_bytes             = 1610612736
  }
}

data "aiven_kafka_topic" "topic" {
  project      = aiven_kafka_topic.foo.project
  service_name = aiven_kafka_topic.foo.service_name
  topic_name   = aiven_kafka_topic.foo.topic_name

  depends_on = [aiven_kafka_topic.foo]
}

resource "aiven_kafka_topic" "topic2" {
  project             = data.aiven_kafka.bar.project
  service_name        = data.aiven_kafka.bar.service_name
  topic_name          = "test-acc-topic2-%[4]s"
  topic_description   = "test-acc-topic2-desc-%[4]s"
  owner_user_group_id = aiven_organization_user_group.foo.group_id
  partitions          = 3
  replication         = 2
}`, orgName, projectName, kafkaName, rName)
}

func testAccKafkaTopicTerminationProtectionResource(projectName, kafkaName, topicName string, protection bool) string {
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "foo" {
  project                = %[1]q
  service_name           = %[2]q
  topic_name             = %[3]q
  partitions             = 3
  replication            = 2
  termination_protection = %t
}
`, projectName, kafkaName, topicName, protection)
}

func testAccCheckAivenKafkaTopicAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] kafka topic attributes %v", a)

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["topic_name"] == "" {
			return fmt.Errorf("expected to get a topic_name from Aiven")
		}

		if a["partitions"] == "" {
			return fmt.Errorf("expected to get a partitions from Aiven")
		}

		if a["replication"] == "" {
			return fmt.Errorf("expected to get a replication from Aiven")
		}

		return nil
	}
}

func testAccCheckAivenKafkaTopicResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each created resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "aiven_kafka_topic" {
			project, serviceName, topicName, err := schemautil.SplitResourceID3(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = c.ServiceGet(ctx, project, serviceName)
			if err != nil {
				if avngen.IsNotFound(err) {
					return nil
				}
				return err
			}

			t, err := c.ServiceKafkaTopicGet(ctx, project, serviceName, topicName)
			if err != nil {
				if avngen.IsNotFound(err) {
					return nil
				}
				return err
			}

			if t != nil {
				return fmt.Errorf("kafka topic (%s) still exists, id %s", topicName, rs.Primary.ID)
			}
		}
		if rs.Type == "aiven_organization_user_group" {
			orgID, userGroupID, err := schemautil.SplitResourceID2(rs.Primary.ID)
			if err != nil {
				return err
			}

			r, err := c.UserGroupGet(ctx, orgID, userGroupID)
			if err != nil {
				if avngen.IsNotFound(err) {
					return nil
				}
				return err
			}

			if r != nil {
				return fmt.Errorf("organization user group (%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccAivenKafkaTopicResourceRecreateMissing(projectName, kafkaName, topicName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "topic" {
  project      = %[1]q
  service_name = %[2]q
  topic_name   = %[3]q
  partitions   = 5
  replication  = 2
}`, projectName, kafkaName, topicName)
}

func testAccAivenKafkaTopicResourceImportMissing(projectName, kafkaName, topicName string) string {
	result := fmt.Sprintf(`
resource "aiven_kafka_topic" "topic" {
  project      = %[1]q
  service_name = %[2]q
  topic_name   = %[3]q
  partitions   = 5
  replication  = 2
}
`, projectName, kafkaName, topicName)
	return result
}

func testAccAivenKafkaTopicConflictsIfExists(projectName, kafkaName, topicName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "topic" {
  project      = %[1]q
  service_name = %[2]q
  topic_name   = %[3]q
  partitions   = 5
  replication  = 2
}

resource "aiven_kafka_topic" "topic_conflict" {
  project      = %[1]q
  service_name = %[2]q
  topic_name   = %[3]q
  partitions   = 5
  replication  = 2

  depends_on = [
    aiven_kafka_topic.topic
  ]
}
`, projectName, kafkaName, topicName)
}

func TestAivenKafkaTopic_local_retention_bytes_overflow_error(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "aiven_kafka_topic" "topic" {
  project      = "foo"
  service_name = "foo"
  topic_name   = "foo"
  partitions   = 5
  replication  = 2

  config {
    local_retention_bytes = 3
    retention_bytes       = 1
  }
}`,
				ExpectError: regexp.MustCompile(`local_retention_bytes cannot be greater than retention_bytes`),
			},
		},
	})
}

func TestAivenKafkaTopic_local_retention_bytes_overflow_dependency(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "aiven_kafka_topic" "topic" {
  project      = "foo"
  service_name = "foo"
  topic_name   = "foo"
  partitions   = 5
  replication  = 2

  config {
    local_retention_bytes = 3
  }
}`,
				ExpectError: regexp.MustCompile(`Attribute "config\[0].retention_bytes" must be specified when\s+"config\[0].local_retention_bytes" is specified`),
			},
		},
	})
}
