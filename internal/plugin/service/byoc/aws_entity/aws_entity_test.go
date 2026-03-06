package awsentity_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenBYOCAWSEntity(t *testing.T) {
	acc.SkipIfNotBeta(t)

	envVars := acc.RequireEnvVars(t, "AIVEN_BYOC_AWS_IAM_ROLE_ARN")
	iamRoleARN := envVars["AIVEN_BYOC_AWS_IAM_ROLE_ARN"]
	organizationName := acc.OrganizationName()

	const resourceName = "aiven_byoc_aws_entity.example"

	baseConfig := fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %q
}`, organizationName)

	initialConfig := baseConfig + fmt.Sprintf(`
resource "aiven_byoc_aws_entity" "example" {
  organization_id  = data.aiven_organization.org.id
  display_name     = "test-byoc-acc"
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
}`, iamRoleARN)

	updatedConfig := baseConfig + fmt.Sprintf(`
resource "aiven_byoc_aws_entity" "example" {
  organization_id  = data.aiven_organization.org.id
  display_name     = "test-byoc-acc-updated"
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

  contact_emails {
    email = "devops@example.com"
    role  = "ops"
  }
}`, iamRoleARN)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_cloud_environment_id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "data.aiven_organization.org", "id"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "test-byoc-acc"),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "cloud_region", "aws-eu-west-1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_model", "standard"),
					resource.TestCheckResourceAttr(resourceName, "reserved_cidr", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "aws_iam_role_arn", iamRoleARN),
					resource.TestCheckResourceAttr(resourceName, "contact_emails.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "contact_emails.*", map[string]string{
						"email":     "ops@example.com",
						"real_name": "Ops Team",
						"role":      "admin",
					}),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "test-byoc-acc-updated"),
					resource.TestCheckResourceAttr(resourceName, "contact_emails.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "contact_emails.*", map[string]string{
						"email": "devops@example.com",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
