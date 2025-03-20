package org_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
)

func TestAccAivenOrganizationResource(t *testing.T) {
	var (
		resourceName = "aiven_organization.test"
		suffix       = acctest.RandStringFromCharSet(acc.DefaultRandomSuffixLength, acctest.CharSetAlphaNum)
		orgName      = acc.DefaultResourceNamePrefix + "-org-" + suffix
		updatedName  = acc.DefaultResourceNamePrefix + "-org-" + suffix + "-1"
		templStore   = template.InitializeTemplateStore(t)
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: templStore.NewBuilder().
					AddResource("aiven_organization", map[string]interface{}{
						"resource_name": "test",
						"name":          orgName,
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					// Check computed fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "tenant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "update_time"),

					// Check values match what we set
					resource.TestCheckResourceAttr(resourceName, "name", orgName),
				),
			},
			{
				Config: templStore.NewBuilder().
					AddResource("aiven_organization", map[string]interface{}{
						"resource_name": "test",
						"name":          updatedName,
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					// Check name was updated
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}
