package influxdb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenInfluxDBDatabase_basic(t *testing.T) {
	resourceName := "aiven_influxdb_database.foo"
	projectName := acc.ProjectName()
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenInfluxDBDatabaseResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfluxDBDatabaseResource(projectName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenInfluxDBDatabaseAttributes("data.aiven_influxdb_database.database"),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "database_name", fmt.Sprintf("test-acc-db-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config:                    testAccInfluxDBDatabaseTerminationProtectionResource(projectName, rName2),
				PreventPostDestroyRefresh: true,
				ExpectNonEmptyPlan:        true,
				PlanOnly:                  true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName2)),
					resource.TestCheckResourceAttr(resourceName, "database_name", fmt.Sprintf("test-acc-db-%s", rName2)),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
				),
			},
			{
				Config:       testAccInfluxDBDatabaseResource(projectName, rName),
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("expected resource '%s' to be present in the state", resourceName)
					}
					if _, ok := rs.Primary.Attributes["database_name"]; !ok {
						return "", fmt.Errorf("expected resource '%s' to have 'database_name' attribute", resourceName)
					}
					return rs.Primary.ID, nil
				},
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return fmt.Errorf("expected only one instance to be imported, state: %#v", s)
					}
					attributes := s[0].Attributes
					if !strings.EqualFold(attributes["project"], projectName) {
						return fmt.Errorf("expected project to match '%s', got: '%s'", projectName, attributes["project_name"])
					}
					databaseName, ok := attributes["database_name"]
					if !ok {
						return errors.New("expected 'database_name' field to be set")
					}
					expectedID := fmt.Sprintf("%s/test-acc-sr-%s/%s", projectName, rName, databaseName)
					if !strings.EqualFold(s[0].ID, expectedID) {
						return fmt.Errorf("expected ID to match '%s', but got: %s", expectedID, s[0].ID)
					}
					return nil
				},
			},
		},
	})
}

func testAccCheckAivenInfluxDBDatabaseResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each database is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_influxdb_database" {
			continue
		}

		projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		db, err := c.Databases.Get(ctx, projectName, serviceName, databaseName)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status != 404 {
				return err
			}
		}

		if db != nil {
			return fmt.Errorf("database (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccInfluxDBDatabaseResource(project string, name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_influxdb" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  influxdb_user_config {
    public_access {
      influxdb = true
    }
  }
}

resource "aiven_influxdb_database" "foo" {
  project       = aiven_influxdb.bar.project
  service_name  = aiven_influxdb.bar.service_name
  database_name = "test-acc-db-%s"
}

data "aiven_influxdb_database" "database" {
  project       = aiven_influxdb_database.foo.project
  service_name  = aiven_influxdb_database.foo.service_name
  database_name = aiven_influxdb_database.foo.database_name

  depends_on = [aiven_influxdb_database.foo]
}`, project, name, name)
}

func testAccInfluxDBDatabaseTerminationProtectionResource(project string, name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_influxdb" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  influxdb_user_config {
    public_access {
      influxdb = true
    }
  }
}

resource "aiven_influxdb_database" "foo" {
  project                = aiven_influxdb.bar.project
  service_name           = aiven_influxdb.bar.service_name
  database_name          = "test-acc-db-%s"
  termination_protection = true
}

data "aiven_influxdb_database" "database" {
  project       = aiven_influxdb_database.foo.project
  service_name  = aiven_influxdb_database.foo.service_name
  database_name = aiven_influxdb_database.foo.database_name

  depends_on = [aiven_influxdb_database.foo]
}`, project, name, name)
}

func testAccCheckAivenInfluxDBDatabaseAttributes(n string) resource.TestCheckFunc {
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

		return nil
	}
}
