package topic_test

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	kafkatopichandler "github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/kafkatopicrepository"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenKafkaTopic(t *testing.T) {
	orgName := acc.OrganizationName()
	projectName := acc.ProjectName()
	kafkaName := acc.RandName("kafka")

	// Largest int64 — used as the value for `message_timestamp_*_max_ms` config
	// fields to verify that the int64 schema can round-trip the full range without
	// overflowing into a different (unsigned) type during HCL → API marshalling.
	maxInt64 := fmt.Sprint(math.MaxInt64)

	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		kafkaName,
		acc.WithServiceType("kafka"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)
	require.NoError(t, <-serviceIsReady)

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	ctx := context.Background()
	list, err := client.ServiceKafkaTopicList(ctx, projectName, kafkaName)
	require.NoError(t, err)
	require.Empty(t, list, "This test assumes there are no pre-existing topics in the Kafka service")

	t.Run("basic", func(t *testing.T) {
		resourceName := "aiven_kafka_topic.foo"
		stringsResourceName := "aiven_kafka_topic.foo_strings"
		topic2ResourceName := "aiven_kafka_topic.topic2"
		topicName := acc.RandName("topic")
		topicStringsName := acc.RandName("topic-strs")
		topic2Name := acc.RandName("topic2")
		topic2Desc := acc.RandName("topic2-desc")
		userGroupName := acc.RandName("u-grp")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaTopicResource(orgName, projectName, kafkaName, topicName, topicStringsName, topic2Name, topic2Desc, userGroupName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenKafkaTopicAttributes("data.aiven_kafka_topic.topic"),
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", kafkaName),
						resource.TestCheckResourceAttr(resourceName, "topic_name", topicName),
						resource.TestCheckResourceAttr(resourceName, "partitions", "3"),
						resource.TestCheckResourceAttr(resourceName, "replication", "2"),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
						resource.TestCheckResourceAttr(resourceName, "config.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "config.0.retention_bytes", "1234"),
						resource.TestCheckResourceAttr(resourceName, "config.0.segment_bytes", "1610612736"),
						resource.TestCheckResourceAttr(resourceName, "config.0.message_timestamp_after_max_ms", maxInt64),
						resource.TestCheckResourceAttr(resourceName, "config.0.message_timestamp_before_max_ms", maxInt64),

						// `tag` is a Set in the schema, so iteration order isn't stable —
						// check the cardinality with `tag.#` and assert each element with
						// the order-independent TestCheckTypeSetElemNestedAttrs helper.
						resource.TestCheckResourceAttr(resourceName, "tag.#", "2"),
						resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
							"key":   "environment",
							"value": "test",
						}),
						resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
							"key":   "owner",
							"value": "tf-acc",
						}),

						resource.TestCheckResourceAttr(topic2ResourceName, "topic_description", topic2Desc),
						resource.TestCheckResourceAttrSet(topic2ResourceName, "owner_user_group_id"),

						// Topic config is not set when it is not defined by user
						resource.TestCheckResourceAttr(topic2ResourceName, "config.#", "0"),
						// topic2 has no tags configured either
						resource.TestCheckResourceAttr(topic2ResourceName, "tag.#", "0"),

						// The "_strings" clone proves that a config written with quoted
						// numeric literals (the SDKv2-era shape) round-trips into the
						// int64 schema without drift. If Terraform failed to coerce the
						// strings, these checks would diff against the apply result.
						resource.TestCheckResourceAttr(stringsResourceName, "topic_name", topicStringsName),
						resource.TestCheckResourceAttr(stringsResourceName, "partitions", "3"),
						resource.TestCheckResourceAttr(stringsResourceName, "replication", "2"),
						resource.TestCheckResourceAttr(stringsResourceName, "config.0.retention_bytes", "1234"),
						resource.TestCheckResourceAttr(stringsResourceName, "config.0.segment_bytes", "1610612736"),
						resource.TestCheckResourceAttr(stringsResourceName, "config.0.message_timestamp_after_max_ms", maxInt64),
						resource.TestCheckResourceAttr(stringsResourceName, "config.0.message_timestamp_before_max_ms", maxInt64),
					),
				},
			},
		})
	})

	// Proves that TF can create hundreds of topics without hitting any server issues.
	t.Run("many_topics", func(t *testing.T) {
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

	// termination_protection is now backed by the framework's built-in plan-time
	// deletion guard (resource.terminationProtection in the YAML). Setting it to
	// true and planning a destroy must fail; updating it to true without a destroy
	// must be a no-op apply.
	t.Run("termination_protection", func(t *testing.T) {
		topicName := acc.RandName("topic")
		resourceName := "aiven_kafka_topic.foo"
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config:                    testAccKafkaTopicTerminationProtectionResource(projectName, kafkaName, topicName),
					PreventPostDestroyRefresh: true,
					ExpectNonEmptyPlan:        true,
					PlanOnly:                  true,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", kafkaName),
						resource.TestCheckResourceAttr(resourceName, "topic_name", topicName),
						resource.TestCheckResourceAttr(resourceName, "partitions", "3"),
						resource.TestCheckResourceAttr(resourceName, "replication", "2"),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
					),
				},
			},
		})
	})

	// Validates that the topic is recreated if it is missing.
	// Kafka loses all topics on power-cycle without backup; refresh + apply should
	// recreate the topic without the user having to clean state up manually.
	t.Run("recreate_missing", func(t *testing.T) {
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
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(topicResource, "id", topicID),
					),
				},
				{
					PreConfig: func() {
						ctx := t.Context()
						require.NoError(t, err)

						err = retry.Do(
							func() error {
								topic, err := client.ServiceKafkaTopicGet(ctx, projectName, kafkaName, topicName)
								if topic != nil {
									assert.Equal(t, kafkatopichandler.TopicStateTypeActive, topic.State)
								}
								return err
							},
							retry.Context(ctx),
							retry.Delay(5*time.Second),
							retry.Attempts(20),
							retry.RetryIf(avngen.IsNotFound),
							retry.LastErrorOnly(true),
						)
						require.NoError(t, err)

						// Forget the cached read so the next refresh hits the live API.
						assert.NoError(t, kafkatopicrepository.ForgetTopic(projectName, kafkaName, topicName))
					},
					ExpectNonEmptyPlan: true,
					RefreshState:       true,
				},
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(topicResource, "id", topicID),
						func(_ *terraform.State) error {
							ctx := t.Context()
							return retry.Do(
								func() error {
									topic, err := client.ServiceKafkaTopicGet(ctx, projectName, kafkaName, topicName)
									if topic != nil {
										assert.Equal(t, kafkatopichandler.TopicStateTypeActive, topic.State)
									}
									return err
								},
								retry.Context(ctx),
								retry.Delay(5*time.Second),
								retry.Attempts(20),
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
					Config:        testAccAivenKafkaTopicResourceImportMissing(projectName, kafkaName, topicName),
					ResourceName:  "aiven_kafka_topic.topic",
					ImportState:   true,
					ImportStateId: fmt.Sprintf("%s/%s/%s", projectName, kafkaName, topicName),
					ExpectError:   regexp.MustCompile(`Cannot import non-existent remote object`),
				},
			},
		})
	})

	t.Run("conflicts_if_exists", func(t *testing.T) {
		topicName := acc.RandName("topic")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAivenKafkaTopicConflictsIfExists(projectName, kafkaName, topicName),
					ExpectError: regexp.MustCompile(fmt.Sprintf(`topic conflict, already exists: %q`, topicName)),
				},
			},
		})
	})

	// Verifies that the data source surfaces a missing topic as an error.
	// Unlike the resource path (which has removeMissing: true and drops the resource
	// from state so that the next apply can create it), a data source must fail loudly
	// when the topic does not exist.
	t.Run("datasource_doesnt_exist", func(t *testing.T) {
		config := fmt.Sprintf(`
data "aiven_kafka_topic" "whatever" {
  project      = %[1]q
  service_name = %[2]q
  topic_name   = "whatever"
}`, projectName, kafkaName)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`not found`),
				},
			},
		})
	})

	// Verifies that a Kafka topic created with the last published provider version
	// (which modelled numeric config fields as `string`) can be re-applied by the
	// current dev version (which models them as `int64` directly from the OpenAPI
	// spec). The HCL is kept identical across the two steps — Terraform auto-converts
	// the string literals in `config { retention_ms = "604800000" ... }` to int64.
	//
	// The old provider only creates the topic against the shared Kafka service; it
	// never tries to manage the Kafka resource itself, so we don't need to import or
	// re-create it across the provider switch.
	t.Run("backward_compat", func(t *testing.T) {
		// This version can handle max int64 values for config fields.
		const oldVersion = "4.56.0"

		topicName := acc.RandName("topic")
		config := testAccAivenKafkaTopicBackwardCompatConfig(projectName, kafkaName, topicName)

		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acc.TestAccPreCheck(t) },
			CheckDestroy: testAccCheckAivenKafkaTopicResourceDestroy,
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig:           config,
				OldProviderVersion: oldVersion,
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aiven_kafka_topic.test", "id"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "topic_name", topicName),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "partitions", "3"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "replication", "2"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "topic_description", "Test topic for backward compatibility"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "termination_protection", "false"),

					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "tag.#", "2"),

					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.cleanup_policy", "delete"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.compression_type", "producer"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.retention_ms", "604800000"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.retention_bytes", "1073741824"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.min_insync_replicas", "1"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.max_message_bytes", "1048576"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.message_timestamp_type", "CreateTime"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.unclean_leader_election_enable", "false"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.preallocate", "false"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.segment_bytes", "536870912"),
					resource.TestCheckResourceAttr("aiven_kafka_topic.test", "config.0.message_timestamp_after_max_ms", maxInt64),

					// API eventual consistency: give the topic a moment to settle so the
					// step-2 plan does not flag a transient drift on the just-created topic.
					func(_ *terraform.State) error {
						time.Sleep(10 * time.Second)
						return nil
					},
				),
			}),
		})
	})
}

