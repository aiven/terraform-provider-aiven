package servicelist_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenServiceListDataSource(t *testing.T) {
	var (
		projectName    = acc.ProjectName()
		dataSourceName = "data.aiven_service_list.services"
	)

	config := fmt.Sprintf(`
data "aiven_service_list" "services" {
  project = %q
}
`, projectName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "project", projectName),
					resource.TestCheckResourceAttrSet(dataSourceName, "services.#"),
				),
			},
		},
	})
}
