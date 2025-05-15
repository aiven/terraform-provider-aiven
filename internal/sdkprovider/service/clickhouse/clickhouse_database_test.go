package clickhouse_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenClickhouseDatabase_basic(t *testing.T) {
	resourceName := "aiven_clickhouse_database.foo"
	rName := acc.RandStr()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClickhouseDatabaseResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-db-%s", rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccClickhouseDatabaseResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%[1]s"
}

resource "aiven_clickhouse" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "hobbyist"
  service_name            = "test-acc-sr-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_clickhouse_database" "foo" {
  project      = aiven_clickhouse.bar.project
  service_name = aiven_clickhouse.bar.service_name
  name         = "test-acc-db-%[2]s"
}

data "aiven_clickhouse_database" "database" {
  project      = aiven_clickhouse_database.foo.project
  service_name = aiven_clickhouse_database.foo.service_name
  name         = aiven_clickhouse_database.foo.name
}`, acc.ProjectName(), name)
}
