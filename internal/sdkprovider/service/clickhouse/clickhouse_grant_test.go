package clickhouse_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	clickhouse2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/clickhouse"
)

func TestAccAivenClickhouseGrant(t *testing.T) {
	serviceName := fmt.Sprintf("test-acc-ch-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	projectName := acc.ProjectName()

	baseConfig := fmt.Sprintf(`
resource "aiven_clickhouse" "bar" {
  project                 = "%s"
  cloud_name              = "google-europe-west1"
  plan                    = "startup-8"
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

resource "aiven_clickhouse_user" "foo-user" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  username     = "foo-user"
}`, projectName, serviceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenClickhouseGrantResourceDestroy,
		Steps: []resource.TestStep{
			{
				// Step 1: Test role grant with a single privilege
				Config: baseConfig + `
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
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-role-grant",
						"privilege_grant.*",
						map[string]string{
							"privilege": "INSERT",
							"database":  "test-db",
							"table":     "test-table",
							"column":    "test-column",
						},
					),
				),
			},
			{
				// Step 2: Test case sensitivity with privilege names. The correct privilege name is "dictGet"
				Config: baseConfig + `
resource "aiven_clickhouse_grant" "foo-role-grant" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  role         = aiven_clickhouse_role.foo-role.role

  privilege_grant {
    privilege = "DICTGET"
    database  = "default"
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-role-grant",
						"privilege_grant.*",
						map[string]string{
							"privilege": "DICTGET", // the privilege name is stored as is
							"database":  "default",
						},
					),
				),
			},
			{
				// Step 3: Test case sensitivity with privilege names. The correct privilege name is "dictGet"
				Config: baseConfig + `
resource "aiven_clickhouse_grant" "foo-role-grant" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  role         = aiven_clickhouse_role.foo-role.role

  privilege_grant {
    privilege = "dictGet"
    database  = "default"
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-role-grant",
						"privilege_grant.*",
						map[string]string{
							"privilege": "dictGet", // the privilege name is stored as is
							"database":  "default",
						},
					),
				),
			},
			{
				// Step 4: Test privileges with asterisk (*)
				Config: baseConfig + `
resource "aiven_clickhouse_grant" "foo-role-grant" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  role         = aiven_clickhouse_role.foo-role.role

  privilege_grant {
    privilege = "CREATE TEMPORARY TABLE"
    database  = "*"
  }

  privilege_grant {
    privilege = "DROP FUNCTION"
    database  = "*"
  }

  privilege_grant {
    privilege = "S3"
    database  = "*"
  }

  privilege_grant {
    privilege = "INSERT"
    database  = aiven_clickhouse_database.testdb.name
    table     = "*"
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-role-grant",
						"privilege_grant.*",
						map[string]string{
							"privilege": "CREATE TEMPORARY TABLE",
							"database":  "*",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-role-grant",
						"privilege_grant.*",
						map[string]string{
							"privilege": "DROP FUNCTION",
							"database":  "*",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-role-grant",
						"privilege_grant.*",
						map[string]string{
							"privilege": "S3",
							"database":  "*",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-role-grant",
						"privilege_grant.*",
						map[string]string{
							"privilege": "INSERT",
							"database":  "test-db",
							"table":     "*",
						},
					),
				),
			},
			{
				// Step 5: Clickhouse allows having a table name as asterisk (*)
				// So this scenario grants SELECT privilege only on one table with name "*", not all tables in the database
				Config: baseConfig + `
resource "aiven_clickhouse_grant" "foo-role-grant" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  role         = aiven_clickhouse_role.foo-role.role

  privilege_grant {
    privilege = "SELECT"
    database  = "default"
    table     = "*"
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-role-grant",
						"privilege_grant.*",
						map[string]string{
							"privilege": "SELECT",
							"database":  "default",
							"table":     "*",
						},
					),
				),
			},
			{
				// Step 6: This test case is to check the error when using asterisk (*) for both database and table
				Config: baseConfig + `
resource "aiven_clickhouse_grant" "foo-role-grant" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  role         = aiven_clickhouse_role.foo-role.role

  privilege_grant {
    privilege = "SELECT"
    database  = "*"
    table     = "*"
  }
}`,
				ExpectError: regexp.MustCompile("DB::Exception: Syntax error.*`\\*`"),
			},
			{
				// Step 7: Test user grant with role
				Config: baseConfig + `
resource "aiven_clickhouse_grant" "foo-user-grant" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  user         = aiven_clickhouse_user.foo-user.username

  role_grant {
    role = aiven_clickhouse_role.foo-role.role
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aiven_clickhouse_grant.foo-user-grant",
						"user",
						"foo-user",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-user-grant",
						"role_grant.*",
						map[string]string{
							"role": "foo-role",
						},
					),
				),
			},
			{
				// Step 8: Test both role grant and user grant together
				Config: baseConfig + `
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

  privilege_grant {
    privilege = "CREATE TEMPORARY TABLE"
    database  = "*"
  }
}

resource "aiven_clickhouse_grant" "foo-user-grant" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  user         = aiven_clickhouse_user.foo-user.username

  role_grant {
    role = aiven_clickhouse_role.foo-role.role
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-role-grant",
						"privilege_grant.*",
						map[string]string{
							"privilege": "INSERT",
							"database":  "test-db",
							"table":     "test-table",
							"column":    "test-column",
						},
					),
					resource.TestCheckResourceAttr(
						"aiven_clickhouse_grant.foo-user-grant",
						"user",
						"foo-user",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo-user-grant",
						"role_grant.*",
						map[string]string{
							"role": "foo-role",
						},
					),
				),
			},
			{
				// Step 9: Test import of role grant
				ResourceName:      "aiven_clickhouse_grant.foo-role-grant",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Step 10: Test import of user grant
				ResourceName:      "aiven_clickhouse_grant.foo-user-grant",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAivenClickhouseGrantRole demonstrates the creation of a ClickHouse grant for a role
// with overlapping privileges. It leads to non-empty plan output.
func TestAccAivenClickhouseOverlappingGrants(t *testing.T) {
	serviceName := fmt.Sprintf("test-acc-ch-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	projectName := acc.ProjectName()

	baseConfig := fmt.Sprintf(`
resource "aiven_clickhouse" "bar" {
  project                 = "%s"
  cloud_name              = "google-europe-west1"
  plan                    = "startup-8"
  service_name            = "%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_clickhouse_database" "testdb" {
  project      = aiven_clickhouse.bar.project
  service_name = aiven_clickhouse.bar.service_name
  name         = "test-db"
}

resource "aiven_clickhouse_user" "foo-user" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  username     = "foo-user"
}

variable "database_privileges_list" {
  description = "List of privileges to grant on the main database"
  type        = list(string)
  default = [
    "SELECT",
    "INSERT",
    "DELETE", # overlapping with ALTER DELETE
    "ALTER UPDATE",
    "ALTER VIEW",
    "ALTER INDEX",
    "ALTER DELETE",
    "ALTER ADD PROJECTION", # overlapping with ALTER PROJECTION
    "ALTER COLUMN",
    "ALTER CONSTRAINT",
    "ALTER FETCH PARTITION",
    "ALTER MATERIALIZE TTL",
    "ALTER MOVE PARTITION",
    "ALTER SETTINGS",
    "ALTER TTL",
    "ALTER MODIFY COMMENT",
    "ALTER PROJECTION",
    "CREATE TABLE",
    "CREATE VIEW",
    "CREATE DICTIONARY",
    "DROP TABLE",
    "DROP VIEW",
    "DROP DICTIONARY",
    "dictGet",
    "TRUNCATE",
    "SHOW"
  ]
}

resource "aiven_clickhouse_grant" "foo" {
  project      = aiven_clickhouse.bar.project
  service_name = aiven_clickhouse.bar.service_name
  user         = aiven_clickhouse_user.foo-user.username

  # Dynamically grant privileges from the list to main_db
  dynamic "privilege_grant" {
    for_each = toset(var.database_privileges_list)
    content {
      privilege = privilege_grant.value
      database  = aiven_clickhouse_database.testdb.name
    }
  }

  # Privileges on all databases ("*")
  privilege_grant {
    privilege = "CREATE TEMPORARY TABLE"
    database  = "*"
  }

  privilege_grant {
    privilege = "CREATE FUNCTION"
    database  = "*"
  }

  privilege_grant {
    privilege = "DROP FUNCTION"
    database  = "*"
  }

  privilege_grant {
    privilege = "POSTGRES"
    database  = "*"
  }
}
`, projectName, serviceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenClickhouseGrantResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: baseConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"aiven_clickhouse_grant.foo",
						"privilege_grant.*",
						map[string]string{
							"privilege": "SELECT",
							"database":  "test-db",
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccAivenClickhouseGrantInvalid tests the case where neither user nor role is specified in the grant.
// This should fail with an error.
func TestAccAivenClickhouseGrantInvalid(t *testing.T) {
	serviceName := fmt.Sprintf("test-acc-ch-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	projectName := acc.ProjectName()

	invalidManifest := fmt.Sprintf(`
resource "aiven_clickhouse" "bar" {
  project                 = "%s"
  cloud_name              = "google-europe-west1"
  plan                    = "startup-8"
  service_name            = "%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_clickhouse_database" "testdb" {
  project      = aiven_clickhouse.bar.project
  service_name = aiven_clickhouse.bar.service_name
  name         = "test-db"
}

resource "aiven_clickhouse_grant" "invalid-grant" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project

  privilege_grant {
    privilege = "SELECT"
    database  = aiven_clickhouse_database.testdb.name
  }
}`, projectName, serviceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test that attempting to create a grant without specifying user or role fails
				Config:      invalidManifest,
				ExpectError: regexp.MustCompile(`"(?:user|role)": one of ` + "`" + `role,user` + "`" + ` must be specified`),
			},
		},
	})
}

func testAccCheckAivenClickhouseGrantResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each aiven_clickhouse_role is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_clickhouse_grant" {
			continue
		}

		projectName, serviceName, granteeType, granteeName, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return err
		}

		grantee := clickhouse2.Grantee{}
		switch granteeType {
		case clickhouse2.GranteeTypeRole:
			grantee.Role = granteeName
		case clickhouse2.GranteeTypeUser:
			grantee.User = granteeName
		}

		if privilegeGrants, err := clickhouse2.ReadPrivilegeGrants(
			ctx,
			c,
			projectName,
			serviceName,
			grantee,
		); err != nil {
			if aiven.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("unable to check if privilege grants for '%s' still exists: %w", granteeName, err)
		} else if len(privilegeGrants) > 0 {
			return fmt.Errorf("'%s' still has privilege grants exists", granteeName)
		}

		if roleGrants, err := clickhouse2.ReadRoleGrants(ctx, c, projectName, serviceName, grantee); err != nil {
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
