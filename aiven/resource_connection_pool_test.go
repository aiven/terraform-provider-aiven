package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

func init() {
	resource.AddTestSweepers("aiven_connection_pool", &resource.Sweeper{
		Name: "aiven_connection_pool",
		F:    sweepConnectionPools,
	})
}

func sweepConnectionPools(region string) error {
	client, err := sharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projects, err := conn.Projects.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of projects : %s", err)
	}

	for _, project := range projects {
		if project.Name == os.Getenv("AIVEN_PROJECT_NAME") {
			services, err := conn.Services.List(project.Name)
			if err != nil {
				return fmt.Errorf("error retrieving a list of services for a project `%s`: %s", project.Name, err)
			}

			for _, service := range services {
				list, err := conn.ConnectionPools.List(project.Name, service.Name)
				if err != nil {
					if err.(aiven.Error).Status == 403 {
						continue
					}

					return fmt.Errorf("error retrieving a list of connection pools for a service `%s`: %s", service.Name, err)
				}

				for _, pool := range list {
					err = conn.ConnectionPools.Delete(project.Name, service.Name, pool.PoolName)
					if err != nil {
						return fmt.Errorf("error destroying connection pool `%s` during sweep: %s", pool.PoolName, err)
					}
				}
			}
		}
	}

	return nil
}

func TestAccAivenConnectionPool_basic(t *testing.T) {
	resourceName := "aiven_connection_pool.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenConnectionPoolResourceDestroy,
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
		},
	})
}

func testAccConnectionPoolResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
			project = "%s"
		}

		resource "aiven_service" "bar" {
			project = data.aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			service_type = "pg"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			pg_user_config {
				pg_version = 11
			}
		}
		
		resource "aiven_service_user" "foo" {
			service_name = aiven_service.bar.service_name
			project = data.aiven_project.foo.project
			username = "user-%s"
		}

		resource "aiven_database" "foo" {
			project = aiven_service.bar.project
			service_name = aiven_service.bar.service_name
			database_name = "test-acc-db-%s"
		}

		resource "aiven_connection_pool" "foo" {
			service_name = aiven_service.bar.service_name
			project = data.aiven_project.foo.project
			database_name = aiven_database.foo.database_name
			username = aiven_service_user.foo.username
			pool_name = "test-acc-pool-%s"
			pool_size = 25
			pool_mode = "transaction"
		}

		data "aiven_connection_pool" "pool" {
			project = aiven_connection_pool.foo.project
			service_name = aiven_connection_pool.foo.service_name
			pool_name = aiven_connection_pool.foo.pool_name
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name, name, name, name)
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
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each connection pool is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_connection_pool" {
			continue
		}

		projectName, serviceName, databaseName := splitResourceID3(rs.Primary.ID)
		pool, err := c.ConnectionPools.Get(projectName, serviceName, databaseName)
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
