package applicationusertoken_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccOrganizationApplicationUserTokenResource(t *testing.T) {
	org := acc.OrganizationName()

	tokenFoo := "aiven_organization_application_user_token.foo"
	tokenBar := "aiven_organization_application_user_token.bar"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%[2]s"
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "test-acc-%[1]s"
}

resource "aiven_organization_application_user_token" "foo" {
  organization_id  = aiven_organization_application_user.foo.organization_id
  user_id          = aiven_organization_application_user.foo.user_id
  description      = "Terraform acceptance tests"
  max_age_seconds  = 3600
  extend_when_used = true
  scopes           = ["user:read"]
  ip_allowlist     = ["10.0.0.0/8"]
}

// Required fields only
resource "aiven_organization_application_user_token" "bar" {
  organization_id = aiven_organization_application_user.foo.organization_id
  user_id         = aiven_organization_application_user.foo.user_id
}
`, acc.RandStr(), org),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tokenFoo, "description", "Terraform acceptance tests"),
					resource.TestCheckResourceAttr(tokenFoo, "max_age_seconds", "3600"),
					resource.TestCheckResourceAttr(tokenFoo, "extend_when_used", "true"),
					resource.TestCheckResourceAttr(tokenFoo, "scopes.#", "1"),
					resource.TestCheckResourceAttr(tokenFoo, "ip_allowlist.#", "1"),
					resource.TestCheckResourceAttr(tokenFoo, "ip_allowlist.0", "10.0.0.0/8"),
					// Bar token has required fields only
					resource.TestCheckResourceAttr(tokenBar, "extend_when_used", "false"),
				),
			},
			{
				// The token must survive the refresh command
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(tokenFoo, "full_token", validToken),
					resource.TestCheckResourceAttrWith(tokenBar, "full_token", validToken),
				),
			},
		},
	})
}

func validToken(s string) error {
	// Not "", "nil" or some other invalid value
	if len(s) < 100 {
		return fmt.Errorf("expected full_token to have value, got length %d", len(s))
	}
	return nil
}