// testAccAivenKafkaTopicBackwardCompatConfig deliberately uses string literals for
// numeric `config` attributes to mimic HCL written for the old (SDKv2) provider where
// those fields were typed as `string`.
func testAccAivenKafkaTopicBackwardCompatConfig(projectName, kafkaName, topicName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "test" {
  project                = %[1]q
  service_name           = %[2]q
  topic_name             = %[3]q
  partitions             = 3
  replication            = 2
  topic_description      = "Test topic for backward compatibility"
  termination_protection = false

  tag {
    key   = "environment"
    value = "test"
  }

  tag {
    key   = "purpose"
    value = "backward-compat-testing"
  }

  config {
    cleanup_policy                 = "delete"
    compression_type               = "producer"
    retention_ms                   = "604800000"
    retention_bytes                = "1073741824"
    min_insync_replicas            = "1"
    max_message_bytes              = "1048576"
    message_timestamp_type         = "CreateTime"
    unclean_leader_election_enable = false
    preallocate                    = false
    segment_bytes                  = "536870912"
    message_timestamp_after_max_ms = "9223372036854775807"
  }
}`, projectName, kafkaName, topicName)
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

func testAccKafkaTopicResource(orgName, projectName, kafkaName, topicName, topicStringsName, topic2Name, topic2Desc, userGroupName string) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = %[1]q
}

data "aiven_kafka" "bar" {
  project      = %[2]q
  service_name = %[3]q
}

resource "aiven_organization_user_group" "foo" {
  name            = %[8]q
  organization_id = data.aiven_organization.foo.id
  description     = "test"
}

resource "aiven_kafka_topic" "foo" {
  project      = data.aiven_kafka.bar.project
  service_name = data.aiven_kafka.bar.service_name
  topic_name   = %[4]q
  partitions   = 3
  replication  = 2

  tag {
    key   = "environment"
    value = "test"
  }

  tag {
    key   = "owner"
    value = "tf-acc"
  }

  config {
    flush_ms                        = 10
    cleanup_policy                  = "compact"
    min_cleanable_dirty_ratio       = 0.01
    delete_retention_ms             = 50000
    retention_bytes                 = 1234
    segment_bytes                   = 1610612736
    message_timestamp_after_max_ms  = 9223372036854775807
    message_timestamp_before_max_ms = 9223372036854775807
  }
}

# foo_strings mirrors foo but writes every numeric value as a quoted HCL string.
# The schema is int64/float64, so Terraform must coerce each literal at parse
# time. This guards the implicit upgrade path from the SDKv2 provider — where these
# fields used to be typed as string — so that an existing HCL config keeps working
# unchanged after switching to the Plugin Framework implementation.
resource "aiven_kafka_topic" "foo_strings" {
  project      = data.aiven_kafka.bar.project
  service_name = data.aiven_kafka.bar.service_name
  topic_name   = %[5]q
  partitions   = "3"
  replication  = "2"

  config {
    flush_ms                        = "10"
    cleanup_policy                  = "compact"
    min_cleanable_dirty_ratio       = "0.01"
    delete_retention_ms             = "50000"
    retention_bytes                 = "1234"
    segment_bytes                   = "1610612736"
    message_timestamp_after_max_ms  = "9223372036854775807"
    message_timestamp_before_max_ms = "9223372036854775807"
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
  topic_name          = %[6]q
  topic_description   = %[7]q
  owner_user_group_id = aiven_organization_user_group.foo.group_id
  partitions          = 3
  replication         = 2
}`, orgName, projectName, kafkaName, topicName, topicStringsName, topic2Name, topic2Desc, userGroupName)
}

