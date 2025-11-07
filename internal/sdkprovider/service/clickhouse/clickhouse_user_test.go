package clickhouse_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenClickhouseUser(t *testing.T) {
	resourceName := "aiven_clickhouse_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source:            "hashicorp/random",
				VersionConstraint: ">= 3.0.0",
			},
		},
		CheckDestroy: testAccCheckAivenClickhouseUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				// create user with generated password
				Config: testAccClickhouseUserResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "password"), // verify generated password is in state
					resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "password_wo_version"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttrSet(resourceName, "required"),
					testAccCheckAivenClickhouseUserAttributes("data.aiven_clickhouse_user.user"),
				),
			},
			{
				// migrate to write-only password
				Config: testAccClickhouseUserResourceWriteOnly(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "password_wo"), // verify password_wo is NOT in state
					resource.TestCheckResourceAttr(resourceName, "password", ""),  // verify password is empty when using write-only
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
				),
			},
			{
				// rotate password
				Config: testAccClickhouseUserResourceWriteOnly(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password_wo_version", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
					resource.TestCheckResourceAttr(resourceName, "password", ""),
				),
			},
			{
				// refresh state without changes
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password_wo_version", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
					resource.TestCheckResourceAttr(resourceName, "password", ""),
				),
			},
			{
				// attempt to decrement version (should fail)
				Config:      testAccClickhouseUserResourceWriteOnly(rName, 1),
				ExpectError: regexp.MustCompile(`password_wo_version must be incremented.*`),
			},
			{
				// transition back to auto-generated password
				Config: testAccClickhouseUserResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "password"), // verify new generated password is in state
					resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
					resource.TestCheckResourceAttr(resourceName, "password_wo_version", "0"),
				),
			},
			{
				// switch back to write-only password with new version 1
				Config: testAccClickhouseUserResourceWriteOnly(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
					resource.TestCheckResourceAttr(resourceName, "password", ""), // verify password is empty when using write-only
				),
			},
		},
	})
}

func testAccCheckAivenClickhouseUserResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each aiven_clickhouse_user is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_clickhouse_user" {
			continue
		}

		projectName, serviceName, uuid, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		users, err := c.ServiceClickHouseUserList(ctx, projectName, serviceName)
		if err != nil {
			if avngen.IsNotFound(err) {
				// if we get an error, consider it destroyed
				return nil
			}

			return err
		}

		for _, u := range users {
			if u.Uuid == uuid {
				return fmt.Errorf("clickhouse user (%q) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccClickhouseUserResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_clickhouse" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-16"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_clickhouse_user" "foo" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  username     = "user-%s"
}

data "aiven_clickhouse_user" "user" {
  service_name = aiven_clickhouse_user.foo.service_name
  project      = aiven_clickhouse_user.foo.project
  username     = aiven_clickhouse_user.foo.username
}`, acc.ProjectName(), name, name)
}

func testAccClickhouseUserResourceWriteOnly(name string, version int) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%[1]s"
}

resource "aiven_clickhouse" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-16"
  service_name            = "test-acc-sr-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "random_password" "pw" {
  length  = 24
  special = true
  keepers = {
    version = %[3]d
  }
}

resource "aiven_clickhouse_user" "foo" {
  service_name        = aiven_clickhouse.bar.service_name
  project             = aiven_clickhouse.bar.project
  username            = "user-%[2]s"
  password_wo         = random_password.pw.result
  password_wo_version = %[3]d
}`, acc.ProjectName(), name, version)
}

func testAccCheckAivenClickhouseUserAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["username"] == "" {
			return fmt.Errorf("expected to get a clikchouse user username from Aiven")
		}

		if a["project"] == "" {
			return fmt.Errorf("expected to get a clickhouse user project from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a clickhouse user service_name from Aiven")
		}

		return nil
	}
}
