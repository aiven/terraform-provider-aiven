package servicelist_test

import (
	"fmt"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAivenServiceListDataSource tests the aiven_service_list data source
// We need to create a new service first to ensure the list is not empty and contains our service.
func TestAccAivenServiceListDataSource(t *testing.T) {
	var (
		projectName    = acc.ProjectName()
		dataSourceName = "data.aiven_service_list.services"
		serviceName    = fmt.Sprintf("test-acc-service-list-%s", acc.RandStr())
	)

	config := fmt.Sprintf(`
resource "aiven_pg" "bar" {
  project      = %q
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = %q
}

data "aiven_service_list" "services" {
  project    = %q
  depends_on = [aiven_pg.bar]
}
`, projectName, serviceName, projectName)

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
					// check that our created service is in the list
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "services.*", map[string]string{
						"service_name": serviceName,
						"service_type": "pg",
						"plan":         "startup-4",
						"cloud_name":   "google-europe-west1",
						"state":        "RUNNING",
					}),
					resource.TestCheckResourceAttrSet(dataSourceName, "services.0.service_uri"),
					resource.TestCheckResourceAttrSet(dataSourceName, "services.0.create_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "services.0.update_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "services.0.node_count"),
					resource.TestCheckResourceAttrSet(dataSourceName, "services.0.node_cpu_count"),
					resource.TestCheckResourceAttrSet(dataSourceName, "services.0.node_memory_mb"),
				),
			},
		},
	})
}