func testAccKafkaTopicTerminationProtectionResource(projectName, kafkaName, topicName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "foo" {
  project                = %[1]q
  service_name           = %[2]q
  topic_name             = %[3]q
  partitions             = 3
  replication            = 2
  termination_protection = true
}

data "aiven_kafka_topic" "topic" {
  project      = aiven_kafka_topic.foo.project
  service_name = aiven_kafka_topic.foo.service_name
  topic_name   = aiven_kafka_topic.foo.topic_name

  depends_on = [aiven_kafka_topic.foo]
}`, projectName, kafkaName, topicName)
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	// There is a test that creates hundreds of topics, so we cache topic lists per service.
	// Instead of checking each topic individually, we check the whole topic list is empty.
	topicListEmpty := make(map[string]bool)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_kafka_topic" {
			continue
		}

		project, serviceName, _, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		serviceKey := fmt.Sprintf("%s/%s", project, serviceName)
		if topicListEmpty[serviceKey] {
			continue
		}

		err = retry.Do(func() error {
			topics, err := c.ServiceKafkaTopicList(ctx, project, serviceName)
			if avngen.IsNotFound(err) || len(topics) == 0 {
				topicListEmpty[serviceKey] = true
				kafkatopicrepository.ForgetService(project, serviceName)
				return nil
			}

			if len(topics) != 0 {
				return fmt.Errorf("kafka topic list for service %s is not empty", serviceKey)
			}

			return err
		}, retry.Context(ctx), retry.Delay(5*time.Second))
		if err != nil {
			return err
		}
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_organization_user_group" {
			continue
		}

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
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "topic" {
  project      = %[1]q
  service_name = %[2]q
  topic_name   = %[3]q
  partitions   = 5
  replication  = 2
}
`, projectName, kafkaName, topicName)
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

