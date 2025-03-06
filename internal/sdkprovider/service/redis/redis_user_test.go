package redis_test

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

func TestAccAivenRedisUser_basic(t *testing.T) {
	t.Skip(redisDeprecated)

	resourceName := "aiven_redis_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenRedisUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRedisUserRedisACLResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_redis_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
				),
			},
		},
	})
}

func testAccCheckAivenRedisUserResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each aiven_redis_user is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_redis_user" {
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

func testAccRedisUserRedisACLResource(name string) string {
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

resource "aiven_redis_user" "foo" {
  service_name = aiven_redis.bar.service_name
  project      = aiven_redis.bar.project
  username     = "user-%s"
  password     = "Test$1234"

  redis_acl_commands   = ["+set"]
  redis_acl_keys       = ["prefix*", "another_key"]
  redis_acl_categories = ["-@all", "+@admin"]
  redis_acl_channels   = ["test"]

  depends_on = [aiven_redis.bar]
}

data "aiven_redis_user" "user" {
  service_name = aiven_redis_user.foo.service_name
  project      = aiven_redis_user.foo.project
  username     = aiven_redis_user.foo.username

  depends_on = [aiven_redis_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}
