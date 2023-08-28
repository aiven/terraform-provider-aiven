package project_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenProjectUserDataSource_basic(t *testing.T) {
	datasourceName := "data.aiven_project_user.user"
	resourceName := "aiven_project_user.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
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
