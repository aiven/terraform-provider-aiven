package byocazureprovision_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenBYOCAzureProvision(t *testing.T) {
	acc.SkipBYOC(t)
	acc.SkipIfNotBeta(t)

	envVars := acc.RequireEnvVars(t, "AIVEN_BYOC_AZURE_SUBSCRIPTION_ID", "AIVEN_BYOC_AZURE_TENANT_ID")
	subscriptionID := envVars["AIVEN_BYOC_AZURE_SUBSCRIPTION_ID"]
	tenantID := envVars["AIVEN_BYOC_AZURE_TENANT_ID"]
	organizationName := acc.OrganizationName()

	const resourceName = "aiven_byoc_azure_provision.example"

	config := fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %q
}

resource "aiven_byoc_azure_entity" "example" {
  organization_id  = data.aiven_organization.org.id
  display_name     = "test-byoc-provision-acc"
  cloud_region     = "westeurope"
  deployment_model = "standard"
  reserved_cidr    = "10.0.0.0/16"

  contact_emails {
    email = "ops@example.com"
    role  = "admin"
  }
}

resource "aiven_byoc_azure_provision" "example" {
  organization_id             = data.aiven_organization.org.id
  custom_cloud_environment_id = aiven_byoc_azure_entity.example.custom_cloud_environment_id
  azure_subscription_id       = %q
  azure_tenant_id             = %q
}`, organizationName, subscriptionID, tenantID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "aiven_azure_cce_client_id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "data.aiven_organization.org", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "custom_cloud_environment_id", "aiven_byoc_azure_entity.example", "custom_cloud_environment_id"),
					resource.TestCheckResourceAttr(resourceName, "azure_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "azure_tenant_id", tenantID),
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
