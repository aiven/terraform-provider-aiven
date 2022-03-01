package clickhouse_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aiven/terraform-provider-aiven/internal/service"

	"github.com/aiven/aiven-go-client"
	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/service/clickhouse/chsql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenClickhouseRole(t *testing.T) {
	t.Parallel()

	t.Run("clickhouse role creation", func(tt *testing.T) {
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

		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { acc.TestAccPreCheck(tt) },
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
		query, _ := chsql.ShowCreateRoleStatement(roleName)

		p, err := c.ClickHouseQuery.Query(projectName, serviceName, chsql.DefaultDatabaseForRoles, query)
		if err != nil {
			if !service.IsUnknownResource(err) {
				return fmt.Errorf("unable to query clickhouse for roles: %w", err)
			}
			return nil
		}

		if len(p.Data) > 0 {
			return fmt.Errorf("clickhouse role (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}
