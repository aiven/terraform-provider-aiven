package valkey_test

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
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
				),
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
}`, acc.ProjectName(), name, name)
}
