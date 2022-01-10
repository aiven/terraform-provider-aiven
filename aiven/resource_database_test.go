// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"os"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aiven_database", &resource.Sweeper{
		Name: "aiven_database",
		F:    sweepDatabases,
		Dependencies: []string{
			"aiven_connection_pool",
		},
	})
}

func sweepDatabases(region string) error {
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
				dbs, err := conn.Databases.List(project.Name, service.Name)
				if err != nil {
					if err.(aiven.Error).Status == 403 || err.(aiven.Error).Status == 501 {
						continue
					}

					return fmt.Errorf("error retrieving a list of databases for a service `%s`: %s", service.Name, err)
				}

				for _, db := range dbs {
					if db.DatabaseName == "defaultdb" {
						continue
					}

					err = conn.Databases.Delete(project.Name, service.Name, db.DatabaseName)
					if err != nil {
						return fmt.Errorf("error destroying database `%s` during sweep: %s", db.DatabaseName, err)
					}
				}
			}
		}
	}

	return nil
}

func TestAccAivenDatabase_basic(t *testing.T) {
	resourceName := "aiven_database.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenDatabaseResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenDatabaseAttributes("data.aiven_database.database"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "database_name", fmt.Sprintf("test-acc-db-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config:                    testAccDatabaseTerminationProtectionResource(rName2),
				PreventPostDestroyRefresh: true,
				ExpectNonEmptyPlan:        true,
				PlanOnly:                  true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName2)),
					resource.TestCheckResourceAttr(resourceName, "database_name", fmt.Sprintf("test-acc-db-%s", rName2)),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
				),
			},
		},
	})
}

func testAccCheckAivenDatabaseResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each database is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_database" {
			continue
		}

		projectName, serviceName, databaseName := splitResourceID3(rs.Primary.ID)
		db, err := c.Databases.Get(projectName, serviceName, databaseName)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if db != nil {
			return fmt.Errorf("databse (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccDatabaseResource(name string) string {
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
		
		  pg_user_config {
		    public_access {
		      pg         = true
		      prometheus = false
		    }
		
		    pg {
		      idle_in_transaction_session_timeout = 900
		    }
		  }
		}
		
		resource "aiven_database" "foo" {
		  project       = aiven_pg.bar.project
		  service_name  = aiven_pg.bar.service_name
		  database_name = "test-acc-db-%s"
		  lc_ctype      = "en_US.UTF-8"
		  lc_collate    = "en_US.UTF-8"
		}
		
		data "aiven_database" "database" {
		  project       = aiven_database.foo.project
		  service_name  = aiven_database.foo.service_name
		  database_name = aiven_database.foo.database_name
		
		  depends_on = [aiven_database.foo]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccDatabaseTerminationProtectionResource(name string) string {
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
		
		  pg_user_config {
		    public_access {
		      pg         = true
		      prometheus = false
		    }
		
		    pg {
		      idle_in_transaction_session_timeout = 900
		    }
		  }
		}
		
		resource "aiven_database" "foo" {
		  project                = aiven_pg.bar.project
		  service_name           = aiven_pg.bar.service_name
		  database_name          = "test-acc-db-%s"
		  termination_protection = true
		}
		
		data "aiven_database" "database" {
		  project       = aiven_database.foo.project
		  service_name  = aiven_database.foo.service_name
		  database_name = aiven_database.foo.database_name
		
		  depends_on = [aiven_database.foo]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccCheckAivenDatabaseAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["database_name"] == "" {
			return fmt.Errorf("expected to get a database_name from Aiven")
		}

		if a["database_name"] == "" {
			return fmt.Errorf("expected to get a database_name from Aiven")
		}

		return nil
	}
}
