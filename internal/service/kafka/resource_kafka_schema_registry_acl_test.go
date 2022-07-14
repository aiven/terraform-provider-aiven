package kafka_test

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/aiven/aiven-go-client"
	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenKafkaSchemaRegistryACL_basic(t *testing.T) {
	resourceName := "aiven_kafka_schema_registry_acl.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenKafkaSchemaRegistryACLResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccKafkaSchemaRegistryACLWrongProjectResource(rName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("invalid value for project"),
			},
			{
				Config:      testAccKafkaSchemaRegistryACLWrongServiceNameResource(rName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("invalid value for service_name"),
			},
			{
				Config:      testAccKafkaSchemaRegistryACLWrongPermissionResource(rName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("expected permission to be one of"),
			},
			{
				Config:      testAccKafkaSchemaRegistryACLWrongUsernameResource(rName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("invalid value for username"),
			},
			{
				Config:      testAccKafkaSchemaRegistryACLInvalidCharsResource(rName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("invalid value for username"),
			},
			{
				Config:             testAccKafkaSchemaRegistryACLPrefixWildcardResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testAccKafkaSchemaRegistryACLWildcardResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccKafkaSchemaRegistryACLResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenKafkaSchemaRegistryACLAttributes("data.aiven_kafka_schema_registry_acl.acl"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "resource", fmt.Sprintf("Subject:test-acc-topic-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "permission", "schema_registry_read"),
					resource.TestCheckResourceAttrSet(resourceName, "acl_id"),
				),
			},
		},
	})
}

func testAccCheckAivenKafkaSchemaRegistryACLAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] kafka schema registry acl attributes %v", a)

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service name from Aiven")
		}

		if a["resource"] == "" {
			return fmt.Errorf("expected to get a resource from Aiven")
		}

		if a["username"] == "" {
			return fmt.Errorf("expected to get a username from Aiven")
		}

		if a["permission"] == "" {
			return fmt.Errorf("expected to get a permission from Aiven")
		}

		return nil
	}
}

func testAccKafkaSchemaRegistryACLWrongProjectResource(_ string) string {
	return `
resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = "!./,£$^&*()_"
  service_name = "test-acc-sr-1"
  resource     = "Subject:topic-1"
  username     = "user-1"
  permission   = "schema_registry_read"
}`

}

func testAccKafkaSchemaRegistryACLWrongServiceNameResource(_ string) string {
	return `
resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = "test-acc-pr-1"
  service_name = "!./,£$^&*()_"
  resource     = "Subject:topic-1"
  username     = "user-1"
  permission   = "schema_registry_read"
}`

}

func testAccKafkaSchemaRegistryACLWrongPermissionResource(_ string) string {
	return `
resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = "test-acc-pr-1"
  service_name = "test-acc-sr-1"
  resource     = "Subject:topic-1"
  username     = "user-1"
  permission   = "wrong-permission"
}`

}

func testAccKafkaSchemaRegistryACLWildcardResource(_ string) string {
	return `
resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = "test-acc-pr-1"
  service_name = "test-acc-sr-1"
  resource     = "Subject:topic-1"
  username     = "*"
  permission   = "schema_registry_read"
}`

}

func testAccKafkaSchemaRegistryACLPrefixWildcardResource(_ string) string {
	return `
resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = "test-acc-pr-1"
  service_name = "test-acc-sr-1"
  resource     = "Subject:topic-1"
  username     = "group-user-*"
  permission   = "schema_registry_read"
}`

}

func testAccKafkaSchemaRegistryACLWrongUsernameResource(_ string) string {
	return `
resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = "test-acc-pr-1"
  service_name = "test-acc-sr-1"
  resource     = "Subject:topic-1"
  username     = "#user"
  permission   = "schema_registry_read"
}`

}

func testAccKafkaSchemaRegistryACLInvalidCharsResource(_ string) string {
	return `
resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = "test-acc-pr-1"
  service_name = "test-acc-sr-1"
  resource     = "Subject:topic-1"
  username     = "!./,£$^&*()_"
  permission   = "schema_registry_read"
}`

}

func testAccKafkaSchemaRegistryACLResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

resource "aiven_kafka_topic" "foo" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.bar.service_name
  topic_name   = "test-acc-topic-%s"
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka_schema_registry_acl" "foo" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.bar.service_name
  resource     = "Subject:${aiven_kafka_topic.foo.topic_name}"
  username     = "user-%s"
  permission   = "schema_registry_read"
}

data "aiven_kafka_schema_registry_acl" "acl" {
  project      = aiven_kafka_schema_registry_acl.foo.project
  service_name = aiven_kafka_schema_registry_acl.foo.service_name
  resource     = aiven_kafka_schema_registry_acl.foo.resource
  username     = aiven_kafka_schema_registry_acl.foo.username
  permission   = aiven_kafka_schema_registry_acl.foo.permission

  depends_on = [aiven_kafka_schema_registry_acl.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name, name)
}

func testAccCheckAivenKafkaSchemaRegistryACLResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each kafka ACL is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_kafka_schema_registry_acl" {
			continue
		}

		project, serviceName, aclID := schemautil.SplitResourceID3(rs.Primary.ID)
		p, err := c.KafkaSchemaRegistryACLs.Get(project, serviceName, aclID)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("kafka Schema Registry ACL (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
