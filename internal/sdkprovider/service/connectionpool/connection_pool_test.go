package connectionpool_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenConnectionPool_basic(t *testing.T) {
	resourceName := "aiven_connection_pool.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenConnectionPoolResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionPoolResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenConnectionPoolAttributes("data.aiven_connection_pool.pool"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "database_name", fmt.Sprintf("test-acc-db-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "pool_name", fmt.Sprintf("test-acc-pool-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "pool_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "pool_mode", "transaction"),
				),
			},
			{
				Config: testAccConnectionPoolNoUserResource(rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName2)),
					resource.TestCheckResourceAttr(resourceName, "database_name", fmt.Sprintf("test-acc-db-%s", rName2)),
					resource.TestCheckResourceAttr(resourceName, "pool_name", fmt.Sprintf("test-acc-pool-%s", rName2)),
					resource.TestCheckResourceAttr(resourceName, "pool_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "pool_mode", "transaction"),
				),
			},
		},
	})
}

func testAccConnectionPoolNoUserResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_pg_database" "foo" {
  project       = aiven_pg.bar.project
  service_name  = aiven_pg.bar.service_name
  database_name = "test-acc-db-%s"
}

resource "aiven_connection_pool" "foo" {
  service_name  = aiven_pg.bar.service_name
  project       = data.aiven_project.foo.project
  database_name = aiven_pg_database.foo.database_name
  pool_name     = "test-acc-pool-%s"
  pool_size     = 25
  pool_mode     = "transaction"

  depends_on = [aiven_pg_database.foo]
}

data "aiven_connection_pool" "pool" {
  project      = aiven_connection_pool.foo.project
  service_name = aiven_connection_pool.foo.service_name
  pool_name    = aiven_connection_pool.foo.pool_name

  depends_on = [aiven_connection_pool.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name, name)
}

func testAccConnectionPoolResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_pg_user" "foo" {
  service_name = aiven_pg.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-%s"
}

resource "aiven_pg_database" "foo" {
  project       = aiven_pg.bar.project
  service_name  = aiven_pg.bar.service_name
  database_name = "test-acc-db-%s"
}

resource "aiven_connection_pool" "foo" {
  service_name  = aiven_pg.bar.service_name
  project       = data.aiven_project.foo.project
  database_name = aiven_pg_database.foo.database_name
  username      = aiven_pg_user.foo.username
  pool_name     = "test-acc-pool-%s"
  pool_size     = 25
  pool_mode     = "transaction"

  depends_on = [aiven_pg_database.foo]
}

data "aiven_connection_pool" "pool" {
  project      = aiven_connection_pool.foo.project
  service_name = aiven_connection_pool.foo.service_name
  pool_name    = aiven_connection_pool.foo.pool_name

  depends_on = [aiven_connection_pool.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name, name, name)
}

func testAccCheckAivenConnectionPoolAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["pool_name"] == "" {
			return fmt.Errorf("expected to get a pool_name from Aiven")
		}

		if a["database_name"] == "" {
			return fmt.Errorf("expected to get a database_name from Aiven")
		}

		if a["username"] == "" {
			return fmt.Errorf("expected to get a username from Aiven")
		}

		if a["pool_size"] != "25" {
			return fmt.Errorf("expected to get a correct pool_size from Aiven")
		}

		if a["pool_mode"] != "transaction" {
			return fmt.Errorf("expected to get a correct pool_mode from Aiven")
		}

		return nil
	}
}

func testAccCheckAivenConnectionPoolResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each connection pool is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_connection_pool" {
			continue
		}

		projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		pool, err := c.ConnectionPools.Get(ctx, projectName, serviceName, databaseName)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if pool != nil {
			return fmt.Errorf("connection pool (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
