package org_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
)

func TestAccAivenOrganizationDataSource(t *testing.T) {
	var (
		resourceName = "aiven_organization.test"
		dsNameByID   = "data.aiven_organization.by_id"
		dsNameByName = "data.aiven_organization.by_name"
		suffix       = acctest.RandStringFromCharSet(acc.DefaultRandomSuffixLength, acctest.CharSetAlphaNum)
		orgName      = acc.DefaultResourceNamePrefix + "-org-" + suffix
		templBuilder = template.InitializeTemplateStore(t).NewBuilder()
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: templBuilder.
					AddResource("aiven_organization", map[string]interface{}{
						"resource_name": "test",
						"name":          orgName,
					}).
					AddDataSource("aiven_organization", map[string]interface{}{
						"resource_name": "by_id",
						"id":            template.Reference("aiven_organization.test.id"),
					}).
					AddDataSource("aiven_organization", map[string]interface{}{
						"resource_name": "by_name",
						"name":          template.Reference("aiven_organization.test.name"),
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					// Check computed fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "tenant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "update_time"),

					// Check data source by ID matches resource
					resource.TestCheckResourceAttrPair(dsNameByID, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dsNameByID, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dsNameByID, "tenant_id", resourceName, "tenant_id"),
					resource.TestCheckResourceAttrPair(dsNameByID, "create_time", resourceName, "create_time"),
					resource.TestCheckResourceAttrPair(dsNameByID, "update_time", resourceName, "update_time"),

					// Check data source by name matches resource
					resource.TestCheckResourceAttrPair(dsNameByName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dsNameByName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dsNameByName, "tenant_id", resourceName, "tenant_id"),
					resource.TestCheckResourceAttrPair(dsNameByName, "create_time", resourceName, "create_time"),
					resource.TestCheckResourceAttrPair(dsNameByName, "update_time", resourceName, "update_time"),

					// Directly check some values for additional confirmation
					resource.TestCheckResourceAttr(resourceName, "name", orgName),
				),
			},
		},
	})
}

func TestAccAivenOrganizationDataSourceValidation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test case: neither id nor name provided
				Config: `data "aiven_organization" "test" {}`,
				ExpectError: regexp.MustCompile(
					`At least one of these attributes must be configured: \[id,name\]`,
				),
			},
			{
				// Test case: both id and name provided
				Config: `
data "aiven_organization" "test" {
  id   = "test-id"
  name = "test-name"
}
				`,
				ExpectError: regexp.MustCompile(
					`These attributes cannot be configured together: \[id,name\]`,
				),
			},
			{
				// Test case: empty id provided
				Config: `
data "aiven_organization" "test" {
  id = ""
}
				`,
				ExpectError: regexp.MustCompile(
					`Organization ID not found`,
				),
			},
		},
	})
}
