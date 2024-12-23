package alloydbomni_test

import (
	"context"
	"errors"
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

func TestAccAivenAlloyDBOmniUser_basic(t *testing.T) {
	resourceName := "aiven_alloydbomni_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAlloyDBOmniUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniUserNewPasswordResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_alloydbomni_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
				),
			},
		},
	})
}

func TestAccAivenAlloyDBOmniUser_alloydbomni_no_password(t *testing.T) {
	resourceName := "aiven_alloydbomni_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAlloyDBOmniUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniUserNoPasswordResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_alloydbomni_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenAlloyDBOmniUser_alloydbomni_replica(t *testing.T) {
	resourceName := "aiven_alloydbomni_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	project := os.Getenv("AIVEN_PROJECT_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAlloyDBOmniUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniUserPgReplicationResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_alloydbomni_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
					resource.TestCheckResourceAttr(resourceName, "pg_allow_replication", "true"),
				),
			},
			{
				Config: testAccAlloyDBOmniUserPgReplicationDisableResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_alloydbomni_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
					resource.TestCheckResourceAttr(resourceName, "pg_allow_replication", "false"),
				),
			},
			{
				Config: testAccAlloyDBOmniUserPgReplicationEnableResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_alloydbomni_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
					resource.TestCheckResourceAttr(resourceName, "pg_allow_replication", "true"),
				),
			},
		},
	})
}

func testAccCheckAivenAlloyDBOmniUserResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	// loop through the resources in state, verifying each aiven_alloydbomni_user is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_alloydbomni_user" {
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

func testAccAlloyDBOmniUserPgReplicationResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-%s"
}

resource "aiven_alloydbomni_user" "foo" {
  service_name         = aiven_alloydbomni.bar.service_name
  project              = aiven_alloydbomni.bar.project
  username             = "user-%s"
  password             = "Test$1234"
  pg_allow_replication = true

  depends_on = [aiven_alloydbomni.bar]
}

data "aiven_alloydbomni_user" "user" {
  service_name = aiven_alloydbomni_user.foo.service_name
  project      = aiven_alloydbomni_user.foo.project
  username     = aiven_alloydbomni_user.foo.username

  depends_on = [aiven_alloydbomni_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccAlloyDBOmniUserPgReplicationDisableResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-%s"
}

resource "aiven_alloydbomni_user" "foo" {
  service_name         = aiven_alloydbomni.bar.service_name
  project              = aiven_alloydbomni.bar.project
  username             = "user-%s"
  password             = "Test$1234"
  pg_allow_replication = false

  depends_on = [aiven_alloydbomni.bar]
}

data "aiven_alloydbomni_user" "user" {
  service_name = aiven_alloydbomni_user.foo.service_name
  project      = aiven_alloydbomni_user.foo.project
  username     = aiven_alloydbomni_user.foo.username

  depends_on = [aiven_alloydbomni_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccAlloyDBOmniUserPgReplicationEnableResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-%s"
}

resource "aiven_alloydbomni_user" "foo" {
  service_name         = aiven_alloydbomni.bar.service_name
  project              = aiven_alloydbomni.bar.project
  username             = "user-%s"
  password             = "Test$1234"
  pg_allow_replication = true

  depends_on = [aiven_alloydbomni.bar]
}

data "aiven_alloydbomni_user" "user" {
  service_name = aiven_alloydbomni_user.foo.service_name
  project      = aiven_alloydbomni_user.foo.project
  username     = aiven_alloydbomni_user.foo.username

  depends_on = [aiven_alloydbomni_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccAlloyDBOmniUserNewPasswordResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_alloydbomni_user" "foo" {
  service_name = aiven_alloydbomni.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-%s"
  password     = "Test$1234"

  depends_on = [aiven_alloydbomni.bar]
}

data "aiven_alloydbomni_user" "user" {
  service_name = aiven_alloydbomni.bar.service_name
  project      = aiven_alloydbomni.bar.project
  username     = aiven_alloydbomni_user.foo.username

  depends_on = [aiven_alloydbomni_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccAlloyDBOmniUserNoPasswordResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_alloydbomni_user" "foo" {
  service_name = aiven_alloydbomni.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-%s"

  depends_on = [aiven_alloydbomni.bar]
}

// check that we can use the password in template interpolations
output "use-template-interpolation" {
  sensitive = true
  value     = "${aiven_alloydbomni_user.foo.password}/testing"
}

data "aiven_alloydbomni_user" "user" {
  service_name = aiven_alloydbomni.bar.service_name
  project      = aiven_alloydbomni.bar.project
  username     = aiven_alloydbomni_user.foo.username

  depends_on = [aiven_alloydbomni_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}
