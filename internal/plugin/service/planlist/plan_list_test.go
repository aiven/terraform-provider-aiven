package planlist_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenPlanListDataSource(t *testing.T) {
	var (
		projectName    = acc.ProjectName()
		serviceType    = "kafka"
		dataSourceName = "data.aiven_service_plan_list.plans"
	)

	config := fmt.Sprintf(`
data "aiven_service_plan_list" "plans" {
  project      = %q
  service_type = %q
}
`, projectName, serviceType)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "project", projectName),
					resource.TestCheckResourceAttr(dataSourceName, "service_type", serviceType),

					resource.TestCheckResourceAttrSet(dataSourceName, "service_plans.#"),

					resource.TestCheckResourceAttrSet(dataSourceName, "service_plans.0.service_plan"),

					resource.TestCheckResourceAttrSet(dataSourceName, "service_plans.0.cloud_names.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "service_plans.0.cloud_names.0"),
				),
			},
		},
	})
}
