package org_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

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

// TestAccAivenOrganizationImport tests importing an organization using both ID (a123345, org123456) formats
func TestAccAivenOrganizationImport(t *testing.T) {
	var (
		resourceName = "aiven_organization.test"
		suffix       = acctest.RandStringFromCharSet(acc.DefaultRandomSuffixLength, acctest.CharSetAlphaNum)
		orgName      = acc.DefaultResourceNamePrefix + "-org-" + suffix
		templStore   = template.InitializeTemplateStore(t)
		orgID        string
		accountID    string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// First create the organization
			{
				Config: templStore.NewBuilder().
					AddResource("aiven_organization", map[string]interface{}{
						"resource_name": "test",
						"name":          orgName,
					}).MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					// Basic attribute checks
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", orgName),

					// Capture ID for import tests and fetch the account ID using the API
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}

						// Get the organization ID (org123456 format)
						orgID = rs.Primary.Attributes["id"]

						client, err := acc.GetTestGenAivenClient()
						require.NoError(t, err)
						// Use the client to fetch the account ID from the API
						org, err := client.OrganizationGet(context.Background(), orgID)
						if err != nil {
							return err
						}

						// Use the actual account ID from the API
						accountID = org.AccountId
						return nil
					},
				),
			},
			// Test 1: Import with organization ID format (org123456)
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false, // Don't verify automatically
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return orgID, nil
				},
				// Custom check for the imported state
				Check: resource.ComposeTestCheckFunc(
					// Verify that the ID is set to the organization ID
					resource.TestCheckResourceAttr(resourceName, "id", orgID),
					// Verify the name is still correct
					resource.TestCheckResourceAttr(resourceName, "name", orgName),
					// Verify other fields are set
					resource.TestCheckResourceAttrSet(resourceName, "tenant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "update_time"),
				),
			},
			// Test 2: Import with account ID format (a123456)
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false, // Don't verify automatically
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return accountID, nil
				},
				// Custom check for the imported state
				Check: resource.ComposeTestCheckFunc(
					// Verify that the ID is set to the organization ID (not the account ID)
					resource.TestCheckResourceAttr(resourceName, "id", orgID),
					// Verify the name is still correct
					resource.TestCheckResourceAttr(resourceName, "name", orgName),
					// Verify other fields are set
					resource.TestCheckResourceAttrSet(resourceName, "tenant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "update_time"),
				),
			},
		},
	})
}
