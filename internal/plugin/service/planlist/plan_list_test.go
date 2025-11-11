package planlist_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

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
					resource.TestCheckResourceAttrSet(dataSourceName, "service_plans.0.regions.aws-eu-west-1.price_usd"),
					func(state *terraform.State) error {
						client, err := acc.GetTestGenAivenClient()
						require.NoError(t, err)

						rsp, err := client.ProjectServicePlanList(t.Context(), projectName, serviceType)
						require.NoError(t, err)
						require.NotEmpty(t, rsp)

						for i, plan := range rsp {
							for k, v := range plan.Regions {
								key := fmt.Sprintf("service_plans.%d.regions.%s.price_usd", i, k)
								vMap, ok := v.(map[string]any)
								require.True(t, ok, "unexpected type for region data: %T", v)

								err = resource.TestCheckResourceAttr(dataSourceName, key, fmt.Sprint(vMap["price_usd"]))(state)
								require.NoError(t, err)
							}
						}
						return nil
					},
				),
			},
		},
	})
}
