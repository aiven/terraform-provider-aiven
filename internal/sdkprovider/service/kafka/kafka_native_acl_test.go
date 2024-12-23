package kafka_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestKafkaNativeAcl tests the kafka acl resource.
func TestKafkaNativeAcl(t *testing.T) {
	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	serviceName := fmt.Sprintf("test-acc-native-acl-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	resourceName := "aiven_kafka_native_acl.foo"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKafkaACLConfig(projectName, serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "project"),
					resource.TestCheckResourceAttrSet(resourceName, "service_name"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_name"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_type"),
					resource.TestCheckResourceAttrSet(resourceName, "pattern_type"),
					resource.TestCheckResourceAttrSet(resourceName, "principal"),
					resource.TestCheckResourceAttrSet(resourceName, "host"),
					resource.TestCheckResourceAttrSet(resourceName, "operation"),
					resource.TestCheckResourceAttrSet(resourceName, "permission_type"),
				),
			},
		},
	})
}

func testKafkaACLConfig(projectName string, serviceName string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
  }
}

resource "aiven_kafka_native_acl" "foo" {
  project         = data.aiven_project.foo.project
  service_name    = aiven_kafka.bar.service_name
  resource_name   = "name-test"
  resource_type   = "Topic"
  pattern_type    = "LITERAL"
  principal       = "User:alice"
  host            = "host-test"
  operation       = "Create"
  permission_type = "ALLOW"
}`, projectName, serviceName)
}
