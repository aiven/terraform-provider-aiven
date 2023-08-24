package organization_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenOrganizationUser_basic(t *testing.T) {
	resourceName := "aiven_organization_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenOrganizationUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationUserResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenOrganizationUserAttributes("data.aiven_organization_user.member"),
					resource.TestCheckResourceAttr(
						resourceName, "user_email", fmt.Sprintf("aleks+%s@aiven.io", rName),
					),
					resource.TestCheckResourceAttr(resourceName, "accepted", "false"),
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
	c := acc.GetTestAivenClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_organization_user" {
			continue
		}

		organizationID, userEmail, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		r, err := c.Organization.Get(organizationID)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		if r.ID == organizationID {
			ri, err := c.OrganizationUserInvitations.List(organizationID)
			if err != nil {
				if err.(aiven.Error).Status != 404 {
					return err
				}

				return nil
			}

			for _, i := range ri.Invitations {
				if i.UserEmail == userEmail {
					return fmt.Errorf("organization user (%s) still exists", rs.Primary.ID)
				}
			}
		}
	}

	return nil
}

func testAccCheckAivenOrganizationUserAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] organization user attributes %v", a)

		if a["organization_id"] == "" {
			return fmt.Errorf("expected to get an organization_id from Aiven")
		}

		if a["user_email"] == "" {
			return fmt.Errorf("expected to get a user_email from Aiven")
		}

		if a["create_time"] == "" {
			return fmt.Errorf("expected to get a create_time from Aiven")
		}

		if a["accepted"] != "false" {
			return fmt.Errorf("expected to get a accepted from Aiven")
		}

		if a["invited_by"] == "" {
			return fmt.Errorf("expected to get a invited_by from Aiven")
		}

		return nil
	}
}
