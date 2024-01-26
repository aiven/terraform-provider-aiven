package opensearch_test

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

// openSearchTestPassword is the password used for the OpenSearch tests.
const openSearchTestPassword = "ThisIsATest123^=^"

func TestAccAivenOpenSearchUser_basic(t *testing.T) {
	resourceName := "aiven_opensearch_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOpenSearchUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenSearchUserResource(rName),
				Check: resource.ComposeTestCheckFunc(
					schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_opensearch_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "password", openSearchTestPassword),
				),
			},
		},
	})
}

func testAccCheckAivenOpenSearchUserResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each aiven_opensearch_user is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_opensearch_user" {
			continue
		}

		projectName, serviceName, username, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		p, err := c.ServiceUsers.Get(ctx, projectName, serviceName, username)
		if err != nil {
			var e *aiven.Error
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

func testAccOpenSearchUserResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%[1]s"
}

resource "aiven_opensearch" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_opensearch_user" "foo" {
  service_name = aiven_opensearch.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-%[2]s"
  password     = "%[3]s"
}

data "aiven_opensearch_user" "user" {
  service_name = aiven_opensearch_user.foo.service_name
  project      = aiven_opensearch_user.foo.project
  username     = aiven_opensearch_user.foo.username
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, openSearchTestPassword)
}
