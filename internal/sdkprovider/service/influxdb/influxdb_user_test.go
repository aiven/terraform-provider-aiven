package influxdb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aiven/aiven-go-client"
<<<<<<< HEAD
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
=======

	acctest3 "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283))
)

func TestAccAivenInfluxDBUser_basic(t *testing.T) {
	resourceName := "aiven_influxdb_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenInfluxDBUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfluxDBUserResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_influxdb_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
				),
			},
		},
	})
}

func testAccCheckAivenInfluxDBUserResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	// loop through the resources in state, verifying each aiven_influxdb_user is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_influxdb_user" {
			continue
		}

		projectName, serviceName, username, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		p, err := c.ServiceUsers.Get(projectName, serviceName, username)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("common user (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccInfluxDBUserResource(name string) string {
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
}

resource "aiven_influxdb_user" "foo" {
  service_name = aiven_influxdb.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-%s"
  password     = "Test$1234"
}

data "aiven_influxdb_user" "user" {
  service_name = aiven_influxdb_user.foo.service_name
  project      = aiven_influxdb_user.foo.project
  username     = aiven_influxdb_user.foo.username
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

// InfluxDB service tests
func TestAccAivenService_influxdb(t *testing.T) {
	resourceName := "aiven_influxdb.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfluxdbServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_influxdb.common"),
					testAccCheckAivenServiceInfluxdbAttributes("data.aiven_influxdb.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "influxdb"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
		},
	})
}

func testAccInfluxdbServiceResource(name string) string {
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

data "aiven_influxdb" "common" {
  service_name = aiven_influxdb.bar.service_name
  project      = aiven_influxdb.bar.project

  depends_on = [aiven_influxdb.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccCheckAivenServiceInfluxdbAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["influxdb_user_config.0.public_access.0.influxdb"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.influxdb from Aiven")
		}

		return nil
	}
}
