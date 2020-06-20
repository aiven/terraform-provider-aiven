package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccAivenProjectDataSource_basic(t *testing.T) {
	datasourceName := "data.aiven_project.project"
	resourceName := "aiven_project.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "project", resourceName, "project"),
					resource.TestCheckResourceAttrPair(datasourceName, "ca_cert", resourceName, "ca_cert"),
					resource.TestCheckResourceAttrPair(datasourceName, "card_id", resourceName, "card_id"),
				),
			},
		},
	})
}
