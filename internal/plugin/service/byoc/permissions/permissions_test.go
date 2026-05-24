package permissions_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenBYOCPermissions(t *testing.T) {
	acc.SkipIfNotBeta(t)

	envVars := acc.RequireEnvVars(t, "AIVEN_BYOC_AWS_IAM_ROLE_ARN")
	iamRoleARN := envVars["AIVEN_BYOC_AWS_IAM_ROLE_ARN"]
	organizationName := acc.OrganizationName()

	const resourceName = "aiven_byoc_permissions.example"

	baseConfig := fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %q
}

resource "aiven_byoc_aws_entity" "example" {
  organization_id  = data.aiven_organization.org.id
  display_name     = "test-byoc-permissions-acc"
  cloud_provider   = "aws"
  cloud_region     = "aws-eu-west-1"
  deployment_model = "standard"
  reserved_cidr    = "10.0.0.0/16"
  aws_iam_role_arn = %q

  contact_emails {
    email     = "ops@example.com"
    real_name = "Ops Team"
    role      = "admin"
  }
}`, organizationName, iamRoleARN)

	initialConfig := baseConfig + `
resource "aiven_byoc_permissions" "example" {
  organization_id             = data.aiven_organization.org.id
  custom_cloud_environment_id = aiven_byoc_aws_entity.example.custom_cloud_environment_id

  accounts = [data.aiven_organization.org.id]
  projects = []
}`

	updatedConfig := baseConfig + `
resource "aiven_byoc_permissions" "example" {
  organization_id             = data.aiven_organization.org.id
  custom_cloud_environment_id = aiven_byoc_aws_entity.example.custom_cloud_environment_id

  accounts = []
  projects = []
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "data.aiven_organization.org", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "custom_cloud_environment_id", "aiven_byoc_aws_entity.example", "custom_cloud_environment_id"),
					resource.TestCheckResourceAttr(resourceName, "accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "projects.#", "0"),
				),
			},
			{
				Config: initialConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "accounts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "projects.#", "0"),
				),
			},
			{
				Config: updatedConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
