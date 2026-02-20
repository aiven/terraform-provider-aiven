package awsentity_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenByocAwsEntity(t *testing.T) {
	var (
		name             = "aiven_byoc_aws_entity.entity"
		organizationName = acc.OrganizationName()
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %q
}

resource "aiven_byoc_aws_entity" "entity" {
  organization_id  = data.aiven_organization.org.id
  cloud_provider   = "aws"
  cloud_region     = "eu-west-1"
  deployment_model = "standard"
  display_name     = "tf-acc-byoc-entity"
  reserved_cidr    = "10.0.0.0/16"

  contact_emails {
    email = "test@example.com"
  }
}`, organizationName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "id"),
					resource.TestCheckResourceAttrSet(name, "custom_cloud_environment_id"),
					resource.TestCheckResourceAttrSet(name, "state"),
					resource.TestCheckResourceAttrPair(name, "organization_id", "data.aiven_organization.org", "id"),
					resource.TestCheckResourceAttr(name, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(name, "cloud_region", "eu-west-1"),
					resource.TestCheckResourceAttr(name, "deployment_model", "standard"),
					resource.TestCheckResourceAttr(name, "display_name", "tf-acc-byoc-entity"),
					resource.TestCheckResourceAttr(name, "reserved_cidr", "10.0.0.0/16"),
				),
			},
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %q
}

resource "aiven_byoc_aws_entity" "entity" {
  organization_id  = data.aiven_organization.org.id
  cloud_provider   = "aws"
  cloud_region     = "eu-west-1"
  deployment_model = "standard"
  display_name     = "tf-acc-byoc-entity-updated"
  reserved_cidr    = "10.0.0.0/16"

  contact_emails {
    email = "updated@example.com"
  }

  tags = {
    environment = "test"
  }
}`, organizationName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "id"),
					resource.TestCheckResourceAttr(name, "display_name", "tf-acc-byoc-entity-updated"),
					resource.TestCheckResourceAttr(name, "tags.environment", "test"),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
