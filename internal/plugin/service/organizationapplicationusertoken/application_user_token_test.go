package organizationapplicationusertoken_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
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
			{
				// Minimal-config token: the config omits max_age_seconds and
				// extend_when_used. Both are echoed by Read, and max_age_seconds
				// was not Computed before the generator's auto-promotion rule;
				// import would reject the Flatten'd value without this fix.
				ResourceName:      tokenBar,
				ImportState:       true,
				ImportStateVerify: true,
				// full_token is returned only once at create time.
				ImportStateVerifyIgnore: []string{"full_token"},
				ImportStateIdFunc:       importIDFunc(tokenBar),
			},
		},
	})
}

func TestAccOrganizationApplicationUserTokenResource_removeMaxAgeSeconds(t *testing.T) {
	org := acc.OrganizationName()
	name := acc.RandStr()
	resourceName := "aiven_organization_application_user_token.foo"
	config := func(maxAgeConfig string) string {
		return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = %q
}

resource "aiven_organization_application_user" "foo" {
  organization_id = data.aiven_organization.foo.id
  name            = "test-acc-%s"
}

resource "aiven_organization_application_user_token" "foo" {
  organization_id  = aiven_organization_application_user.foo.organization_id
  user_id          = aiven_organization_application_user.foo.user_id
  extend_when_used = false
  %s
}
`, org, name, maxAgeConfig)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config("max_age_seconds = 3600"),
				Check:  resource.TestCheckResourceAttr(resourceName, "max_age_seconds", "3600"),
			},
			{
				// Removing the lifetime must replace the token, rather than
				// silently retaining the previously configured expiration.
				Config:             config(""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func importIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, err := acc.ResourceFromState(state, resourceName)
		if err != nil {
			return "", err
		}
		return util.ComposeID(
			rs.Primary.Attributes["organization_id"],
			rs.Primary.Attributes["user_id"],
			rs.Primary.Attributes["token_prefix"],
		), nil
	}
}

func validToken(s string) error {
	// Not "", "nil" or some other invalid value
	if len(s) < 100 {
		return fmt.Errorf("expected full_token to have value, got length %d", len(s))
	}
	return nil
}
