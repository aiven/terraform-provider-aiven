package valkey_test

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

func TestAccAivenValkeyUser_basic(t *testing.T) {
	resourceName := "aiven_valkey_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenValkeyUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccValkeyUserValkeyACLResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_valkey_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.0", "+set"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_channels.#", "1"),
				),
			},
			{
				// update ACL fields in-place (no recreation)
				Config: testAccValkeyUserUpdatedACL(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.0", "+set"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.1", "+get"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.0", "prefix*"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.1", "another_key"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.2", "new_key*"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.0", "-@all"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.1", "+@admin"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.2", "+@read"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_channels.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_channels.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "valkey_acl_channels.1", "notifications"),
				),
			},
			{
				// test import - verifies backward compatibility with existing state
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccCheckAivenValkeyUserResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each aiven_valkey_user is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_valkey_user" {
			continue
		}

		projectName, serviceName, username, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

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

func testAccValkeyUserValkeyACLResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_valkey" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-%s"
}

resource "aiven_valkey_user" "foo" {
  service_name = aiven_valkey.bar.service_name
  project      = aiven_valkey.bar.project
  username     = "user-%s"
  password     = "Test$1234"

  valkey_acl_commands   = ["+set"]
  valkey_acl_keys       = ["prefix*", "another_key"]
  valkey_acl_categories = ["-@all", "+@admin"]
  valkey_acl_channels   = ["test"]

  depends_on = [aiven_valkey.bar]
}

data "aiven_valkey_user" "user" {
  service_name = aiven_valkey_user.foo.service_name
  project      = aiven_valkey_user.foo.project
  username     = aiven_valkey_user.foo.username

  depends_on = [aiven_valkey_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccValkeyUserUpdatedACL(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_valkey" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-%s"
}

resource "aiven_valkey_user" "foo" {
  service_name = aiven_valkey.bar.service_name
  project      = aiven_valkey.bar.project
  username     = "user-%s"
  password     = "Test$1234"

  valkey_acl_commands   = ["+set", "+get"]
  valkey_acl_keys       = ["prefix*", "another_key", "new_key*"]
  valkey_acl_categories = ["-@all", "+@admin", "+@read"]
  valkey_acl_channels   = ["test", "notifications"]

  depends_on = [aiven_valkey.bar]
}

data "aiven_valkey_user" "user" {
  service_name = aiven_valkey_user.foo.service_name
  project      = aiven_valkey_user.foo.project
  username     = aiven_valkey_user.foo.username

  depends_on = [aiven_valkey_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}
