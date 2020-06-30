package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccAivenProjectUserDataSource_basic(t *testing.T) {
	datasourceName := "data.aiven_project_user.user"
	resourceName := "aiven_project_user.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectUserResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "project", resourceName, "project"),
					resource.TestCheckResourceAttrPair(datasourceName, "email", resourceName, "email"),
					resource.TestCheckResourceAttrPair(datasourceName, "member_type", resourceName, "member_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "accepted", resourceName, "accepted"),
				),
			},
		},
	})
}
