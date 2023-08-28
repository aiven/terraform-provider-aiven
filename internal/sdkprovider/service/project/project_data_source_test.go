package project_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenProjectDataSource_basic(t *testing.T) {
	datasourceName := "data.aiven_project.project"
	resourceName := "aiven_project.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "project", resourceName, "project"),
					resource.TestCheckResourceAttrPair(datasourceName, "ca_cert", resourceName, "ca_cert"),
				),
			},
		},
	})
}
