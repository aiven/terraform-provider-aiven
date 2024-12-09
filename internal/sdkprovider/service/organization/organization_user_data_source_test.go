package organization_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenOrganizationUserDataSource_using_email(t *testing.T) {
	var (
		orgID          = os.Getenv("AIVEN_ORG_ID")
		email          = os.Getenv("AIVEN_ORG_USER_EMAIL")
		datasourceName = "data.aiven_organization_user.member"
	)

	if orgID == "" || email == "" {
		t.Skip("Skipping test due to missing AIVEN_ORG_ID or AIVEN_ORG_USER_EMAIL environment variable")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationUserDataResourceByEmail(orgID, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "user_email"),
					resource.TestCheckResourceAttrSet(datasourceName, "create_time"),
					resource.TestCheckResourceAttrSet(datasourceName, "user_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "organization_id"),
				),
			},
		},
	})
}

func TestAccAivenOrganizationUserDataSource_using_userid(t *testing.T) {
	var (
		orgID          = os.Getenv("AIVEN_ORG_ID")
		userID         = os.Getenv("AIVEN_ORG_USER_ID")
		datasourceName = "data.aiven_organization_user.member"
	)

	if orgID == "" || userID == "" {
		t.Skip("Skipping test due to missing AIVEN_ORG_ID or AIVEN_ORG_USER_ID environment variable")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationUserDataResourceByUserID(orgID, userID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "user_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "create_time"),
					resource.TestCheckResourceAttrSet(datasourceName, "user_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "organization_id"),
				),
			},
		},
	})
}

func testAccOrganizationUserDataResourceByUserID(orgID, userID string) string {
	return fmt.Sprintf(`
data "aiven_organization_user" "member" {
  organization_id = "%s"
  user_id         = "%s"
}`, orgID, userID)
}

func testAccOrganizationUserDataResourceByEmail(orgID, email string) string {
	return fmt.Sprintf(`
data "aiven_organization_user" "member" {
  organization_id = "%s"
  user_email      = "%s"
}`, orgID, email)
}
