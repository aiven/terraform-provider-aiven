package clickhouse_test

import (
	"fmt"
	"os"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/service/clickhouse"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenClickhouseRole(t *testing.T) {
	serviceName := fmt.Sprintf("test-acc-ch-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	resourceName := "aiven_clickhouse_role.foo"

	manifest := fmt.Sprintf(`
			resource "aiven_clickhouse" "bar" {
			  project                 = "%s"
			  cloud_name              = "google-europe-west1"
			  plan                    = "startup-beta-8"
			  service_name            = "%s"
			  maintenance_window_dow  = "monday"
			  maintenance_window_time = "10:00:00"
			}
			
			resource "aiven_clickhouse_role" "foo" {
			  service_name = aiven_clickhouse.bar.service_name
			  project      = aiven_clickhouse.bar.project
			  role         = "writer"
			}`,
		projectName, serviceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenClickhouseRoleResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "role", "writer"),
				),
			},
		},
	})
}

func testAccCheckAivenClickhouseRoleResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each aiven_clickhouse_role is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_clickhouse_role" {
			continue
		}

		projectName, serviceName, roleName := schemautil.SplitResourceID3(rs.Primary.ID)
		if exists, err := clickhouse.RoleExists(c, projectName, serviceName, roleName); err != nil {
			if aiven.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("unable to check if role '%s' still exists: %w", roleName, err)
		} else if exists {
			return fmt.Errorf("role '%s' still exists", roleName)
		}
	}
	return nil
}
