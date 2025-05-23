package pg_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenPGUser_basic(t *testing.T) {
	resourceName := "aiven_pg_user.foo.0" // checking the first user only
	projectName := acc.ProjectName()
	serviceName := "test-acc-sr-" + acc.RandStr()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGUserNewPasswordResource(projectName, serviceName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_pg_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "username", "user-1"),
					resource.TestCheckResourceAttr(resourceName, "password", "P4$$word"),
				),
			},
			{
				// Validates that the import sets all ForceNew fields
				// https://github.com/aiven/terraform-provider-aiven/issues/2065
				Config: fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  project      = %q
  service_name = %q
  username     = "user-1"
}
`, projectName, serviceName),
				ResourceName:      "aiven_pg_user.foo",
				ImportStateId:     util.ComposeID(projectName, serviceName, "user-1"),
				ImportState:       true,
				ImportStateVerify: true,
				// ImportState uses the current state
				// ImportStateVerify compares it with the state after importing, they must match.
				// Without a fix it outputs:
				// Step 2/2 error running import: ImportStateVerify attributes not equivalent.
				// map[string]string{
				//	"project":      "...",
				//	"service_name": "...",
				//}
			},
		},
	})
}

func TestAccAivenPGUser_pg_no_password(t *testing.T) {
	resourceName := "aiven_pg_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGUserNoPasswordResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_pg_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenPGUser_pg_replica(t *testing.T) {
	resourceName := "aiven_pg_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGUserPgReplicationResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_pg_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
					resource.TestCheckResourceAttr(resourceName, "pg_allow_replication", "true"),
				),
			},
			{
				Config: testAccPGUserPgReplicationDisableResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_pg_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
					resource.TestCheckResourceAttr(resourceName, "pg_allow_replication", "false"),
				),
			},
			{
				Config: testAccPGUserPgReplicationEnableResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_pg_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
					resource.TestCheckResourceAttr(resourceName, "pg_allow_replication", "true"),
				),
			},
		},
	})
}

func testAccCheckAivenPGUserResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	// loop through the resources in state, verifying each aiven_pg_user is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_pg_user" {
			continue
		}

		projectName, serviceName, username, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		ctx := context.Background()

		p, err := c.ServiceUsers.Get(ctx, projectName, serviceName, username)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status != 404 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("common user (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccPGUserPgReplicationResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-%s"
}

resource "aiven_pg_user" "foo" {
  service_name         = aiven_pg.bar.service_name
  project              = aiven_pg.bar.project
  username             = "user-%s"
  password             = "Test$1234"
  pg_allow_replication = true

  depends_on = [aiven_pg.bar]
}

data "aiven_pg_user" "user" {
  service_name = aiven_pg_user.foo.service_name
  project      = aiven_pg_user.foo.project
  username     = aiven_pg_user.foo.username

  depends_on = [aiven_pg_user.foo]
}`, acc.ProjectName(), name, name)
}

func testAccPGUserPgReplicationDisableResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-%s"
}

resource "aiven_pg_user" "foo" {
  service_name         = aiven_pg.bar.service_name
  project              = aiven_pg.bar.project
  username             = "user-%s"
  password             = "Test$1234"
  pg_allow_replication = false

  depends_on = [aiven_pg.bar]
}

data "aiven_pg_user" "user" {
  service_name = aiven_pg_user.foo.service_name
  project      = aiven_pg_user.foo.project
  username     = aiven_pg_user.foo.username

  depends_on = [aiven_pg_user.foo]
}`, acc.ProjectName(), name, name)
}

func testAccPGUserPgReplicationEnableResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-%s"
}

resource "aiven_pg_user" "foo" {
  service_name         = aiven_pg.bar.service_name
  project              = aiven_pg.bar.project
  username             = "user-%s"
  password             = "Test$1234"
  pg_allow_replication = true

  depends_on = [aiven_pg.bar]
}

data "aiven_pg_user" "user" {
  service_name = aiven_pg_user.foo.service_name
  project      = aiven_pg_user.foo.project
  username     = aiven_pg_user.foo.username

  depends_on = [aiven_pg_user.foo]
}`, acc.ProjectName(), name, name)
}

// testAccPGUserNewPasswordResource creates 100 users to test bulk creation
func testAccPGUserNewPasswordResource(projectName, serviceName string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = %[1]q
}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = %[2]q
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_pg_user" "foo" {
  count        = 42
  service_name = aiven_pg.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-${count.index + 1}"
  password     = "P4$$word"
}

data "aiven_pg_user" "user" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project
  username     = aiven_pg_user.foo.0.username

  depends_on = [aiven_pg_user.foo]
}`, projectName, serviceName)
}

func testAccPGUserNoPasswordResource(name string) string {
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

  depends_on = [aiven_pg.bar]
}

// check that we can use the password in template interpolations
output "use-template-interpolation" {
  sensitive = true
  value     = "${aiven_pg_user.foo.password}/testing"
}

data "aiven_pg_user" "user" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project
  username     = aiven_pg_user.foo.username

  depends_on = [aiven_pg_user.foo]
}`, acc.ProjectName(), name, name)
}
