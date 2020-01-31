package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"os"
	"testing"
)

func TestAccAivenKafkaACL_basic(t *testing.T) {
	t.Parallel()

	resourceName := "aiven_kafka_acl.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenKafkaACLResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaACLResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenKafkaACLAttributes("data.aiven_kafka_acl.acl"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "topic", fmt.Sprintf("test-acc-topic-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "permission", "admin"),
				),
			},
		},
	})
}

func testAccCheckAivenKafkaACLAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] kafka acl attributes %v", a)

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service name from Aiven")
		}

		if a["topic"] == "" {
			return fmt.Errorf("expected to get a topic from Aiven")
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

func testAccKafkaACLResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
			card_id="%s"	
		}

		resource "aiven_service" "bar" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "business-4"
			service_name = "test-acc-sr-%s"
			service_type = "kafka"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			kafka_user_config {
				kafka_version = "2.4"
				kafka {
				  group_max_session_timeout_ms = 70000
				  log_retention_bytes = 1000000000
				}
			}
		}
		
		resource "aiven_kafka_topic" "foo" {
			project = aiven_project.foo.project
			service_name = aiven_service.bar.service_name
			topic_name = "test-acc-topic-%s"
			partitions = 3
			replication = 2
		}

		resource "aiven_kafka_acl" "foo" {
			project = aiven_project.foo.project
			service_name = aiven_service.bar.service_name
			topic = aiven_kafka_topic.foo.topic_name
			username = "user-%s"
			permission = "admin"
		}
		
		data "aiven_kafka_acl" "acl" {
			project = aiven_kafka_acl.foo.project
			service_name = aiven_kafka_acl.foo.service_name
			topic = aiven_kafka_acl.foo.topic
			username = aiven_kafka_acl.foo.username
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name, name, name)
}

func testAccCheckAivenKafkaACLResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each kafka ACL is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_kafka_acl" {
			continue
		}

		project, serviceName, aclID := splitResourceID3(rs.Primary.ID)
		p, err := c.KafkaACLs.Get(project, serviceName, aclID)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("kafka ACL (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