func TestAccAivenKafkaTopic_config_max_items(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
resource "aiven_kafka_topic" "topic" {
  project      = "foo"
  service_name = "bar"
  topic_name   = "foo"
  partitions   = 5
  replication  = 2

  config {
    retention_ms = 1000
  }

  config {
    segment_ms = 2000
  }
}`,
				ExpectError: regexp.MustCompile(`(?s)((config|Config).*(at most 1|No more than 1)|(at most 1|No more than 1).*(config|Config))`),
			},
		},
	})
}

func TestAccAivenKafkaTopic_local_retention_bytes_validation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
resource "aiven_kafka_topic" "topic" {
  project      = "foo"
  service_name = "bar"
  topic_name   = "foo"
  partitions   = 5
  replication  = 2

  config {
    local_retention_bytes = 3
    retention_bytes       = 1
  }
}`,
				ExpectError: regexp.MustCompile(`local_retention_bytes must not be more than retention_bytes value`),
			},
			{
				// retention_bytes = -1 means Infinite
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config: `
resource "aiven_kafka_topic" "topic" {
  project      = "foo"
  service_name = "bar"
  topic_name   = "foo"
  partitions   = 5
  replication  = 2

  config {
    local_retention_bytes = 15000000000
    retention_bytes       = -1
  }
}`,
			},
		},
	})
}

func TestAccAivenKafkaTopic_local_retention_bytes_overflow_dependency(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config: `
resource "aiven_kafka_topic" "topic" {
  project      = "foo"
  service_name = "bar"
  topic_name   = "foo"
  partitions   = 5
  replication  = 2

  config {
    local_retention_bytes = 3
  }

}`,
				// AlsoRequires validator: "Attribute X must be specified when ..."
				ExpectError: regexp.MustCompile(`retention_bytes.*must be specified when`),
			},
		},
	})
}
