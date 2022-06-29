package clickhouse_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aiven/terraform-provider-aiven/internal/service/clickhouse"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenClickhouseGrant(t *testing.T) {
	serviceName := fmt.Sprintf("test-acc-ch-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	projectName := os.Getenv("AIVEN_PROJECT_NAME")

	manifest := fmt.Sprintf(`
resource "aiven_clickhouse" "bar" {
  project                 = "%s"
  cloud_name              = "google-europe-west1"
  plan                    = "startup-beta-8"
  service_name            = "%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_clickhouse_database" "testdb" {
  project      = aiven_clickhouse.bar.project
  service_name = aiven_clickhouse.bar.service_name
  name         = "test-db"
}

resource "aiven_clickhouse_role" "foo-role" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  role         = "foo-role"
}

resource "aiven_clickhouse_grant" "foo-role-grant" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  role         = aiven_clickhouse_role.foo-role.role

  privilege_grant {
    privilege = "INSERT"
    database  = aiven_clickhouse_database.testdb.name
    table     = "test-table"
    column    = "test-column"
  }
}

resource "aiven_clickhouse_user" "foo-user" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  username     = "foo-user"
}

resource "aiven_clickhouse_grant" "foo-user-grant" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  user         = aiven_clickhouse_user.foo-user.username

  role_grant {
    role = aiven_clickhouse_role.foo-role.role
  }
}`, projectName, serviceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenClickhouseGrantResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					// privilege grant checks
					resource.TestCheckResourceAttr("aiven_clickhouse_grant.foo-role-grant", "privilege_grant.0.privilege", "INSERT"),
					resource.TestCheckResourceAttr("aiven_clickhouse_grant.foo-role-grant", "privilege_grant.0.database", "test-db"),
					resource.TestCheckResourceAttr("aiven_clickhouse_grant.foo-role-grant", "privilege_grant.0.table", "test-table"),
					resource.TestCheckResourceAttr("aiven_clickhouse_grant.foo-role-grant", "privilege_grant.0.column", "test-column"),

					// role grant checks
					resource.TestCheckResourceAttr("aiven_clickhouse_grant.foo-user-grant", "role_grant.0.role", "foo-role"),
				),
			},
		},
	})
}

func testAccCheckAivenClickhouseGrantResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each aiven_clickhouse_role is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_clickhouse_grant" {
			continue
		}

		projectName, serviceName, granteeType, granteeName := schemautil.SplitResourceID4(rs.Primary.ID)

		grantee := clickhouse.Grantee{}
		switch granteeType {
		case clickhouse.GranteeTypeRole:
			grantee.Role = granteeName
		case clickhouse.GranteeTypeUser:
			grantee.User = granteeName
		}

		if privilegeGrants, err := clickhouse.ReadPrivilegeGrants(c, projectName, serviceName, grantee); err != nil {
			if aiven.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("unable to check if privilege grants for '%s' still exists: %w", granteeName, err)
		} else if len(privilegeGrants) > 0 {
			return fmt.Errorf("'%s' still has privilege grants exists", granteeName)
		}

		if roleGrants, err := clickhouse.ReadRoleGrants(c, projectName, serviceName, grantee); err != nil {
			if aiven.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("unable to check if privilege grants for '%s' still exists: %w", granteeName, err)
		} else if len(roleGrants) > 0 {
			return fmt.Errorf("'%s' still has privilege grants exists", granteeName)
		}
	}
	return nil
}
