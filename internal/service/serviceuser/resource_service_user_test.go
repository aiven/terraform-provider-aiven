package serviceuser_test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aiven/aiven-go-client"
	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenServiceUser_basic(t *testing.T) {
	resourceName := "aiven_service_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserNewPasswordResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceUserAttributes("data.aiven_service_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
				),
			},
		},
	})
}

// pg no password, password is used in template interpolation
func TestAccAivenServiceUser_pg_no_password(t *testing.T) {
	resourceName := "aiven_service_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserNoPasswordResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceUserAttributes("data.aiven_service_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenServiceUser_pg_replication(t *testing.T) {
	resourceName := "aiven_service_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserPgReplicationResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceUserAttributes("data.aiven_service_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
					resource.TestCheckResourceAttr(resourceName, "pg_allow_replication", "true"),
				),
			},
		},
	})
}

func TestAccAivenServiceUser_redis_acls(t *testing.T) {
	resourceName := "aiven_service_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserRedisACLResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceUserAttributes("data.aiven_service_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
				),
			},
		},
	})
}

func testAccCheckAivenServiceUserResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each aiven_service_user is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_service_user" {
			continue
		}

		projectName, serviceName, username, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		p, err := c.ServiceUsers.Get(projectName, serviceName, username)
		if err != nil {
			var aivenError *aiven.Error

			if ok := errors.As(err, &aivenError); !ok || aivenError.Status != 404 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("common user (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccServiceUserRedisACLResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_redis" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-%s"
}

resource "aiven_service_user" "foo" {
  service_name = aiven_redis.bar.service_name
  project      = aiven_redis.bar.project
  username     = "user-%s"

  redis_acl_commands   = ["+set"]
  redis_acl_keys       = ["prefix*", "another_key"]
  redis_acl_categories = ["-@all", "+@admin"]
  redis_acl_channels   = ["test"]

  depends_on = [aiven_redis.bar]
}

data "aiven_service_user" "user" {
  service_name = aiven_service_user.foo.service_name
  project      = aiven_service_user.foo.project
  username     = aiven_service_user.foo.username

  depends_on = [aiven_service_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccServiceUserPgReplicationResource(name string) string {
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

resource "aiven_service_user" "foo" {
  service_name         = aiven_pg.bar.service_name
  project              = aiven_pg.bar.project
  username             = "user-%s"
  password             = "Test$1234"
  pg_allow_replication = true

  depends_on = [aiven_pg.bar]
}

data "aiven_service_user" "user" {
  service_name = aiven_service_user.foo.service_name
  project      = aiven_service_user.foo.project
  username     = aiven_service_user.foo.username

  depends_on = [aiven_service_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccServiceUserNewPasswordResource(name string) string {
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

resource "aiven_service_user" "foo" {
  service_name = aiven_pg.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-%s"
  password     = "Test$1234"

  depends_on = [aiven_pg.bar]
}

data "aiven_service_user" "user" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project
  username     = aiven_service_user.foo.username

  depends_on = [aiven_service_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccServiceUserNoPasswordResource(name string) string {
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

resource "aiven_service_user" "foo" {
  service_name = aiven_pg.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-%s"

  depends_on = [aiven_pg.bar]
}

// check that we can use the password in template interpolations
output "use-template-interpolation" {
  sensitive = true
  value     = "${aiven_service_user.foo.password}/testing"
}

data "aiven_service_user" "user" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project
  username     = aiven_service_user.foo.username

  depends_on = [aiven_service_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccCheckAivenServiceUserAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] user service attributes %v", a)

		if a["username"] == "" {
			return fmt.Errorf("expected to get a service user username from Aiven")
		}

		if a["password"] == "" {
			return fmt.Errorf("expected to get a service user password from Aiven")
		}

		if a["type"] == "" {
			return fmt.Errorf("expected to get a service user type from Aiven")
		}

		if a["project"] == "" {
			return fmt.Errorf("expected to get a service user project from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service user service_name from Aiven")
		}

		return nil
	}
}
