package organization_test

import (
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAivenOrganizationUserDataSource_basic(t *testing.T) {
	datasourceName := "data.aiven_organization_user.member"
	resourceName := "aiven_organization_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationUserResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						datasourceName, "organization_id", resourceName, "organization_id",
					),
					resource.TestCheckResourceAttrPair(datasourceName, "user_email", resourceName, "user_email"),
					resource.TestCheckResourceAttrPair(datasourceName, "create_time", resourceName, "create_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "accepted", resourceName, "accepted"),
				),
			},
		},
	})
}