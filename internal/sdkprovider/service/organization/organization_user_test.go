package organization_test

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

func TestAccAivenOrganizationUser_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationUserResource(rName),
				ExpectError: regexp.MustCompile(
					"creation of organization user is not supported anymore via Terraform.*",
				),
			},
		},
	})
}

func testAccOrganizationUserResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%s"
}

resource "aiven_organization_user" "foo" {
  organization_id = aiven_organization.foo.id
  user_email      = "aleks+%s@aiven.io"
}

data "aiven_organization_user" "member" {
  organization_id = aiven_organization_user.foo.organization_id
  user_email      = aiven_organization_user.foo.user_email

  depends_on = [aiven_organization_user.foo]
}`, name, name)
}

func testAccCheckAivenOrganizationUserResourceDestroy(s *terraform.State) error {
	var (
		c, err = acc.GetTestGenAivenClient()
		ctx    = context.Background()
	)

	if err != nil {
		return fmt.Errorf("error getting generated Aiven client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_organization_user" {
			continue
		}

		organizationID, userEmail, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := c.OrganizationGet(ctx, organizationID)
		if err != nil {
			if avngen.IsNotFound(err) {
				return nil
			}

			return err
		}

		if resp.OrganizationId == organizationID {
			respI, err := c.OrganizationUserInvitationsList(ctx, organizationID)
			if err != nil {
				if avngen.IsNotFound(err) {
					return nil
				}

				return err
			}

			for _, i := range respI {
				if i.UserEmail == userEmail {
					return fmt.Errorf("organization user (%s) still exists", rs.Primary.ID)
				}
			}
		}
	}

	return nil
}
