package awsprovision_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenBYOCAwsProvision(t *testing.T) {
	acc.SkipIfNotBeta(t)

	envVars := acc.RequireEnvVars(t, "AIVEN_BYOC_AWS_IAM_ROLE_ARN")
	iamRoleARN := envVars["AIVEN_BYOC_AWS_IAM_ROLE_ARN"]
	organizationName := acc.OrganizationName()

	const resourceName = "aiven_byoc_aws_provision.example"

	config := fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %q
}

resource "aiven_byoc_aws_entity" "example" {
  organization_id  = data.aiven_organization.org.id
  display_name     = "test-byoc-provision-acc"
  cloud_provider   = "aws"
  cloud_region     = "aws-eu-west-1"
  deployment_model = "standard"
  reserved_cidr    = "10.0.0.0/16"
  aws_iam_role_arn = %[2]q

  contact_emails {
    email = "ops@example.com"
    role  = "admin"
  }
}

resource "aiven_byoc_aws_provision" "example" {
  organization_id             = data.aiven_organization.org.id
  custom_cloud_environment_id = aiven_byoc_aws_entity.example.custom_cloud_environment_id
  aws_iam_role_arn            = %[2]q
}`, organizationName, iamRoleARN)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "aiven_aws_account_principal"),
					resource.TestCheckResourceAttrSet(resourceName, "aiven_aws_assume_role_external_id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "data.aiven_organization.org", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "custom_cloud_environment_id", "aiven_byoc_aws_entity.example", "custom_cloud_environment_id"),
					resource.TestCheckResourceAttr(resourceName, "aws_iam_role_arn", iamRoleARN),
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
