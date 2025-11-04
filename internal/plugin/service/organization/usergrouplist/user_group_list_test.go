package usergrouplist_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenOrganizationUserGroupListDataSource(t *testing.T) {
	const dataSourceName = "data.aiven_organization_user_group_list.ds"
	groupName := "test-group-" + acc.RandStr()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %[1]q
}

resource "aiven_organization_user_group" "group" {
  organization_id = data.aiven_organization.org.id
  name            = %[2]q
  description     = "test group description"
}

data "aiven_organization_user_group_list" "ds" {
  organization_id = data.aiven_organization.org.id
  depends_on      = [aiven_organization_user_group.group]
}
`, acc.OrganizationName(), groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "user_groups.#"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "user_groups.*", map[string]string{
						"user_group_name": groupName,
						"description":     "test group description",
					}),
				),
			},
		},
	})
}
