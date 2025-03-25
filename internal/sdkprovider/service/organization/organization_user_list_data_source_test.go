package organization_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/organization"
)

func testAccAivenOrganizationUserListByName(name string) string {
	return fmt.Sprintf(`
data "aiven_organization_user_list" "org" {
  name = "%s"
}
`, name)
}

func TestAccAivenOrganizationUserListByName(t *testing.T) {
	resourceName := "data.aiven_organization_user_list.org"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAivenOrganizationUserListByName(acc.OrganizationName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "users.#"),
					resource.TestMatchResourceAttr(resourceName, "users.0.user_info.0.user_email", regexp.MustCompile(`.*@.*`)),
				),
			},
		},
	})
}

func testAccAivenOrganizationUserListByID(id string) string {
	return fmt.Sprintf(`
data "aiven_organization_user_list" "org" {
  id = "%s"
}
`, id)
}

func TestAccAivenOrganizationUserListByID(t *testing.T) {
	// Skip test if TF_ACC is not set
	acc.TestAccPreCheck(t)

	// This test creates Aiven client before running PreCheck part
	// Runs checks manually
	resourceName := "data.aiven_organization_user_list.org"
	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	id, err := organization.GetOrganizationByName(
		context.Background(),
		client,
		acc.OrganizationName(),
	)
	require.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAivenOrganizationUserListByID(id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "users.#"),
					resource.TestMatchResourceAttr(resourceName, "users.0.user_info.0.user_email", regexp.MustCompile(`.*@.*`)),
				),
			},
		},
	})
}
