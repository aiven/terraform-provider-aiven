package externalidentity_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestExternalIdentityDataSource tests the external_identity datasource.
func TestExternalIdentityDataSource(t *testing.T) {
	acc.SkipIfNotBeta(t)

	organizationName := acc.OrganizationName()
	prefix := acc.DefaultResourceNamePrefix
	suffix := acctest.RandStringFromCharSet(acc.DefaultRandomSuffixLength, acctest.CharSetAlphaNum)
	userGroupName := fmt.Sprintf("%s-usr-group-%s", prefix, suffix)
	resourceName := "data.aiven_external_identity.foo"
	externalUserID := "alice"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testExternalIdentityDataSourceBasic(organizationName, userGroupName, externalUserID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "external_user_id", externalUserID),
				),
			},
		},
	})
}

func testExternalIdentityDataSourceBasic(organizationName string, userGroupName string, externalUserID string) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%s"
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "foo"
}

resource "aiven_organization_user_group" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "%s"
  description     = "Terraform acceptance tests"
}

data "aiven_external_identity" "foo" {
  organization_id       = data.aiven_organization.foo.id
  internal_user_id      = aiven_organization_application_user.foo.user_id
  external_user_id      = "%s"
  external_service_name = "github"
}
`, organizationName, userGroupName, externalUserID)
}
