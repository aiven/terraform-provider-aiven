package plan_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenPlanDataSource(t *testing.T) {
	var (
		projectName    = acc.ProjectName()
		serviceType    = "kafka"
		servicePlan    = "business-4"
		cloudName      = "google-us-east1"
		dataSourceName = "data.aiven_service_plan.business_kafka_plan"
	)

	config := fmt.Sprintf(`
data "aiven_service_plan" "business_kafka_plan" {
  project      = %q
  service_type = %q
  cloud_name   = %q
  service_plan = %q
}
`, projectName, serviceType, cloudName, servicePlan)

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
					resource.TestCheckResourceAttr(dataSourceName, "service_plan", servicePlan),
					resource.TestCheckResourceAttr(dataSourceName, "cloud_name", cloudName),

					resource.TestCheckResourceAttrSet(dataSourceName, "base_price_usd"),
					resource.TestCheckResourceAttrSet(dataSourceName, "object_storage_gb_price_usd"),

					resource.TestCheckResourceAttrSet(dataSourceName, "node_count"),
					resource.TestCheckResourceAttrSet(dataSourceName, "disk_space_mb"),
					resource.TestCheckResourceAttrSet(dataSourceName, "disk_space_cap_mb"),

					resource.TestCheckResourceAttrSet(dataSourceName, "backup_config.#"),
				),
			},
		},
	})
}
