package registryacl_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenKafkaSchemaRegistryACL(t *testing.T) {
	projectName := acc.ProjectName()
	serviceName := acc.RandName("kafka")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("kafka"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	t.Run("base test", func(t *testing.T) {
		resourceName := "aiven_kafka_schema_registry_acl.foo"
		datasourceName := "data.aiven_kafka_schema_registry_acl.foo"
		datasourceNameByPermission := "data.aiven_kafka_schema_registry_acl.bar"
		username := acc.RandName("user")
		topic := acc.RandName("topic")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaSchemaRegistryACLResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() { require.NoError(t, <-serviceIsReady) },
					Config:    testAccKafkaSchemaRegistryACL(projectName, serviceName, topic, username),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "resource", "Subject:"+topic),
						resource.TestCheckResourceAttr(resourceName, "username", username),
						resource.TestCheckResourceAttr(resourceName, "permission", "schema_registry_read"),
						resource.TestCheckResourceAttrSet(resourceName, "acl_id"),

						resource.TestCheckResourceAttr(datasourceName, "project", projectName),
						resource.TestCheckResourceAttr(datasourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(datasourceName, "resource", "Subject:"+topic),
						resource.TestCheckResourceAttr(datasourceName, "username", username),
						resource.TestCheckResourceAttr(datasourceName, "permission", "schema_registry_read"),
						resource.TestCheckResourceAttrPair(datasourceName, "acl_id", resourceName, "acl_id"),

						// The data source can be looked up by permission/resource/username instead of acl_id
						resource.TestCheckResourceAttrPair(datasourceNameByPermission, "acl_id", resourceName, "acl_id"),
					),
				},
				{
					Config:            testAccKafkaSchemaRegistryACL(projectName, serviceName, topic, username),
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("invalid permission", func(t *testing.T) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccKafkaSchemaRegistryACLWithPermission(projectName, serviceName, "user-1", "wrong_permission"),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Attribute permission value must be one of`),
				},
			},
		})
	})

	t.Run("wildcard usernames accepted", func(t *testing.T) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:             testAccKafkaSchemaRegistryACLWithPermission(projectName, serviceName, "*", "schema_registry_read"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
				{
					Config:             testAccKafkaSchemaRegistryACLWithPermission(projectName, serviceName, "group-user-*", "schema_registry_read"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
			},
		})
	})

	t.Run("backward compatibility", func(t *testing.T) {
		resourceName := "aiven_kafka_schema_registry_acl.foo"
		username := acc.RandName("user")
		topic := acc.RandName("topic")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				PreConfig: func() {
					require.NoError(t, <-serviceIsReady)
				},
				TFConfig:           testAccKafkaSchemaRegistryACLBackwardCompat(projectName, serviceName, topic, username),
				OldProviderVersion: "4.47.0",
				Checks: resource.ComposeTestCheckFunc(
					func(state *terraform.State) error {
						// Lets the BE to reach the eventual consistency before "refresh" and the step 2
						time.Sleep(5 * time.Second)
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "resource", "Subject:"+topic),
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "permission", "schema_registry_read"),
					resource.TestCheckResourceAttrSet(resourceName, "acl_id"),
				),
			}),
		})
	})
}

func testAccCheckAivenKafkaSchemaRegistryACLResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_kafka_schema_registry_acl" {
			continue
		}

		project, serviceName, aclID, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		acls, err := c.ServiceSchemaRegistryAclList(ctx, project, serviceName)
		if err != nil {
			var e avngen.Error
			if errors.As(err, &e) && e.Status != 404 {
				return err
			}
			continue
		}

		for _, acl := range acls {
			if acl.Id != nil && *acl.Id == aclID {
				return fmt.Errorf("kafka schema registry ACL (%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccKafkaSchemaRegistryACL(project, serviceName, topic, username string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "foo" {
  project      = %[1]q
  service_name = %[2]q
  topic_name   = %[3]q
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = %[1]q
  service_name = %[2]q
  resource     = "Subject:${aiven_kafka_topic.foo.topic_name}"
  username     = %[4]q
  permission   = "schema_registry_read"
}

data "aiven_kafka_schema_registry_acl" "foo" {
  project      = aiven_kafka_schema_registry_acl.foo.project
  service_name = aiven_kafka_schema_registry_acl.foo.service_name
  acl_id       = aiven_kafka_schema_registry_acl.foo.acl_id
}

data "aiven_kafka_schema_registry_acl" "bar" {
  project      = aiven_kafka_schema_registry_acl.foo.project
  service_name = aiven_kafka_schema_registry_acl.foo.service_name
  permission   = aiven_kafka_schema_registry_acl.foo.permission
  resource     = aiven_kafka_schema_registry_acl.foo.resource
  username     = aiven_kafka_schema_registry_acl.foo.username
}
`, project, serviceName, topic, username)
}

func testAccKafkaSchemaRegistryACLWithPermission(project, serviceName, username, permission string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = %[1]q
  service_name = %[2]q
  resource     = "Subject:topic-1"
  username     = %[3]q
  permission   = %[4]q
}
`, project, serviceName, username, permission)
}

func testAccKafkaSchemaRegistryACLBackwardCompat(project, serviceName, topic, username string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_topic" "foo" {
  project      = %[1]q
  service_name = %[2]q
  topic_name   = %[3]q
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = %[1]q
  service_name = %[2]q
  resource     = "Subject:${aiven_kafka_topic.foo.topic_name}"
  username     = %[4]q
  permission   = "schema_registry_read"
}
`, project, serviceName, topic, username)
}
